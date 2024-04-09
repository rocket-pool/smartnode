package minipool

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	beacon "github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/node-manager-core/eth"
	nmc_validator "github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/utils/input"
	rpbeacon "github.com/rocket-pool/rocketpool-go/v2/beacon"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
)

// ===============
// === Factory ===
// ===============

type minipoolRescueDissolvedContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolRescueDissolvedContextFactory) Create(args url.Values) (*minipoolRescueDissolvedContext, error) {
	c := &minipoolRescueDissolvedContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
		server.ValidateArgBatch("deposit-amounts", args, minipoolAddressBatchSize, input.ValidateBigInt, &c.depositAmounts),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolRescueDissolvedContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolRescueDissolvedContext, types.BatchTxInfoData](
		router, "rescue-dissolved", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolRescueDissolvedContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
	depositAmounts    []*big.Int
	rp                *rocketpool.RocketPool
	w                 *wallet.Wallet
	vMgr              *validator.ValidatorManager
	bc                beacon.IBeaconClient
	rs                *config.NetworkResources

	mpMgr *minipool.MinipoolManager
}

func (c *minipoolRescueDissolvedContext) PrepareData(data *types.BatchTxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Sanity check
	if len(c.minipoolAddresses) != len(c.depositAmounts) {
		return types.ResponseStatus_InvalidArguments, fmt.Errorf("addresses and deposit amounts must have the same length (%d vs. %d)", len(c.minipoolAddresses), len(c.depositAmounts))
	}

	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.vMgr = sp.GetValidatorManager()
	c.w = sp.GetWallet()
	c.bc = sp.GetBeaconClient()
	c.rs = sp.GetNetworkResources()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}
	err = sp.RequireBeaconClientSynced(c.handler.ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}

	// Bindings
	c.mpMgr, err = minipool.NewMinipoolManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool manager binding: %w", err)
	}

	// Get the TXs
	txInfos := make([]*eth.TransactionInfo, len(c.minipoolAddresses))
	for i, address := range c.minipoolAddresses {
		amount := c.depositAmounts[i]
		opts.Value = amount
		txInfo, err := c.getDepositTx(address, amount, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error simulating deposit transaction for minipool %s: %w", address.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	data.TxInfos = txInfos
	return types.ResponseStatus_Success, nil
}

// Create a transaction for submitting a rescue deposit, optionally simulating it only for gas estimation
func (c *minipoolRescueDissolvedContext) getDepositTx(minipoolAddress common.Address, amount *big.Int, opts *bind.TransactOpts) (*eth.TransactionInfo, error) {
	beaconDeposit, err := rpbeacon.NewBeaconDeposit(c.rp)
	if err != nil {
		return nil, fmt.Errorf("error creating Beacon deposit contract binding: %w", err)
	}

	// Create minipool
	mp, err := c.mpMgr.CreateMinipoolFromAddress(minipoolAddress, false, nil)
	if err != nil {
		return nil, err
	}
	mpCommon := mp.Common()

	// Get the contract state
	err = c.rp.Query(nil, nil, mpCommon.WithdrawalCredentials, mpCommon.Pubkey)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Get minipool withdrawal credentials and keys
	withdrawalCredentials := mpCommon.WithdrawalCredentials.Get()
	validatorPubkey := mpCommon.Pubkey.Get()
	validatorKey, err := c.vMgr.LoadValidatorKey(validatorPubkey)
	if err != nil {
		return nil, fmt.Errorf("error getting validator private key for pubkey %s: %w", validatorPubkey.Hex(), err)
	}

	// Get validator deposit data
	amountGwei := big.NewInt(0).Div(amount, big.NewInt(1e9)).Uint64()
	depositData, err := nmc_validator.GetDepositData(validatorKey, withdrawalCredentials, c.rs.GenesisForkVersion, amountGwei, c.rs.EthNetworkName)
	if err != nil {
		return nil, err
	}
	signature := beacon.ValidatorSignature(depositData.Signature)

	// Get the tx info
	depositDataRoot := common.BytesToHash(depositData.DepositDataRoot)
	txInfo, err := beaconDeposit.Deposit(opts, validatorPubkey, withdrawalCredentials, signature, depositDataRoot)
	if err != nil {
		return nil, fmt.Errorf("error performing rescue deposit: %s", err.Error())
	}
	return txInfo, nil
}
