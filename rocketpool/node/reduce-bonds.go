package node

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	rpgas "github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
)

// The fraction of the timeout period to trigger overdue transactions
const reduceBondTimeoutSafetyFactor int = 2

// Reduce bonds task
type reduceBonds struct {
	c               *cli.Context
	log             log.ColorLogger
	cfg             *config.RocketPoolConfig
	w               *wallet.Wallet
	rp              *rocketpool.RocketPool
	bc              beacon.Client
	d               *client.Client
	gasThreshold    float64
	maxFee          *big.Int
	maxPriorityFee  *big.Int
	gasLimit        uint64
	isAtlasDeployed bool
}

// Details required to check for bond reduction eligibility
type minipoolBondReductionDetails struct {
	Address             common.Address
	DepositBalance      *big.Int
	ReduceBondTime      time.Time
	ReduceBondCancelled bool
	Status              types.MinipoolStatus
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	d, err := services.GetDocker(c)
	if err != nil {
		return nil, err
	}

	// Check if auto-staking is disabled
	gasThreshold := cfg.Smartnode.MinipoolStakeGasThreshold.Value.(float64)

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
		c:               c,
		log:             logger,
		cfg:             cfg,
		w:               w,
		rp:              rp,
		bc:              bc,
		d:               d,
		gasThreshold:    gasThreshold,
		maxFee:          maxFee,
		maxPriorityFee:  priorityFee,
		gasLimit:        0,
		isAtlasDeployed: false,
	}, nil

}

// Reduce bonds
func (t *reduceBonds) run() error {

	// Reload the wallet (in case a call to `node deposit` changed it)
	if err := t.w.Reload(); err != nil {
		return err
	}

	// Wait for eth client to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}

	// Check if Atlas has been deployed yet
	if !t.isAtlasDeployed {
		isAtlasDeployed, err := rp.IsAtlasDeployed(t.rp)
		if err != nil {
			return fmt.Errorf("error checking if Atlas is deployed: %w", err)
		}
		if isAtlasDeployed {
			t.isAtlasDeployed = true
		} else {
			return nil
		}
	}

	// Log
	t.log.Println("Checking for minipool bonds to reduce...")

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get the bond reduction details
	windowStartRaw, err := trustednode.GetBondReductionWindowStart(t.rp, nil)
	if err != nil {
		return fmt.Errorf("error getting bond reduction window start: %w", err)
	}
	windowStart := time.Duration(windowStartRaw) * time.Second
	windowLengthRaw, err := trustednode.GetBondReductionWindowLength(t.rp, nil)
	if err != nil {
		return fmt.Errorf("error getting bond reduction window length: %w", err)
	}
	windowLength := time.Duration(windowLengthRaw) * time.Second

	// Get the time of the latest block
	latestEth1Block, err := t.rp.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("can't get the latest block time: %w", err)
	}
	latestBlockTime := time.Unix(int64(latestEth1Block.Time), 0)

	// Get reduceable minipools
	minipools, err := t.getReduceableMinipools(nodeAccount.Address, windowStart, windowLength, latestBlockTime)
	if err != nil {
		return err
	}
	if len(minipools) == 0 {
		return nil
	}

	// Log
	t.log.Printlnf("%d minipool(s) are ready for bond reduction...", len(minipools))

	// Reduce bonds
	successCount := 0
	for _, mp := range minipools {
		success, err := t.reduceBond(mp, windowStart, windowLength, latestBlockTime)
		if err != nil {
			t.log.Println(fmt.Errorf("could not reduce bond for minipool %s: %w", mp.Address.Hex(), err))
			return err
		}
		if success {
			successCount++
		}
	}

	// Return
	return nil

}

