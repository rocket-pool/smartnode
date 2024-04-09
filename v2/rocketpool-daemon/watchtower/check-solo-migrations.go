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
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
)

const (
	soloMigrationCheckThreshold float64 = 0.85 // Fraction of PromotionStakePeriod that can go before a minipool gets scrubbed for not having changed to 0x01
	blsPrefix                   byte    = 0x00
	elPrefix                    byte    = 0x01
	migrationBalanceBuffer      float64 = 0.01
)

type CheckSoloMigrations struct {
	ctx       context.Context
	sp        *services.ServiceProvider
	logger    *slog.Logger
	cfg       *config.SmartNodeConfig
	w         *wallet.Wallet
	rp        *rocketpool.RocketPool
	ec        eth.IExecutionClient
	bc        beacon.IBeaconClient
	mpMgr     *minipool.MinipoolManager
	coll      *collectors.SoloMigrationCollector
	lock      *sync.Mutex
	isRunning bool
}

// Create check solo migrations task
func NewCheckSoloMigrations(ctx context.Context, sp *services.ServiceProvider, logger *log.Logger, coll *collectors.SoloMigrationCollector) *CheckSoloMigrations {
	lock := &sync.Mutex{}
	return &CheckSoloMigrations{
		ctx:       ctx,
		sp:        sp,
		logger:    logger.With(slog.String(keys.RoutineKey, "Solo Migration")),
		cfg:       sp.GetConfig(),
		w:         sp.GetWallet(),
		rp:        sp.GetRocketPool(),
		ec:        sp.GetEthClient(),
		bc:        sp.GetBeaconClient(),
		coll:      coll,
		lock:      lock,
		isRunning: false,
	}
}

