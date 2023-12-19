package minipool

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/validator"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		server.ValidateArg("addresses", args, input.ValidateAddresses, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *minipoolStakeContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*minipoolStakeContext, api.BatchTxInfoData](
		router, "stake", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type minipoolStakeContext struct {
	handler           *MinipoolHandler
	minipoolAddresses []common.Address
}

func (c *minipoolStakeContext) PrepareData(data *api.BatchTxInfoData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	w := sp.GetWallet()
	bc := sp.GetBeaconClient()

	// Requirements
	err := errors.Join(
		sp.RequireNodeRegistered(),
		sp.RequireWalletReady(),
	)
	if err != nil {
		return err
	}

	// Get eth2 config
	eth2Config, err := bc.GetEth2Config()
	if err != nil {
		return fmt.Errorf("error getting Beacon config: %w", err)
	}

	// Create minipools
	mpMgr, err := minipool.NewMinipoolManager(rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager binding: %w", err)
	}
	mps, err := mpMgr.CreateMinipoolsFromAddresses(c.minipoolAddresses, false, nil)
	if err != nil {
		return fmt.Errorf("error creating minipool bindings: %w", err)
	}

	// Get the relevant details
	err = rp.BatchQuery(len(c.minipoolAddresses), minipoolBatchSize, func(mc *batch.MultiCaller, i int) error {
		mpCommon := mps[i].Common()
		core.AddQueryablesToMulticall(mc,
			mpCommon.WithdrawalCredentials,
			mpCommon.Pubkey,
			mpCommon.DepositType,
		)
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting minipool details: %w", err)
	}

	// Get the TXs
	txInfos := make([]*core.TransactionInfo, len(c.minipoolAddresses))
	for i, mp := range mps {
		mpCommon := mp.Common()
		pubkey := mpCommon.Pubkey.Get()

		withdrawalCredentials := mpCommon.WithdrawalCredentials.Get()
		validatorKey, err := w.GetValidatorKeyByPubkey(pubkey)
		if err != nil {
			return fmt.Errorf("error getting validator %s (minipool %s) key: %w", pubkey.Hex(), mpCommon.Address.Hex(), err)
		}
		depositType := mpCommon.DepositType.Formatted()

		var depositAmount uint64
		switch depositType {
		case types.Full, types.Half, types.Empty:
			depositAmount = uint64(16e9) // 16 ETH in gwei
		case types.Variable:
			depositAmount = uint64(31e9) // 31 ETH in gwei
		default:
			return fmt.Errorf("error staking minipool %s: unknown deposit type %d", mpCommon.Address.Hex(), depositType)
		}

		// Get validator deposit data
		depositData, depositDataRoot, err := validator.GetDepositData(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
		if err != nil {
			return fmt.Errorf("error getting deposit data for validator %s: %w", pubkey.Hex(), err)
		}
		signature := types.BytesToValidatorSignature(depositData.Signature)

		txInfo, err := mpCommon.Stake(signature, depositDataRoot, opts)
		if err != nil {
			return fmt.Errorf("error simulating stake transaction for minipool %s: %w", mpCommon.Address.Hex(), err)
		}
		txInfos[i] = txInfo
	}

	data.TxInfos = txInfos
	return nil
}
