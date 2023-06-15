package watchtower

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/utils/log"
	"github.com/urfave/cli"
)

// Process balances and rewards task
type processBalancesAndRewards struct {
	c         *cli.Context
	log       log.ColorLogger
	errLog    log.ColorLogger
	cfg       *config.RocketPoolConfig
	w         *wallet.Wallet
	ec        rocketpool.ExecutionClient
	rp        *rocketpool.RocketPool
	bc        beacon.Client
	recordMgr *RollingRecordManager
}

// Create process balances and rewards task
func newProcessBalancesAndRewards(c *cli.Context, logger log.ColorLogger, errorLogger log.ColorLogger, stateMgr *state.NetworkStateManager) (*processBalancesAndRewards, error) {

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
	task := &processBalancesAndRewards{
		c:      c,
		log:    logger,
		errLog: errorLogger,
		cfg:    cfg,
		ec:     ec,
		w:      w,
		rp:     rp,
		bc:     bc,
	}

	// Get the beacon config
	beaconCfg, err := bc.GetEth2Config()
	if err != nil {
		return nil, fmt.Errorf("error getting beacon config: %w", err)
	}

	// Get the current interval index
	currentIndexBig, err := rewards.GetRewardIndex(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting rewards index: %w", err)
	}
	currentIndex := currentIndexBig.Uint64()

	// Get the previous RocketRewardsPool addresses
	prevAddresses := cfg.Smartnode.GetPreviousRewardsPoolAddresses()

	// Get the last rewards event and starting epoch
	startSlot := uint64(0)
	found, event, err := rewards.GetRewardsEvent(rp, currentIndex-1, prevAddresses, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting event for rewards interval %d: %w", currentIndex-1, err)
	}
	if !found {
		logger.Printlnf("NOTE: event for previous rewards interval %d not found. Starting from slot zero.", currentIndex-1)
	} else {
		// Get the start slot of the current interval
		previousEpoch := event.ConsensusBlock.Uint64() / beaconCfg.SlotsPerEpoch
		newEpoch := previousEpoch + 1
		startSlot = newEpoch * beaconCfg.SlotsPerEpoch
	}

	// Make a new rolling manager
	mgr, err := NewRollingRecordManager(&task.log, &task.errLog, cfg, rp, bc, stateMgr, w, startSlot, beaconCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating rolling record manager: %w", err)
	}

	// Return
	task.recordMgr = mgr
	return task, nil

}

// Process balances and rewards submissions
func (t *processBalancesAndRewards) run(isOnOdao bool, state *state.NetworkState) error {
	return t.recordMgr.ProcessNewHeadState(state)
}