// Start the solo migration checking thread
func (t *CheckSoloMigrations) Run(state *state.NetworkState) error {
	// Log
	t.logger.Info("Starting solo migration check.")

	// Check if the check is already running
	t.lock.Lock()
	if t.isRunning {
		t.logger.Info("Solo migration check is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Run the check
	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()
		t.logger.Info("Starting solo migration check in a separate thread.")

		err := t.checkSoloMigrations(state)
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

// Check for solo staker migration validity
func (t *CheckSoloMigrations) checkSoloMigrations(state *state.NetworkState) error {
	// Update contract bindings
	var err error
	t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

	t.logger.Info(fmt.Sprintf("Checking solo migrations...", slog.Uint64(keys.SlotKey, state.BeaconSlotNumber), slog.Uint64(keys.BlockKey, state.ElBlockNumber)))
	oneGwei := eth.GweiToWei(1)
	scrubThreshold := time.Duration(state.NetworkDetails.PromotionScrubPeriod.Seconds()*soloMigrationCheckThreshold) * time.Second

	genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
	secondsForSlot := time.Duration(state.BeaconSlotNumber*state.BeaconConfig.SecondsPerSlot) * time.Second
	blockTime := genesisTime.Add(secondsForSlot)

	// Metrics
	totalCount := float64(0)
	doesntExistCount := float64(0)
	invalidStateCount := float64(0)
	timedOutCount := float64(0)
	invalidCredentialsCount := float64(0)
	balanceTooLowCount := float64(0)

	// Go through each minipool
	threshold := uint64(32000000000)
	buffer := uint64(migrationBalanceBuffer * eth.WeiPerGwei)
	for _, mpd := range state.MinipoolDetails {
		if mpd.Status == types.MinipoolStatus_Dissolved {
			// Ignore minipools that are already dissolved
			continue
		}

		if !mpd.IsVacant {
			// Ignore minipools that aren't vacant
			continue
		}

		totalCount += 1

		// Scrub minipools that aren't seen on Beacon yet
		validator := state.ValidatorDetails[mpd.Pubkey]
		if !validator.Exists {
			t.scrubVacantMinipool(state, mpd.MinipoolAddress, fmt.Sprintf("minipool %s (pubkey %s) did not exist on Beacon yet, but is required to be active_ongoing for migration", mpd.MinipoolAddress.Hex(), mpd.Pubkey.Hex()))
			doesntExistCount += 1
			continue
		}

		// Scrub minipools that are in the wrong state
		if validator.Status != beacon.ValidatorState_ActiveOngoing {
			t.scrubVacantMinipool(state, mpd.MinipoolAddress, fmt.Sprintf("minipool %s (pubkey %s) was in state %v, but is required to be active_ongoing for migration", mpd.MinipoolAddress.Hex(), mpd.Pubkey.Hex(), validator.Status))
			invalidStateCount += 1
			continue
		}

		// Check the withdrawal credentials
		withdrawalCreds := validator.WithdrawalCredentials
		switch withdrawalCreds[0] {
		case blsPrefix:
			creationTime := time.Unix(mpd.StatusTime.Int64(), 0)
			remainingTime := creationTime.Add(scrubThreshold).Sub(blockTime)
			if remainingTime < 0 {
				t.scrubVacantMinipool(state, mpd.MinipoolAddress, fmt.Sprintf("minipool timed out (created %s, current time %s, scrubbed after %s)", creationTime, blockTime, scrubThreshold))
				timedOutCount += 1
				continue
			}
			continue
		case elPrefix:
			if withdrawalCreds != mpd.WithdrawalCredentials {
				t.scrubVacantMinipool(state, mpd.MinipoolAddress, fmt.Sprintf("withdrawal credentials do not match (expected %s, actual %s)", mpd.WithdrawalCredentials.Hex(), withdrawalCreds.Hex()))
				invalidCredentialsCount += 1
				continue
			}
		default:
			t.scrubVacantMinipool(state, mpd.MinipoolAddress, fmt.Sprintf("unexpected prefix in withdrawal credentials: %s", withdrawalCreds.Hex()))
			invalidCredentialsCount += 1
			continue
		}

		// Check the balance
		creationBalanceGwei := big.NewInt(0).Div(mpd.PreMigrationBalance, oneGwei).Uint64()
		currentBalance := validator.Balance

		// Add the minipool balance to the Beacon balance in case it already got skimmed
		minipoolBalanceGwei := big.NewInt(0).Div(mpd.Balance, oneGwei).Uint64()
		currentBalance += minipoolBalanceGwei

		if currentBalance < threshold {
			t.scrubVacantMinipool(state, mpd.MinipoolAddress, fmt.Sprintf("current balance of %d is lower than the threshold of %d", currentBalance, threshold))
			balanceTooLowCount += 1
			continue
		}
		if currentBalance < (creationBalanceGwei - buffer) {
			t.scrubVacantMinipool(state, mpd.MinipoolAddress, fmt.Sprintf("current balance of %d is lower than the creation balance of %d, and below the acceptable buffer threshold of %d", currentBalance, creationBalanceGwei, buffer))
			balanceTooLowCount += 1
			continue
		}

	}

	// Update the metrics collector
	if t.coll != nil {
		t.coll.UpdateLock.Lock()
		defer t.coll.UpdateLock.Unlock()

		// Get the time of the state's EL block
		genesisTime := time.Unix(int64(state.BeaconConfig.GenesisTime), 0)
		secondsSinceGenesis := time.Duration(state.BeaconSlotNumber*state.BeaconConfig.SecondsPerSlot) * time.Second
		stateBlockTime := genesisTime.Add(secondsSinceGenesis)

		t.coll.LatestBlockTime = float64(stateBlockTime.Unix())
		t.coll.TotalMinipools = totalCount

		t.coll.DoesntExist = doesntExistCount
		t.coll.InvalidState = invalidStateCount
		t.coll.TimedOut = timedOutCount
		t.coll.InvalidCredentials = invalidCredentialsCount
		t.coll.BalanceTooLow = balanceTooLowCount
	}

	return nil
}

// Scrub a vacant minipool
func (t *CheckSoloMigrations) scrubVacantMinipool(state *state.NetworkState, address common.Address, reason string) {
	// Log
	mpLogger := t.logger.With(slog.String(keys.MinipoolKey, address.Hex()))
	mpLogger.Warn("=== SCRUBBING SOLO MIGRATION ===", slog.String(keys.ReasonKey, reason))

	// Make the binding
	mpd := state.MinipoolDetailsByAddress[address]
	mp, err := t.mpMgr.NewMinipoolFromVersion(address, mpd.Version)
	if err != nil {
		mpLogger.Error("Error creating minipool binding", log.Err(err))
		return
	}

	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		mpLogger.Error("Error getting node account transactor", log.Err(err))
		return
	}

	// Get the tx info
	txInfo, err := mp.Common().VoteScrub(opts)
	if err != nil {
		mpLogger.Error("Error getting Scrub TX for minipool", log.Err(err))
		return
	}
	if txInfo.SimulationResult.SimulationError != "" {
		mpLogger.Error("Simulating Scrub TX failed", slog.String(log.ErrorKey, txInfo.SimulationResult.SimulationError))
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
	err = tx.PrintAndWaitForTransaction(t.cfg, t.rp, mpLogger, txInfo, opts)
	if err != nil {
		mpLogger.Error("Error waiting for scrub transaction", log.Err(err))
		return
	}

	// Log
	mpLogger.Info("Successfully voted to scrub minipool.")
}

func (t *CheckSoloMigrations) handleError(err error) {
	t.logger.Error("*** Solo migration check failed. ***", log.Err(err))
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}
