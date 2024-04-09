package watchtower

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower/collectors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rpstate "github.com/rocket-pool/rocketpool-go/v2/utils/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

const (
	scrubBuffer uint64 = 10000000 // 0.01 ETH
)

type CancelBondReductions struct {
	ctx       context.Context
	sp        *services.ServiceProvider
	logger    *slog.Logger
	cfg       *config.SmartNodeConfig
	w         *wallet.Wallet
	rp        *rocketpool.RocketPool
	ec        eth.IExecutionClient
	mpMgr     *minipool.MinipoolManager
	coll      *collectors.BondReductionCollector
	lock      *sync.Mutex
	isRunning bool
}

// Create cancel bond reductions task
func NewCancelBondReductions(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger, coll *collectors.BondReductionCollector) *CancelBondReductions {
	lock := &sync.Mutex{}
	return &CancelBondReductions{
		ctx:       ctx,
		sp:        sp,
		logger:    logger.With(slog.String(keys.RoutineKey, "Bond Reduction")),
		cfg:       sp.GetConfig(),
		w:         sp.GetWallet(),
		rp:        sp.GetRocketPool(),
		ec:        sp.GetEthClient(),
		coll:      coll,
		lock:      lock,
		isRunning: false,
	}
}

// Start the bond reduction cancellation thread
func (t *CancelBondReductions) Run(state *state.NetworkState) error {
	// Log
	t.logger.Info("Starting bond reduction check.")

	// Check if the check is already running
	t.lock.Lock()
	if t.isRunning {
		t.logger.Info("Bond reduction cancel check is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Run the check
	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()

		err := t.checkBondReductions(state)
		if err != nil {
			t.handleError(err)
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
	// Update contract bindings
	var err error
	t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

	t.logger.Info(fmt.Sprintf("Checking bond reductions... %d (EL block %d)", slog.Uint64(keys.SlotKey, state.BeaconSlotNumber), slog.Uint64(keys.BlockKey, state.ElBlockNumber)))

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
		t.logger.Info("No minipools have requested a bond reduction.")
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
	mpLogger := t.logger.With(slog.String(keys.MinipoolKey, address.Hex()))
	mpLogger.Warn("=== CANCELLING BOND REDUCTION ===", slog.String(keys.ReasonKey, reason))

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		mpLogger.Error("Error getting node account transactor", log.Err(err))
		return
	}

	// Make the binding
	mpd := state.MinipoolDetailsByAddress[address]
	mp, err := t.mpMgr.NewMinipoolFromVersion(address, mpd.Version)
	if err != nil {
		mpLogger.Error("Error creating minipool binding", log.Err(err))
		return
	}
	mpv3, success := minipool.GetMinipoolAsV3(mp)
	if !success {
		mpLogger.Error("Error converting minipool to v3", log.Err(err))
		return
	}

	// Get the tx info
	txInfo, err := mpv3.VoteCancelReduction(opts)
	if err != nil {
		mpLogger.Error("Error getting VoteCancelReduction TX for minipool", log.Err(err))
		return
	}
	if txInfo.SimulationResult.SimulationError != "" {
		mpLogger.Error("Simulating VoteCancelReduction TX failed", slog.String(log.ErrorKey, txInfo.SimulationResult.SimulationError))
		return
	}

	// Print the gas info
	maxFee := eth.GweiToWei(utils.GetWatchtowerMaxFee(t.cfg))
	if !gas.PrintAndCheckGasInfo(txInfo.SimulationResult, false, 0, t.logger, maxFee, 0) {
		return
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(utils.GetWatchtowerPrioFee(t.cfg))
	opts.GasLimit = txInfo.SimulationResult.SafeGasLimit

	// Print TX info and wait for it to be included in a block
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, t.logger, txInfo, opts)
	if err != nil {
		mpLogger.Error("Error waiting for cancel transaction", log.Err(err))
		return
	}

	// Log
	t.logger.Info("Successfully voted to cancel bond reduction.", slog.String(keys.MinipoolKey, address.Hex()))
}

func (t *CancelBondReductions) handleError(err error) {
	t.logger.Error("*** Bond reduction cancel check failed. ***", log.Err(err))
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}
