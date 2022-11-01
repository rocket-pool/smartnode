package watchtower

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/api"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/rocket-pool/smartnode/shared/utils/rp"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

const (
	soloMigrationCheckThreshold time.Duration = 24 * time.Hour
	blsPrefix                   byte          = 0x00
	elPrefix                    byte          = 0x01
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
	lock             *sync.Mutex
	isRunning        bool
	generationPrefix string
	isAtlasDeployed  bool
}

// Create check solo migrations task
func newCheckSoloMigrations(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger) (*checkSoloMigrations, error) {

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
		lock:             lock,
		isRunning:        false,
		generationPrefix: "[Solo Migration]",
		isAtlasDeployed:  false,
	}, nil

}

// Start the solo migration checking thread
func (t *checkSoloMigrations) run() error {

	// Wait for eth clients to sync
	if err := services.WaitEthClientSynced(t.c, true); err != nil {
		return err
	}
	if err := services.WaitBeaconClientSynced(t.c, true); err != nil {
		return err
	}

	// Get node account
	nodeAccount, err := t.w.GetNodeAccount()
	if err != nil {
		return err
	}

	// Get trusted node status
	nodeTrusted, err := trustednode.GetMemberExists(t.rp, nodeAccount.Address, nil)
	if err != nil {
		return err
	}
	if !(nodeTrusted) {
		return nil
	}

	// Check if Atlas is deployed
	if !t.isAtlasDeployed {
		isAtlasDeployed, err := rp.IsAtlasDeployed(t.rp)
		if err != nil {
			return fmt.Errorf("error checking if Atlas is deployed: %w", err)
		}
		t.isAtlasDeployed = isAtlasDeployed
		if !t.isAtlasDeployed {
			return nil
		}
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

		err := t.checkSoloMigrations()
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
func (t *checkSoloMigrations) checkSoloMigrations() error {

	// Data
	var wg1 errgroup.Group
	var eth2Config beacon.Eth2Config
	var beaconHead beacon.BeaconHead

	// Get eth2 config
	wg1.Go(func() error {
		var err error
		eth2Config, err = t.bc.GetEth2Config()
		if err != nil {
			return fmt.Errorf("error getting Beacon config: %w", err)
		}
		return nil
	})

	// Get beacon head
	wg1.Go(func() error {
		var err error
		beaconHead, err = t.bc.GetBeaconHead()
		if err != nil {
			return fmt.Errorf("error getting Beacon head: %w", err)
		}
		return nil
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return err
	}

	// Get the latest finalized slot that exists, and the EL block for it
	finalizedEpoch := beaconHead.FinalizedEpoch
	lastSlot := (finalizedEpoch+1)*eth2Config.SlotsPerEpoch - 1
	var elBlock uint64
	for lastSlot > 0 {
		block, exists, err := t.bc.GetBeaconBlock(fmt.Sprint(lastSlot))
		if err != nil {
			return fmt.Errorf("error getting Beacon block %d: %w", lastSlot, err)
		}
		if !exists {
			lastSlot--
		} else {
			elBlock = block.ExecutionBlockNumber
			break
		}
	}

	t.printMessage(fmt.Sprintf("Latest finalized epoch is %d, checking for Beacon slot %d (EL block %d)", finalizedEpoch, lastSlot, elBlock))
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(elBlock),
	}
	beaconOpts := &beacon.ValidatorStatusOptions{
		Slot: &lastSlot,
	}

	// Get vacant count
	vacantCount, err := minipool.GetVacantMinipoolCount(t.rp, opts)
	if err != nil {
		return fmt.Errorf("error getting vacant minipool count: %w", err)
	}

	if vacantCount == 0 {
		return nil
	}

	// Go through each minipool
	// TODO: does this need to be multithreaded?
	for i := uint64(0); i < vacantCount; i++ {
		address, err := minipool.GetVacantMinipoolAt(t.rp, i, opts)
		if err != nil {
			return fmt.Errorf("error getting vacant minipool %d address: %w", i, err)
		}

		pubkey, err := minipool.GetMinipoolPubkey(t.rp, address, opts)
		if err != nil {
			return fmt.Errorf("error getting minipool %s pubkey: %w", address.Hex(), err)
		}

		status, err := t.bc.GetValidatorStatus(pubkey, beaconOpts)
		if err != nil {
			return fmt.Errorf("error getting minipool %s Beacon status: %w", address.Hex(), err)
		}

		withdrawalCreds := status.WithdrawalCredentials
		switch withdrawalCreds[0] {
		case blsPrefix:
			// Hasn't migrated yet, so ignore for now
			// TODO: Handle timeouts once they're added
			continue
		case elPrefix:
			expectedCreds, err := minipool.GetMinipoolWithdrawalCredentials(t.rp, address, opts)
			if err != nil {
				return fmt.Errorf("error getting expected withdrawal credentials for minipool %s: %w", address.Hex(), err)
			}
			if withdrawalCreds != expectedCreds {
				t.scrubVacantMinipool(address, fmt.Sprintf("withdrawal credentials do not match (expected %s, actual %s)", expectedCreds.Hex(), withdrawalCreds.Hex()))
			}
		default:
			t.scrubVacantMinipool(address, fmt.Sprintf("unexpected prefix in withdrawal credentials: %s", withdrawalCreds.Hex()))
		}
	}

	return nil

}

// Scrub a vacant minipool
func (t *checkSoloMigrations) scrubVacantMinipool(address common.Address, reason string) error {

	// Log
	t.printMessage("=== SCRUBBING SOLO MIGRATION ===")
	t.printMessage(fmt.Sprintf("Minipool: %s", address.Hex()))
	t.printMessage(fmt.Sprintf("Reason:   %s", reason))
	t.printMessage("================================")

	// Make the binding
	mp, err := minipool.NewMinipool(t.rp, address)
	if err != nil {
		return fmt.Errorf("error scrubbing migration of minipool %s: %w", address.Hex(), err)
	}

	// Get transactor
	opts, err := t.w.GetNodeAccountTransactor()
	if err != nil {
		return err
	}

	// Get the gas limit
	gasInfo, err := mp.EstimateVoteScrubGas(opts)
	if err != nil {
		return fmt.Errorf("could not estimate the gas required to scrub the minipool: %w", err)
	}

	// Print the gas info
	maxFee := eth.GweiToWei(WatchtowerMaxFee)
	if !api.PrintAndCheckGasInfo(gasInfo, false, 0, t.log, maxFee, 0) {
		return nil
	}

	// Set the gas settings
	opts.GasFeeCap = maxFee
	opts.GasTipCap = eth.GweiToWei(WatchtowerMaxPriorityFee)
	opts.GasLimit = gasInfo.SafeGasLimit

	// Cancel the reduction
	hash, err := mp.VoteScrub(opts)
	if err != nil {
		return err
	}

	// Print TX info and wait for it to be included in a block
	err = api.PrintAndWaitForTransaction(t.cfg, hash, t.rp.Client, t.log)
	if err != nil {
		return err
	}

	// Log
	t.log.Printlnf("Successfully voted to scrub minipool %s.", mp.Address.Hex())

	// Return
	return nil

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
