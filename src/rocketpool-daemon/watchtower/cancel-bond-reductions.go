package watchtower

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/collectors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/node-manager-core/utils/log"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/config"
)

const (
	scrubBuffer uint64 = 10000000 // 0.01 ETH
)

type CancelBondReductions struct {
	sp               *services.ServiceProvider
	log              log.ColorLogger
	errLog           log.ColorLogger
	cfg              *config.SmartNodeConfig
	w                *wallet.Wallet
	rp               *rocketpool.RocketPool
	ec               eth.IExecutionClient
	mpMgr            *minipool.MinipoolManager
	coll             *collectors.BondReductionCollector
	lock             *sync.Mutex
	isRunning        bool
	generationPrefix string
}

// Create cancel bond reductions task
func NewCancelBondReductions(sp *services.ServiceProvider, logger log.ColorLogger, errorLogger log.ColorLogger, coll *collectors.BondReductionCollector) *CancelBondReductions {
	lock := &sync.Mutex{}
	return &CancelBondReductions{
		sp:               sp,
		log:              logger,
		errLog:           errorLogger,
		coll:             coll,
		lock:             lock,
		isRunning:        false,
		generationPrefix: "[Bond Reduction]",
	}
}

// Start the bond reduction cancellation thread
func (t *CancelBondReductions) Run(state *state.NetworkState) error {
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
func (t *CancelBondReductions) checkBondReductions(state *state.NetworkState) error {
	// Get services
	t.cfg = t.sp.GetConfig()
	t.w = t.sp.GetWallet()
	t.rp = t.sp.GetRocketPool()
	t.ec = t.sp.GetEthClient()
	var err error
	t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

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
		validator := state.ValidatorDetails[mpd.Pubkey]
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
					t.cancelBondReduction(state, mpd.MinipoolAddress, fmt.Sprintf("minipool balance is %d (below the threshold)", validator.Balance))
					balanceTooLowCount += 1
				}

			case beacon.ValidatorState_ActiveExiting,
				beacon.ValidatorState_ActiveSlashed,
				beacon.ValidatorState_ExitedUnslashed,
				beacon.ValidatorState_ExitedSlashed,
				beacon.ValidatorState_WithdrawalPossible,
				beacon.ValidatorState_WithdrawalDone:
				t.cancelBondReduction(state, mpd.MinipoolAddress, "minipool is already slashed, exiting, or exited")
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
func (t *CancelBondReductions) updateMetricsCollector(state *state.NetworkState, minipoolCount float64, invalidStateCount float64, balanceTooLowCount float64) {
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
func (t *CancelBondReductions) cancelBondReduction(state *state.NetworkState, address common.Address, reason string) {
	// Log
	t.printMessage("=== CANCELLING BOND REDUCTION ===")
	t.printMessage(fmt.Sprintf("Minipool: %s", address.Hex()))
	t.printMessage(fmt.Sprintf("Reason:   %s", reason))
	t.printMessage("=================================")

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		t.printMessage(fmt.Sprintf("error getting node account transactor: %s", err.Error()))
		return
	}

	// Make the binding
	mpd := state.MinipoolDetailsByAddress[address]
	mp, err := t.mpMgr.NewMinipoolFromVersion(address, mpd.Version)
	if err != nil {
		t.printMessage(fmt.Sprintf("error creating binding for minipool %s: %s", address.Hex(), err.Error()))
		return
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		t.printMessage(fmt.Sprintf("error converting minipool %s to v3: %s", address.Hex(), err.Error()))
	}

	// Get the tx info
	txInfo, err := mpv3.VoteCancelReduction(opts)
	if err != nil {
		t.printMessage(fmt.Sprintf("error getting voteCancelReduction tx for minipool %s: %s", address.Hex(), err.Error()))
		return
	}
	if txInfo.SimulationResult.SimulationError != "" {
		t.printMessage(fmt.Sprintf("simulating voteCancelReduction tx for minipool %s failed: %s", address.Hex(), txInfo.SimulationResult.SimulationError))
		return
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, &t.log, maxFee, 0) {
		return
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, &t.log, txInfo, opts)
	if err != nil {
		t.printMessage(fmt.Sprintf("error waiting for cancel transaction: %s", err.Error()))
		return
	}

	// Log
	t.log.Printlnf("Successfully voted to cancel the bond reduction of minipool %s.", address.Hex())
}

func (t *CancelBondReductions) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Bond reduction cancel check failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Print a message from the tree generation goroutine
func (t *CancelBondReductions) printMessage(message string) {
	t.log.Printlnf("%s %s", t.generationPrefix, message)
}
