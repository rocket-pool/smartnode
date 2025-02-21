package watchtower

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/rocket-pool/smartnode/rocketpool/watchtower/collectors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
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
	soloMigrationCheckThreshold float64 = 0.85 // Fraction of PromotionStakePeriod that can go before a minipool gets scrubbed for not having changed to 0x01
	blsPrefix                   byte    = 0x00
	elPrefix                    byte    = 0x01
	migrationBalanceBuffer      float64 = 0.01
)

type checkSoloMigrations struct {
	c                *cli.Context
	log              log.ColorLogger
	errLog           log.ColorLogger
	cfg              *config.RocketPoolConfig
	w                *wallet.Wallet
	rp               *rocketpool.RocketPool
	ec               rocketpool.ExecutionClient
	bc               beacon.Client
	coll             *collectors.SoloMigrationCollector
	lock             *sync.Mutex
	isRunning        bool
	generationPrefix string
}

// Create check solo migrations task
func newCheckSoloMigrations(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger, coll *collectors.SoloMigrationCollector) (*checkSoloMigrations, error) {

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Return task
	lock := &sync.Mutex{}
	return &checkSoloMigrations{
		c:                c,
		log:              logger,
		errLog:           errorLogger,
		cfg:              cfg,
		w:                w,
		rp:               rp,
		ec:               ec,
		bc:               bc,
		coll:             coll,
		lock:             lock,
		isRunning:        false,
		generationPrefix: "[Solo Migration]",
	}, nil

}

// Start the solo migration checking thread
func (t *checkSoloMigrations) run(state *state.NetworkState) error {

	// Wait for eth clients to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
		return err
	}

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
func (t *checkSoloMigrations) checkSoloMigrations(state *state.NetworkState) error {

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
		if mpd.Status == types.Dissolved {
			// Ignore minipools that are already dissolved
			continue
		}

		if !mpd.IsVacant {
			// Ignore minipools that aren't vacant
			continue
		}

		totalCount += 1

		// Scrub minipools that aren't seen on Beacon yet
		validator := state.MinipoolValidatorDetails[mpd.Pubkey]
		if !validator.Exists {
			t.scrubVacantMinipool(mpd.MinipoolAddress, fmt.Sprintf("minipool %s (pubkey %s) did not exist on Beacon yet, but is required to be active_ongoing for migration", mpd.MinipoolAddress.Hex(), mpd.Pubkey.Hex()))
			doesntExistCount += 1
			continue
		}

		// Scrub minipools that are in the wrong state
		if validator.Status != beacon.ValidatorState_ActiveOngoing {
			t.scrubVacantMinipool(mpd.MinipoolAddress, fmt.Sprintf("minipool %s (pubkey %s) was in state %v, but is required to be active_ongoing for migration", mpd.MinipoolAddress.Hex(), mpd.Pubkey.Hex(), validator.Status))
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
				t.scrubVacantMinipool(mpd.MinipoolAddress, fmt.Sprintf("minipool timed out (created %s, current time %s, scrubbed after %s)", creationTime, blockTime, scrubThreshold))
				timedOutCount += 1
				continue
			}
			continue
		case elPrefix:
			if withdrawalCreds != mpd.WithdrawalCredentials {
				t.scrubVacantMinipool(mpd.MinipoolAddress, fmt.Sprintf("withdrawal credentials do not match (expected %s, actual %s)", mpd.WithdrawalCredentials.Hex(), withdrawalCreds.Hex()))
				invalidCredentialsCount += 1
				continue
			}
		default:
			t.scrubVacantMinipool(mpd.MinipoolAddress, fmt.Sprintf("unexpected prefix in withdrawal credentials: %s", withdrawalCreds.Hex()))
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
			t.scrubVacantMinipool(mpd.MinipoolAddress, fmt.Sprintf("current balance of %d is lower than the threshold of %d", currentBalance, threshold))
			balanceTooLowCount += 1
			continue
		}
		if currentBalance < (creationBalanceGwei - buffer) {
			t.scrubVacantMinipool(mpd.MinipoolAddress, fmt.Sprintf("current balance of %d is lower than the creation balance of %d, and below the acceptable buffer threshold of %d", currentBalance, creationBalanceGwei, buffer))
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
func (t *checkSoloMigrations) scrubVacantMinipool(address common.Address, reason string) {

	// Log
	t.printMessage("=== SCRUBBING SOLO MIGRATION ===")
	t.printMessage(fmt.Sprintf("Minipool: %s", address.Hex()))
	t.printMessage(fmt.Sprintf("Reason:   %s", reason))
	t.printMessage("================================")

	// Make the binding
	mp, err := minipool.NewMinipool(t.rp, address, nil)
	if err != nil {
		t.printMessage(fmt.Sprintf("error scrubbing migration of minipool %s: %s", address.Hex(), err.Error()))
		return
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		t.printMessage(fmt.Sprintf("error getting node account transactor: %s", err.Error()))
		return
	}

	// Get the gas limit
	gasInfo, err := mp.EstimateVoteScrubGas(opts)
	if err != nil {
		t.printMessage(fmt.Sprintf("could not estimate the gas required to scrub the minipool: %s", err.Error()))
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
	hash, err := mp.VoteScrub(opts)
	if err != nil {
		t.printMessage(fmt.Sprintf("could not vote to scrub the minipool: %s", err.Error()))
		return
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, &t.log)
	if err != nil {
		t.printMessage(fmt.Sprintf("error waiting for scrub transaction: %s", err.Error()))
		return
	}

	// Log
	t.log.Printlnf("Successfully voted to scrub minipool %s.", mp.GetAddress().Hex())

}

func (t *checkSoloMigrations) handleError(err error) {
	t.errLog.Println(err)
	t.errLog.Println("*** Solo migration check failed. ***")
	t.lock.Lock()
	t.isRunning = false
	t.lock.Unlock()
}

// Print a message from the tree generation goroutine
func (t *checkSoloMigrations) printMessage(message string) {
	t.log.Printlnf("%s %s", t.generationPrefix, message)
}
