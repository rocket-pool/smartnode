package minipool

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	nmc_validator "github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
)

// ===============
// === Factory ===
// ===============

type minipoolStakeContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolStakeContextFactory) Create(args url.Values) (*minipoolStakeContext, error) {
	c := &minipoolStakeContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("addresses", args, minipoolAddressBatchSize, input.ValidateAddress, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolStakeContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolStakeContext, types.BatchTxInfoData](
		router, "stake", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolStakeContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolStakeContext) PrepareData(data *types.BatchTxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	vMgr := sp.GetValidatorManager()
	rs := sp.GetNetworkResources()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}
	err = sp.RequireWalletReady()
	if err != nil {
		return types.ResponseStatus_WalletNotReady, err
	}

	// Create minipools
	mpMgr, err := minipool.NewMinipoolManager(rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool manager binding: %w", err)
	}
	mps, err := mpMgr.CreateMinipoolsFromAddresses(c.minipoolAddresses, false, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool bindings: %w", err)
	}

	// Get the relevant details
	err = rp.BatchQuery(len(c.minipoolAddresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpCommon := mps[i].Common()
		eth.AddQueryablesToMulticall(mc,
			mpCommon.WithdrawalCredentials,
			mpCommon.Pubkey,
			mpCommon.DepositType,
		)
		return nil
	}, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting minipool details: %w", err)
	}

	// Get the TXs
	txInfos := make([]*eth.TransactionInfo, len(c.minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.Common()
		pubkey := mpCommon.Pubkey.Get()

		withdrawalCredentials := mpCommon.WithdrawalCredentials.Get()
		validatorKey, err := vMgr.LoadValidatorKey(pubkey)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting validator %s (minipool %s) key: %w", pubkey.Hex(), mpCommon.Address.Hex(), err)
		}
		depositType := mpCommon.DepositType.Formatted()

		var depositAmount uint64
		switch depositType {
		case rptypes.Full, rptypes.Half, rptypes.Empty:
			depositAmount = uint64(16e9) // 16 ETH in gwei
		case rptypes.Variable:
			depositAmount = uint64(31e9) // 31 ETH in gwei
		default:
			return types.ResponseStatus_Error, fmt.Errorf("error staking minipool %s: unknown deposit type %d", mpCommon.Address.Hex(), depositType)
		}

		// Get validator deposit data
		depositData, err := nmc_validator.GetDepositData(validatorKey, withdrawalCredentials, rs.GenesisForkVersion, depositAmount, rs.EthNetworkName)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting deposit data for validator %s: %w", pubkey.Hex(), err)
		}
		signature := beacon.ValidatorSignature(depositData.Signature)

		depositDataRoot := common.BytesToHash(depositData.DepositDataRoot)
		txInfo, err := mpCommon.Stake(signature, depositDataRoot, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error simulating stake transaction for minipool %s: %w", mpCommon.Address.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	data.TxInfos = txInfos
	return types.ResponseStatus_Success, nil
}
