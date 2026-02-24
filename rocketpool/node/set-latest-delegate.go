package node

import (
	"fmt"
	"math/big"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Distribute minipools task
type setUseLatestDelegate struct {
	c              *cli.Context
	log            log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              wallet.Wallet
	rp             *rocketpool.RocketPool
	bc             beacon.Client
	d              *client.Client
	gasThreshold   float64
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Create distribute minipools task
func newSetUseLatestDelegate(c *cli.Context, logger log.ColorLogger) (*setUseLatestDelegate, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetHdWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	d, err := services.GetDocker(c)
	if err != nil {
		return nil, err
	}

	// Get the user-requested max fee
	maxFeeGwei := cfg.Smartnode.ManualMaxFee.Value.(float64)
	var maxFee *big.Int
	if maxFeeGwei == 0 {
		maxFee = nil
	} else {
		maxFee = eth.GweiToWei(maxFeeGwei)
	}

	// Get the user-requested max fee
	priorityFeeGwei := cfg.Smartnode.PriorityFee.Value.(float64)
	var priorityFee *big.Int
	if priorityFeeGwei == 0 {
		logger.Printlnf("WARNING: priority fee was missing or 0, setting a default of %.2f.", rpgas.DefaultPriorityFeeGwei)
		priorityFee = eth.GweiToWei(rpgas.DefaultPriorityFeeGwei)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	gasThreshold := cfg.Smartnode.AutoTxGasThreshold.Value.(float64)

	// Return task
	return &setUseLatestDelegate{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		bc:             bc,
		d:              d,
		gasThreshold:   gasThreshold,
		maxFee:         maxFee,
		maxPriorityFee: priorityFee,
		gasLimit:       0,
	}, nil

}

// Distribute minipools
func (t *setUseLatestDelegate) run(state *state.NetworkState) error {
	// Log
	t.log.Println("Checking for minipools to set use latest delegate...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get non-finalised minipools that don't have use latest delegate set
	minipools, err := t.getSettableMinipools(nodeAccount.Address, state, opts)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) can have their use latest delegate set...", len(minipools))

	// Set use latest delegate for minipools
	successCount := 0
	for _, mpd := range minipools {
		success, err := t.setUseLatestDelegate(mpd, opts)
		if err != nil {
			t.log.Println(fmt.Errorf("Could not set use latest delegate for minipool %s: %w", mpd.MinipoolAddress.Hex(), err))
			return err
		}
		if success {
			successCount++
		}
	}

	// Return
	return nil

}

// Get minipools that can have use latest delegate set
func (t *setUseLatestDelegate) getSettableMinipools(nodeAddress common.Address, state *state.NetworkState, opts *bind.CallOpts) ([]*rpstate.NativeMinipoolDetails, error) {

	// Filter minipools by status
	settableMinipools := []*rpstate.NativeMinipoolDetails{}
	for _, mpd := range state.MinipoolDetailsByNode[nodeAddress] {
		if mpd.Finalised {
			// Ignore minipools that are finalised
			continue
		}
		if mpd.UseLatestDelegate {
			// Ignore minipools that already have use latest delegate set
			continue
		}
		settableMinipools = append(settableMinipools, mpd)
	}

	// Return
	return settableMinipools, nil

}

// Set use latest delegate for a minipool
func (t *setUseLatestDelegate) setUseLatestDelegate(mpd *rpstate.NativeMinipoolDetails, callOpts *bind.CallOpts) (bool, error) {

	// Log
	t.log.Printlnf("Setting use latest delegate for minipool %s...", mpd.MinipoolAddress.Hex())

	mp, err := minipool.NewMinipoolFromVersion(t.rp, mpd.MinipoolAddress, mpd.Version, callOpts)
	if err != nil {
		return false, fmt.Errorf("cannot create binding for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return false, err
	}

	// Get the gas limit
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		return false, fmt.Errorf("minipool %s cannot be converted to v3 (current version: %d)", mpd.MinipoolAddress.Hex(), mp.GetVersion())
	}
	gasInfo, err := mpv3.EstimateSetUseLatestDelegateGas(opts)
	if err != nil {
		return false, fmt.Errorf("Could not estimate the gas required to distribute minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
	}
	var gas *big.Int
	if t.gasLimit != 0 {
		gas = new(big.Int).SetUint64(t.gasLimit)
	} else {
		gas = new(big.Int).SetUint64(gasInfo.SafeGasLimit)
	}

	// Get the max fee
	maxFee := t.maxFee
	if maxFee == nil || maxFee.Uint64() == 0 {
		maxFee, err = rpgas.GetHeadlessMaxFeeWeiWithLatestBlock(t.cfg, t.rp)
		if err != nil {
			return false, err
		}
	}

	// Print the gas info
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return false, nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = GetPriorityFee(t.maxPriorityFee, maxFee)
	opts.GasLimit = gas.Uint64()

	// Distribute minipool
	hash, err := mpv3.SetUseLatestDelegate(opts)
	if err != nil {
		return false, err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return false, err
	}

	// Log
	t.log.Printlnf("Successfully set use latest delegate for minipool %s.", mp.GetAddress().Hex())

	// Return
	return true, nil

}
