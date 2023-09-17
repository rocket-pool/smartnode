package minipool

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/beacon"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	rpbeacon "github.com/rocket-pool/smartnode/rocketpool/common/beacon"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/validator"
)

// ===============
// === Factory ===
// ===============

type minipoolRescueDissolvedContextFactory struct {
	handler *MinipoolHandler
}

func (f *minipoolRescueDissolvedContextFactory) Create(vars map[string]string) (*minipoolRescueDissolvedContext, error) {
	c := &minipoolRescueDissolvedContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("addresses", vars, cliutils.ValidateAddresses, &c.minipoolAddresses),
		server.ValidateArg("depositAmounts", vars, cliutils.ValidateBigInts, &c.depositAmounts),
	}
	return c, errors.Join(inputErrs...)
}

// ===============
// === Context ===
// ===============

type minipoolRescueDissolvedContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
	depositAmounts    []*big.Int
}

func (c *minipoolRescueDissolvedContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	// Sanity check
	if len(c.minipoolAddresses) != len(c.depositAmounts) {
		return fmt.Errorf("addresses and deposit amounts must have the same length (%d vs. %d)", len(c.minipoolAddresses), len(c.depositAmounts))
	}

	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	w := sp.GetWallet()
	bc := sp.GetBeaconClient()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
		sp.RequireBeaconClientSynced(),
	)
	if err != nil {
		return err
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(c.minipoolAddresses))
	for i, address := range c.minipoolAddresses {
		amount := c.depositAmounts[i]
		opts.Value = amount
		txInfo, err := getDepositTx(rp, w, bc, address, amount, opts)
		if err != nil {
			return fmt.Errorf("error simulating deposit transaction for minipool %s: %w", address.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	data.TxInfos = txInfos
	return nil
}

// Create a transaction for submitting a rescue deposit, optionally simulating it only for gas estimation
func getDepositTx(rp *rocketpool.RocketPool, w *wallet.LocalWallet, bc rpbeacon.Client, minipoolAddress common.Address, amount *big.Int, opts *bind.TransactOpts) (*core.TransactionInfo, error) {
	beaconDeposit, err := beacon.NewBeaconDeposit(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating Beacon deposit contract binding: %w", err)
	}

	// Create minipool
	mp, err := minipool.CreateMinipoolFromAddress(rp, minipoolAddress, false, nil)
	if err != nil {
		return nil, err
	}
	mpCommon := mp.GetMinipoolCommon()

	// Get eth2 config
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return nil, err
	}

	// Get the contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		mpCommon.GetWithdrawalCredentials(mc)
		mpCommon.GetPubkey(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Get minipool withdrawal credentials and keys
	withdrawalCredentials := mpCommon.Details.WithdrawalCredentials
	validatorPubkey := mpCommon.Details.Pubkey
	validatorKey, err := w.GetValidatorKeyByPubkey(validatorPubkey)
	if err != nil {
		return nil, fmt.Errorf("error getting validator private key for pubkey %s: %w", validatorPubkey.Hex(), err)
	}

	// Get validator deposit data
	amountGwei := big.NewInt(0).Div(amount, big.NewInt(1e9)).Uint64()
	depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, amountGwei)
	if err != nil {
		return nil, err
	}
	signature := types.BytesToValidatorSignature(depositData.Signature)

	// Get the tx info
	txInfo, err := beaconDeposit.Deposit(opts, validatorPubkey, withdrawalCredentials, signature, depositDataRoot)
	if err != nil {
		return nil, fmt.Errorf("error performing rescue deposit: %s", err.Error())
	}
	return txInfo, nil
}
