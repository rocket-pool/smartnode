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
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/gas"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/tx"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/utils"
	"github.com/rocket-pool/smartnode/shared/config"
)

const (
	soloMigrationCheckThreshold float64 = 0.85 // Fraction of PromotionStakePeriod that can go before a minipool gets scrubbed for not having changed to 0x01
	blsPrefix                   byte    = 0x00
	elPrefix                    byte    = 0x01
	migrationBalanceBuffer      float64 = 0.01
)

type CheckSoloMigrations struct {
	sp               *services.ServiceProvider
	log              log.ColorLogger
	errLog           log.ColorLogger
	cfg              *config.SmartNodeConfig
	w                *wallet.Wallet
	rp               *rocketpool.RocketPool
	ec               eth.IExecutionClient
	bc               beacon.IBeaconClient
	mpMgr            *minipool.MinipoolManager
	coll             *collectors.SoloMigrationCollector
	lock             *sync.Mutex
	isRunning        bool
	generationPrefix string
}

// Create check solo migrations task
func NewCheckSoloMigrations(sp *services.ServiceProvider, logger log.ColorLogger, errorLogger log.ColorLogger, coll *collectors.SoloMigrationCollector) *CheckSoloMigrations {
	lock := &sync.Mutex{}
	return &CheckSoloMigrations{
		sp:               sp,
		log:              logger,
		errLog:           errorLogger,
		cfg:              sp.GetConfig(),
		w:                sp.GetWallet(),
		rp:               sp.GetRocketPool(),
		ec:               sp.GetEthClient(),
		bc:               sp.GetBeaconClient(),
		coll:             coll,
		lock:             lock,
		isRunning:        false,
		generationPrefix: "[Solo Migration]",
	}

}

// Start the solo migration checking thread
func (t *CheckSoloMigrations) Run(state *state.NetworkState) error {
	// Log
	t.log.Println("Checking for solo migrations...")

	// Check if the check is already running
	t.lock.Lock()
	if t.isRunning {
		t.log.Println("Solo migration check is already running in the background.")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()

	// Run the check
	go func() {
		t.lock.Lock()
		t.isRunning = true
		t.lock.Unlock()
		t.printMessage("Starting solo migration check in a separate thread.")

		err := t.checkSoloMigrations(state)
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

// Check for solo staker migration validity
func (t *CheckSoloMigrations) checkSoloMigrations(state *state.NetworkState) error {
	// Update contract bindings
	var err error
	t.mpMgr, err = minipool.NewMinipoolManager(t.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool manager: %w", err)
	}

	t.printMessage(fmt.Sprintf("Checking for Beacon slot %d (EL block %d)", state.BeaconSlotNumber, state.ElBlockNumber))
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
	t.printMessage("=== SCRUBBING SOLO MIGRATION ===")
	t.printMessage(fmt.Sprintf("Minipool: %s", address.Hex()))
	t.printMessage(fmt.Sprintf("Reason:   %s", reason))
	t.printMessage("================================")

	// Make the binding
	mpd := state.MinipoolDetailsByAddress[address]
	mp, err := t.mpMgr.NewMinipoolFromVersion(address, mpd.Version)
	if err != nil {
		t.printMessage(fmt.Sprintf("error creating binding for minipool %s: %s", address.Hex(), err.Error()))
		return
	}
	// Get transactor
	opts, err := t.w.GetTransactor()
	if err != nil {
		t.printMessage(fmt.Sprintf("error getting node account transactor: %s", err.Error()))
		return
	}

	// Get the tx info
	txInfo, err := mp.Common().VoteScrub(opts)
	if err != nil {
		t.printMessage(fmt.Sprintf("error getting scrub tx for minipool: %s", err.Error()))
		return
	}
	if txInfo.SimulationResult.SimulationError != "" {
		t.printMessage(fmt.Sprintf("simulating scrub TX failed: %s", txInfo.SimulationResult.SimulationError))
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
		t.printMessage(fmt.Sprintf("error waiting for scrub transaction: %s", err.Error()))
		return
	}

	// Log
	t.log.Printlnf("Successfully voted to scrub minipool %s.", address.Hex())
}

func (t *CheckSoloMigrations) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Solo migration check failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Print a message from the tree generation goroutine
func (t *CheckSoloMigrations) printMessage(message string) {
	t.log.Printlnf("%s %s", t.generationPrefix, message)
}