// Get reduceable minipools
func (t *reduceBonds) getReduceableMinipools(nodeAddress common.Address, windowStart time.Duration, windowLength time.Duration, latestBlockTime time.Time) ([]minipoolBondReductionDetails, error) {

	// Get node minipool addresses
	addresses, err := minipool.GetNodeMinipoolAddresses(t.rp, nodeAddress, nil)
	if err != nil {
		return nil, err
	}

	// Create minipool contracts
	minipools := make([]minipool.Minipool, len(addresses))
	for mi, address := range addresses {
		mp, err := minipool.NewMinipool(t.rp, address, nil)
		if err != nil {
			return nil, err
		}
		minipools[mi] = mp
	}

	// Data
	var wg errgroup.Group
	details := make([]minipoolBondReductionDetails, len(minipools))

	// Load minipool details
	for mi, mp := range minipools {
		mi, mp := mi, mp
		wg.Go(func() error {
			details[mi].Address = mp.GetAddress()

			depositBalance, err := mp.GetNodeDepositBalance(nil)
			if err != nil {
				return fmt.Errorf("error getting node deposit balance for minipool %s: %w", mp.GetAddress().Hex(), err)
			}
			details[mi].DepositBalance = depositBalance

			reduceBondTime, err := minipool.GetReduceBondTime(t.rp, mp.GetAddress(), nil)
			if err != nil {
				return fmt.Errorf("error getting bond reduction time for minipool %s: %w", mp.GetAddress().Hex(), err)
			}
			details[mi].ReduceBondTime = reduceBondTime

			reduceBondCancelled, err := minipool.GetReduceBondCancelled(t.rp, mp.GetAddress(), nil)
			if err != nil {
				return fmt.Errorf("error getting bond reduction cancel status for minipool %s: %w", mp.GetAddress().Hex(), err)
			}
			details[mi].ReduceBondCancelled = reduceBondCancelled

			status, err := mp.GetStatus(nil)
			if err != nil {
				return fmt.Errorf("error getting status for minipool %s: %w", mp.GetAddress().Hex(), err)
			}
			details[mi].Status = status

			return nil
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Filter minipools
	reduceableMinipools := []minipoolBondReductionDetails{}
	for mi, mp := range minipools {

		minipoolDetails := details[mi]
		depositBalance := eth.WeiToEth(minipoolDetails.DepositBalance)
		timeSinceReductionStart := latestBlockTime.Sub(minipoolDetails.ReduceBondTime)

		if depositBalance == 16 &&
			timeSinceReductionStart < (windowStart+windowLength) &&
			!minipoolDetails.ReduceBondCancelled &&
			minipoolDetails.Status == types.Staking {
			if timeSinceReductionStart > windowStart {
				reduceableMinipools = append(reduceableMinipools, minipoolDetails)
			} else {
				remainingTime := windowStart - timeSinceReductionStart
				t.log.Printlnf("Minipool %s has %s left until it can have its bond reduced.", mp.GetAddress().Hex(), remainingTime)
			}
		}
	}

	// Return
	return reduceableMinipools, nil

}

// Reduce a minipool's bond
func (t *reduceBonds) reduceBond(mp minipoolBondReductionDetails, windowStart time.Duration, windowLength time.Duration, latestBlockTime time.Time) (bool, error) {

	// Log
	t.log.Printlnf("Reducing bond for minipool %s...", mp.Address.Hex())

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return false, err
	}

	// Make the minipool binding
	mpBinding, err := minipool.NewMinipool(t.rp, mp.Address, nil)
	if err != nil {
		return false, fmt.Errorf("error creating minipool binding for %s: %w", mp.Address.Hex(), err)
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

	// Print the gas info
	if !api.PrintAndCheckGasInfo(gasInfo, true, t.gasThreshold, t.log, maxFee, t.gasLimit) {
		// Check for the timeout buffer
		timeSinceReductionStart := latestBlockTime.Sub(mp.ReduceBondTime)
		remainingTime := (windowStart + windowLength) - timeSinceReductionStart
		isDue := remainingTime < (windowLength / time.Duration(reduceBondTimeoutSafetyFactor))
		if !isDue {
			timeUntilDue := remainingTime - (windowLength / time.Duration(reduceBondTimeoutSafetyFactor))
			t.log.Printlnf("Time until bond reduction will be forced: %s", timeUntilDue)
			return false, nil
		}

		t.log.Println("NOTICE: The minipool has exceeded half of the timeout period, so its bond reduction will be forced at the current gas price.")
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
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return false, err
	}

	// Log
	t.log.Printlnf("Successfully reduced bond for minipool %s.", mp.Address.Hex())

	// Return
	return true, nil

}
