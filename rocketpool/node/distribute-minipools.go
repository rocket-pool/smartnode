package node

import (
	"fmt"
	"math/big"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/alerting"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Distribute minipools task
type distributeMinipools struct {
	c                   *cli.Context
	log                 log.ColorLogger
	cfg                 *config.RocketPoolConfig
	w                   wallet.Wallet
	rp                  *rocketpool.RocketPool
	bc                  beacon.Client
	d                   *client.Client
	gasThreshold        float64
	distributeThreshold *big.Int
	disabled            bool
	eight               *big.Int
	maxFee              *big.Int
	maxPriorityFee      *big.Int
	gasLimit            uint64
}

// Create distribute minipools task
func newDistributeMinipools(c *cli.Context, logger log.ColorLogger) (*distributeMinipools, error) {

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

	// Check if auto-distributing is disabled
	gasThreshold := cfg.Smartnode.AutoTxGasThreshold.Value.(float64)
	distributeThreshold := cfg.Smartnode.DistributeThreshold.Value.(float64)
	disabled := false
	if gasThreshold == 0 {
		logger.Println("Automatic tx gas threshold is 0, disabling auto-distribute.")
		disabled = true
	} else {
		// Safety clamp
		if distributeThreshold >= 8 {
			logger.Printlnf("WARNING: Auto-distribute threshold is more than 8 ETH (%.6f ETH), reducing to 7.5 ETH for safety", distributeThreshold)
			distributeThreshold = 7.5
		} else if distributeThreshold == 0 {
			logger.Println("Auto-distribute threshold is 0, disabling auto-distribute.")
			disabled = true
		}
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
		logger.Println("WARNING: priority fee was missing or 0, setting a default of 2.")
		priorityFee = eth.GweiToWei(2)
	} else {
		priorityFee = eth.GweiToWei(priorityFeeGwei)
	}

	// Return task
	return &distributeMinipools{
		c:                   c,
		log:                 logger,
		cfg:                 cfg,
		w:                   w,
		rp:                  rp,
		bc:                  bc,
		d:                   d,
		gasThreshold:        gasThreshold,
		distributeThreshold: eth.EthToWei(distributeThreshold),
		disabled:            disabled,
		eight:               eth.EthToWei(8),
		maxFee:              maxFee,
		maxPriorityFee:      priorityFee,
		gasLimit:            0,
	}, nil

}

// Distribute minipools
func (t *distributeMinipools) run(state *state.NetworkState) error {

	// Check if auto-distribute is disabled
	if t.disabled {
		return nil
	}

	// Log
	t.log.Println("Checking for minipools to distribute...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get prelaunch minipools
	minipools, err := t.getDistributableMinipools(nodeAccount.Address, state, opts)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) can have their balances distributed...", len(minipools))

	// Distribute minipools
	successCount := 0
	for _, mpd := range minipools {
		success, err := t.distributeMinipool(mpd, opts)
		alerting.AlertMinipoolBalanceDistributed(t.cfg, mpd.MinipoolAddress, err == nil)
		if err != nil {
			t.log.Println(fmt.Errorf("Could not distribute balance of minipool %s: %w", mpd.MinipoolAddress.Hex(), err))
			return err
		}
		if success {
			successCount++
		}
	}

	// Return
	return nil

}

// Get distributable minipools
func (t *distributeMinipools) getDistributableMinipools(nodeAddress common.Address, state *state.NetworkState, opts *bind.CallOpts) ([]*rpstate.NativeMinipoolDetails, error) {

	// Filter minipools by status
	distributableMinipools := []*rpstate.NativeMinipoolDetails{}
	for _, mpd := range state.MinipoolDetailsByNode[nodeAddress] {
		if mpd.Status != rptypes.Staking || mpd.Finalised {
			// Ignore minipools that aren't staking and non-finalized
			continue
		}
		if mpd.Version < 3 {
			// Ignore minipools with legacy delegates
			continue
		}
		if mpd.DistributableBalance.Cmp(t.eight) >= 0 {
			// Ignore minipools with distributable balances >= 8 ETH
			continue
		}
		if mpd.DistributableBalance.Cmp(t.distributeThreshold) >= 0 {
			distributableMinipools = append(distributableMinipools, mpd)
		}
	}

	// Return
	return distributableMinipools, nil

}

// Distribute a minipool
func (t *distributeMinipools) distributeMinipool(mpd *rpstate.NativeMinipoolDetails, callOpts *bind.CallOpts) (bool, error) {

	// Log
	t.log.Printlnf("Distributing minipool %s (total balance of %.6f ETH)...", mpd.MinipoolAddress.Hex(), eth.WeiToEth(mpd.Balance))

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
	gasInfo, err := mpv3.EstimateDistributeBalanceGas(true, opts)
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
		maxFee, err = rpgas.GetHeadlessMaxFeeWei(t.cfg)
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
	hash, err := mpv3.DistributeBalance(true, opts)
	if err != nil {
		return false, err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return false, err
	}

	// Log
	t.log.Printlnf("Successfully distributed balance of minipool %s.", mp.GetAddress().Hex())

	// Return
	return true, nil

}
