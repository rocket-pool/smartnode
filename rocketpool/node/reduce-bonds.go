package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Reduce bonds task
type reduceBonds struct {
	c              *cli.Context
	log            log.ColorLogger
	cfg            *config.RocketPoolConfig
	w              *wallet.Wallet
	rp             *rocketpool.RocketPool
	d              *client.Client
	gasThreshold   float64
	disabled       bool
	maxFee         *big.Int
	maxPriorityFee *big.Int
	gasLimit       uint64
}

// Create reduce bonds task
func newReduceBonds(c *cli.Context, logger log.ColorLogger) (*reduceBonds, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	d, err := services.GetDocker(c)
	if err != nil {
		return nil, err
	}

	// Check if auto-bond-reduction is disabled
	gasThreshold := cfg.Smartnode.AutoTxGasThreshold.Value.(float64)
	disabled := false
	if gasThreshold == 0 {
		logger.Println("Automatic tx gas threshold is 0, disabling auto-reduce.")
		disabled = true
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
	return &reduceBonds{
		c:              c,
		log:            logger,
		cfg:            cfg,
		w:              w,
		rp:             rp,
		d:              d,
		gasThreshold:   gasThreshold,
		disabled:       disabled,
		maxFee:         maxFee,
		maxPriorityFee: priorityFee,
		gasLimit:       0,
	}, nil

}

// Reduce bonds
func (t *reduceBonds) run(state *state.NetworkState) error {

	// Check if auto-reduce is disabled
	if t.disabled {
		return nil
	}

	// Log
	t.log.Println("Checking for minipool bonds to reduce...")

	// Get the latest state
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(state.ElBlockNumber),
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get the bond reduction details
	windowStart := state.NetworkDetails.BondReductionWindowStart
	windowLength := state.NetworkDetails.BondReductionWindowLength

	// Get the time of the latest block
	latestEth1Block, err := t.rp.Client.HeaderByNumber(context.Background(), opts.BlockNumber)
	if err != nil {
		return fmt.Errorf("can't get the latest block time: %w", err)
	}
	latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

	// Get reduceable minipools
	minipools, err := t.getReduceableMinipools(nodeAccount.Address, windowStart, windowLength, latestBlockTime, state, opts)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) are ready for bond reduction...", len(minipools))

	// Workaround for the fee distribution issue
	success, err := t.forceFeeDistribution()
	if err != nil {
		return err
	}
	if !success {
		return nil
	}

	// Reduce bonds
	successCount := 0
	for _, mp := range minipools {
		success, err := t.reduceBond(mp, windowStart, windowLength, latestBlockTime, opts)
		if err != nil {
			t.log.Println(fmt.Errorf("could not reduce bond for minipool %s: %w", mp.MinipoolAddress.Hex(), err))
			return err
		}
		if success {
			successCount++
		}
	}

	// Return
	return nil

}

// Temp mitigation for the
func (t *reduceBonds) forceFeeDistribution() (bool, error) {

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return false, err
	}

	// Get fee distributor
	distributorAddress, err := node.GetDistributorAddress(t.rp, nodeAccount.Address, nil)
	if err != nil {
		return false, err
	}
	distributor, err := node.NewDistributor(t.rp, distributorAddress, nil)
	if err != nil {
		return false, err
	}

	// Sync
	var wg errgroup.Group
	var balanceRaw *big.Int
	var nodeShare float64

	// Get the contract's balance
	wg.Go(func() error {
		var err error
		balanceRaw, err = t.rp.Client.BalanceAt(context.Background(), distributorAddress, nil)
		return err
	})

	// Get the node share of the balance
	wg.Go(func() error {
		nodeShareRaw, err := distributor.GetNodeShare(nil)
		if err != nil {
			return fmt.Errorf("error getting node share for distributor %s: %w", distributorAddress.Hex(), err)
		}
		nodeShare = eth.WeiToEth(nodeShareRaw)
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return false, err
	}

	balance := eth.WeiToEth(balanceRaw)
	if balance == 0 {
		t.log.Println("Your fee distributor does not have any ETH and does not need to be distributed.")
		return true, nil
	}
	t.log.Println("NOTE: prior to bond reduction, you must distribute the funds in your fee distributor.")

	// Print info
	rEthShare := balance - nodeShare
	t.log.Printlnf("Your fee distributor's balance of %.6f ETH will be distributed as follows:\n", balance)
	t.log.Printlnf("\tYour withdrawal address will receive %.6f ETH.", nodeShare)
	t.log.Printlnf("\trETH pool stakers will receive %.6f ETH.\n", rEthShare)

	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return false, err
	}

	// Get the gas limit
	gasInfo, err := distributor.EstimateDistributeGas(opts)
	if err != nil {
		return false, fmt.Errorf("could not estimate the gas required to distribute node fees: %w", err)
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
		maxFee, err = rpgas.GetHeadlessMaxFeeWei()
		if err != nil {
			return false, err
		}
	}

	// Print the gas info
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		return false, nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee
	opts.GasLimit = gas.Uint64()

	// Distribute
	fmt.Printf("Distributing rewards...\n")
	hash, err := distributor.Distribute(opts)
	if err != nil {
		return false, err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return false, err
	}

	// Log & return
	fmt.Println("Successfully distributed your fee distributor's balance. Your rewards should arrive in your withdrawal address shortly.")
	return true, nil
}

