package watchtower

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/rocket-pool/smartnode/rocketpool/watchtower/collectors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	rpstate "github.com/rocket-pool/smartnode/bindings/utils/state"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

const (
	scrubBuffer uint64 = 10000000 // 0.01 ETH
)

type cancelBondReductions struct {
	c                *cli.Context
	log              log.ColorLogger
	errLog           log.ColorLogger
	cfg              *config.RocketPoolConfig
	w                wallet.Wallet
	rp               *rocketpool.RocketPool
	ec               rocketpool.ExecutionClient
	coll             *collectors.BondReductionCollector
	lock             *sync.Mutex
	isRunning        bool
	generationPrefix string
}

// Create cancel bond reductions task
func newCancelBondReductions(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger, coll *collectors.BondReductionCollector) (*cancelBondReductions, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetHdWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Return task
	lock := &sync.Mutex{}
	return &cancelBondReductions{
		c:                c,
		log:              logger,
		errLog:           errorLogger,
		cfg:              cfg,
		w:                w,
		rp:               rp,
		ec:               ec,
		coll:             coll,
		lock:             lock,
		isRunning:        false,
		generationPrefix: "[Bond Reduction]",
	}, nil

}

// Start the bond reduction cancellation thread
func (t *cancelBondReductions) run(state *state.NetworkState) error {

	if state.IsSaturnDeployed {
		return nil
	}

	// Wait for eth clients to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
		return err
	}

	// Log
	t.log.Println("Checking for bond reductions to cancel...")

	// Check if the check is already running
	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Bond reduction cancel check is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Run the check
	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()
		t.printMessage("Starting bond reduction cancel check in a separate thread.")

		err := t.checkBondReductions(state)
		if err != nil {
			t.handleError(fmt.Errorf("%s %w", t.generationPrefix, err))
			return
		}

		t.lock.Lock()
		t.isRunning = false
		t.lock.Unlock()
	}()

	// Return
	return nil

}

// Check for bond reductions to cancel
func (t *cancelBondReductions) checkBondReductions(state *state.NetworkState) error {

	t.printMessage(fmt.Sprintf("Checking for Beacon slot %d (EL block %d)", state.BeaconSlotNumber, state.ElBlockNumber))

	// Check if any of the minipools have bond reduction requests
	zero := big.NewInt(0)
	reductionMps := []*rpstate.NativeMinipoolDetails{}
	for i, mpd := range state.MinipoolDetails {
		if mpd.ReduceBondTime.Cmp(zero) == 1 {
			reductionMps = append(reductionMps, &state.MinipoolDetails[i])
		}
	}

	// If there aren't any, return
	if len(reductionMps) == 0 {
		t.printMessage("No minipools have requested a bond reduction.")
		return nil
	}

	// Metrics
	balanceTooLowCount := float64(0)
	invalidStateCount := float64(0)

	// Check the status of each one
	threshold := uint64(32000000000) - scrubBuffer
	for _, mpd := range reductionMps {
		validator := state.MinipoolValidatorDetails[mpd.Pubkey]
		if validator.Exists {
			switch validator.Status {
			case beacon.ValidatorState_PendingInitialized,
				beacon.ValidatorState_PendingQueued:
				// Do nothing because this validator isn't live yet
				continue

			case beacon.ValidatorState_ActiveOngoing:
				// Check the balance
				if validator.Balance < threshold {
					// Cancel because it's under-balance
					t.cancelBondReduction(mpd.MinipoolAddress, fmt.Sprintf("minipool balance is %d (below the threshold)", validator.Balance))
					balanceTooLowCount += 1
				}

			case beacon.ValidatorState_ActiveExiting,
				beacon.ValidatorState_ActiveSlashed,
				beacon.ValidatorState_ExitedUnslashed,
				beacon.ValidatorState_ExitedSlashed,
				beacon.ValidatorState_WithdrawalPossible,
				beacon.ValidatorState_WithdrawalDone:
				t.cancelBondReduction(mpd.MinipoolAddress, "minipool is already slashed, exiting, or exited")
				invalidStateCount += 1

			default:
				t.updateMetricsCollector(state, float64(len(reductionMps)), invalidStateCount, balanceTooLowCount)
				return fmt.Errorf("unknown validator state: %v", validator.Status)
			}
		}
	}

	t.updateMetricsCollector(state, float64(len(reductionMps)), invalidStateCount, balanceTooLowCount)

	return nil

}

// Update the bond reduction metrics collector
func (t *cancelBondReductions) updateMetricsCollector(state *state.NetworkState, minipoolCount float64, invalidStateCount float64, balanceTooLowCount float64) {
	if t.coll != nil {
		t.coll.UpdateLock.Lock()
		defer t.coll.UpdateLock.Unlock()

		// Get the time of the state's EL block
		genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
		secondsSinceGenesis := time.Duration(state.BeaconSlotNumber*state.BeaconConfig.SecondsPerSlot) * time.Second
		stateBlockTime := genesisTime.Add(secondsSinceGenesis)

		t.coll.LatestBlockTime = float64(stateBlockTime.Unix())
		t.coll.TotalMinipools = float64(minipoolCount)
		t.coll.InvalidState = invalidStateCount
		t.coll.BalanceTooLow = balanceTooLowCount
	}
}

// Cancel a bond reduction
func (t *cancelBondReductions) cancelBondReduction(address common.Address, reason string) {

	// Log
	t.printMessage("=== CANCELLING BOND REDUCTION ===")
	t.printMessage(fmt.Sprintf("Minipool: %s", address.Hex()))
	t.printMessage(fmt.Sprintf("Reason:   %s", reason))
	t.printMessage("=================================")

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		t.printMessage(fmt.Sprintf("error getting node account transactor: %s", err.Error()))
		return
	}

	// Get the gas limit
	gasInfo, err := minipool.EstimateVoteCancelReductionGas(t.rp, address, opts)
	if err != nil {
		t.printMessage(fmt.Sprintf("could not estimate the gas required to voteCancelReduction the minipool: %s", err.Error()))
		return
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, &t.log, maxFee, 0) {
		return
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = gasInfo.SafeGasLimit

	// Cancel the reduction
	hash, err := minipool.VoteCancelReduction(t.rp, address, opts)
	if err != nil {
		t.printMessage(fmt.Sprintf("could not vote to cancel bond reduction: %s", err.Error()))
		return
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		t.printMessage(fmt.Sprintf("error waiting for cancel transaction: %s", err.Error()))
		return
	}

	// Log
	t.log.Printlnf("Successfully voted to cancel the bond reduction of minipool %s.", address.Hex())

}

func (t *cancelBondReductions) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Bond reduction cancel check failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Print a message from the tree generation goroutine
func (t *cancelBondReductions) printMessage(message string) {
	t.log.Printlnf("%s %s", t.generationPrefix, message)
}