// Get reduceable minipools
func (t *reduceBonds) getReduceableMinipools(nodeAddress common.Address, windowStart time.Duration, windowLength time.Duration, latestBlockTime time.Time, state *state.NetworkState, opts *bind.CallOpts) ([]*rpstate.NativeMinipoolDetails, error) {

	// Filter minipools
	reduceableMinipools := []*rpstate.NativeMinipoolDetails{}
	for _, mpd := range state.MinipoolDetailsByNode[nodeAddress] {

		// TEMP
		reduceBondTime, err := minipool.GetReduceBondTime(t.rp, mpd.MinipoolAddress, opts)
		if err != nil {
			return nil, fmt.Errorf("error getting reduce bond time for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
		}
		reduceBondCancelled, err := minipool.GetReduceBondCancelled(t.rp, mpd.MinipoolAddress, opts)
		if err != nil {
			return nil, fmt.Errorf("error getting reduce bond cancelled for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
		}

		depositBalance := eth.WeiToEth(mpd.NodeDepositBalance)
		timeSinceReductionStart := latestBlockTime.Sub(reduceBondTime)

		if depositBalance == 16 &&
			timeSinceReductionStart < (windowStart+windowLength) &&
			!reduceBondCancelled &&
			mpd.Status == types.Staking {
			if timeSinceReductionStart > windowStart {
				reduceableMinipools = append(reduceableMinipools, mpd)
			} else {
				remainingTime := windowStart - timeSinceReductionStart
				t.log.Printlnf("Minipool %s has %s left until it can have its bond reduced.", mpd.MinipoolAddress.Hex(), remainingTime)
			}
		}
	}

	// Return
	return reduceableMinipools, nil

}

// Reduce a minipool's bond
func (t *reduceBonds) reduceBond(mpd *rpstate.NativeMinipoolDetails, windowStart time.Duration, windowLength time.Duration, latestBlockTime time.Time, callOpts *bind.CallOpts) (bool, error) {

	// Log
	t.log.Printlnf("Reducing bond for minipool %s...", mpd.MinipoolAddress.Hex())

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return false, err
	}

	// Make the minipool binding
	mpBinding, err := minipool.NewMinipoolFromVersion(t.rp, mpd.MinipoolAddress, mpd.Version, callOpts)
	if err != nil {
		return false, fmt.Errorf("error creating minipool binding for %s: %w", mpd.MinipoolAddress.Hex(), err)
	}

	// Get the updated minipool interface
	mpv3, success := minipool.GetMinipoolAsV3(mpBinding)
	if !success {
		return false, fmt.Errorf("cannot reduce bond for minipool %s because its delegate version is too low (v%d); please update the delegate", mpBinding.GetAddress().Hex(), mpBinding.GetVersion())
	}

	// Get the gas limit
	gasInfo, err := mpv3.EstimateReduceBondAmountGas(opts)
	if err != nil {
		return false, fmt.Errorf("could not estimate the gas required to reduce bond: %w", err)
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
		maxFee, err = rpgas.GetHeadlessMaxFeeWei()
		if err != nil {
			return false, err
		}
	}

	// TEMP
	reduceBondTime, err := minipool.GetReduceBondTime(t.rp, mpd.MinipoolAddress, callOpts)
	if err != nil {
		return false, fmt.Errorf("error getting reduce bond time for minipool %s: %w", mpd.MinipoolAddress.Hex(), err)
	}

	// Print the gas info
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, &t.log, maxFee, t.gasLimit) {
		timeSinceReductionStart := latestBlockTime.Sub(reduceBondTime)
		remainingTime := (windowStart + windowLength) - timeSinceReductionStart
		t.log.Printlnf("Time until bond reduction times out: %s", remainingTime)
		return false, nil
	}

	opts.GasFeeCap = maxFee
	opts.GasTipCap = t.maxPriorityFee
	opts.GasLimit = gas.Uint64()

	// Reduce bond
	hash, err := mpv3.ReduceBondAmount(opts)
	if err != nil {
		return false, err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		return false, err
	}

	// Log
	t.log.Printlnf("Successfully reduced bond for minipool %s.", mpd.MinipoolAddress.Hex())

	// Return
	return true, nil

}
