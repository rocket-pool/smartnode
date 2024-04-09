package watchtower

import (
	"context"
	"fmt"
	"time"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/watchtower/collectors"
	"github.com/rocket-pool/smartnode/v2/shared/config"
)

// Config
const (
	minTasksInterval       time.Duration = time.Minute * 4
	maxTasksInterval       time.Duration = time.Minute * 6
	taskCooldown           time.Duration = time.Second * 5
	metricsShutdownTimeout time.Duration = time.Second * 5
)

type TaskManager struct {
	// Services
	logger *log.Logger
	ctx    context.Context
	sp     *services.ServiceProvider
	cfg    *config.SmartNodeConfig
	rp     *rocketpool.RocketPool
	bc     beacon.IBeaconClient

	// Tasks
	generateRewardsTree         *GenerateRewardsTree
	respondChallenges           *RespondChallenges
	submitRplPrice              *SubmitRplPrice
	submitNetworkBalances       *SubmitNetworkBalances
	dissolveTimedOutMinipools   *DissolveTimedOutMinipools
	submitScrubMinipools        *SubmitScrubMinipools
	submitRewardsTree_Stateless *SubmitRewardsTree_Stateless
	submitRewardsTree_Rolling   *SubmitRewardsTree_Rolling
	cancelBondReductions        *CancelBondReductions
	checkSoloMigrations         *CheckSoloMigrations
	finalizePdaoProposals       *FinalizePdaoProposals

	// Collectors
	ScrubCollector         *collectors.ScrubCollector
	BondReductionCollector *collectors.BondReductionCollector
	SoloMigrationCollector *collectors.SoloMigrationCollector

	// Internal
	useRollingRecords bool
	initialized       bool
}

func NewTaskManager(
	sp *services.ServiceProvider,
	stateMgr *state.NetworkStateManager,
	scrubCollector *collectors.ScrubCollector,
	bondReductionCollector *collectors.BondReductionCollector,
	soloMigrationCollector *collectors.SoloMigrationCollector,
) *TaskManager {
	logger := sp.GetWatchtowerLogger()
	ctx := logger.CreateContextWithLogger(sp.GetBaseContext())
	cfg := sp.GetConfig()
	rp := sp.GetRocketPool()
	bc := sp.GetBeaconClient()

	// Print the current mode
	if cfg.IsNativeMode {
		logger.Info("Starting watchtower daemon in Native Mode.")
	} else {
		logger.Info("Starting watchtower daemon in Docker Mode.")
	}

	// Check if rolling records are enabled
	useRollingRecords := cfg.UseRollingRecords.Value
	if useRollingRecords {
		logger.Info("Rolling records are enabled.")
	} else {
		logger.Info("Rolling records are disabled.")
	}

	// Initialize tasks
	generateRewardsTree := NewGenerateRewardsTree(ctx, sp, logger)
	respondChallenges := NewRespondChallenges(sp, logger, stateMgr)
	submitRplPrice := NewSubmitRplPrice(ctx, sp, logger)
	submitNetworkBalances := NewSubmitNetworkBalances(ctx, sp, logger)
	dissolveTimedOutMinipools := NewDissolveTimedOutMinipools(sp, logger)
	submitScrubMinipools := NewSubmitScrubMinipools(sp, logger, scrubCollector)
	var submitRewardsTree_Stateless *SubmitRewardsTree_Stateless
	var submitRewardsTree_Rolling *SubmitRewardsTree_Rolling
	/*processPenalties, err := newProcessPenalties(c, log.NewColorLogger(ProcessPenaltiesColor), errorLog)
	if err != nil {
		return fmt.Errorf("error during penalties check: %w", err)
	}*/
	cancelBondReductions := NewCancelBondReductions(ctx, sp, logger, bondReductionCollector)
	checkSoloMigrations := NewCheckSoloMigrations(ctx, sp, logger, soloMigrationCollector)
	finalizePdaoProposals := NewFinalizePdaoProposals(sp, logger)

	return &TaskManager{
		sp:                          sp,
		logger:                      logger,
		ctx:                         ctx,
		cfg:                         cfg,
		rp:                          rp,
		bc:                          bc,
		generateRewardsTree:         generateRewardsTree,
		respondChallenges:           respondChallenges,
		submitRplPrice:              submitRplPrice,
		submitNetworkBalances:       submitNetworkBalances,
		dissolveTimedOutMinipools:   dissolveTimedOutMinipools,
		submitScrubMinipools:        submitScrubMinipools,
		submitRewardsTree_Stateless: submitRewardsTree_Stateless,
		submitRewardsTree_Rolling:   submitRewardsTree_Rolling,
		cancelBondReductions:        cancelBondReductions,
		checkSoloMigrations:         checkSoloMigrations,
		finalizePdaoProposals:       finalizePdaoProposals,
		useRollingRecords:           useRollingRecords,
		ScrubCollector:              scrubCollector,
		BondReductionCollector:      bondReductionCollector,
		SoloMigrationCollector:      soloMigrationCollector,
	}
}

func (t *TaskManager) Initialize(stateMgr *state.NetworkStateManager) error {
	if t.initialized {
		return nil
	}

	var err error
	if !t.useRollingRecords {
		t.submitRewardsTree_Stateless = NewSubmitRewardsTree_Stateless(t.ctx, t.sp, t.logger, stateMgr)
	} else {
		t.submitRewardsTree_Rolling, err = NewSubmitRewardsTree_Rolling(t.ctx, t.sp, t.logger, stateMgr)
		if err != nil {
			return fmt.Errorf("error creating rolling rewards tree builder: %w", err)
		}
	}

	t.initialized = true
	return nil
}

// Run the task loop
func (t *TaskManager) Run(isOnOdao bool, state *state.NetworkState) error {
	// Run the manual rewards tree generation
	if err := t.generateRewardsTree.Run(); err != nil {
		t.logger.Error(err.Error())
	}
	if utils.SleepWithCancel(t.ctx, taskCooldown) {
		return nil
	}

	if isOnOdao {
		// Run the challenge check
		if err := t.respondChallenges.Run(); err != nil {
			t.logger.Error(err.Error())
		}
		if utils.SleepWithCancel(t.ctx, taskCooldown) {
			return nil
		}

		// Run the network balance submission check
		if err := t.submitNetworkBalances.Run(state); err != nil {
			t.logger.Error(err.Error())
		}
		if utils.SleepWithCancel(t.ctx, taskCooldown) {
			return nil
		}

		if !t.useRollingRecords {
			// Run the rewards tree submission check
			if err := t.submitRewardsTree_Stateless.Run(isOnOdao, state, state.BeaconSlotNumber); err != nil {
				t.logger.Error(err.Error())
			}
		} else {
			// Run the network balance and rewards tree submission check
			if err := t.submitRewardsTree_Rolling.Run(state); err != nil {
				t.logger.Error(err.Error())
			}
		}
		if utils.SleepWithCancel(t.ctx, taskCooldown) {
			return nil
		}

		// Run the price submission check
		if err := t.submitRplPrice.Run(state); err != nil {
			t.logger.Error(err.Error())
		}
		if utils.SleepWithCancel(t.ctx, taskCooldown) {
			return nil
		}

		// Run the minipool dissolve check
		if err := t.dissolveTimedOutMinipools.Run(state); err != nil {
			t.logger.Error(err.Error())
		}
		if utils.SleepWithCancel(t.ctx, taskCooldown) {
			return nil
		}

		// Run the finalize proposals check
		if err := t.finalizePdaoProposals.Run(state); err != nil {
			t.logger.Error(err.Error())
		}
		if utils.SleepWithCancel(t.ctx, taskCooldown) {
			return nil
		}

		// Run the minipool scrub check
		if err := t.submitScrubMinipools.Run(state); err != nil {
			t.logger.Error(err.Error())
		}
		if utils.SleepWithCancel(t.ctx, taskCooldown) {
			return nil
		}

		// Run the bond cancel check
		if err := t.cancelBondReductions.Run(state); err != nil {
			t.logger.Error(err.Error())
		}
		if utils.SleepWithCancel(t.ctx, taskCooldown) {
			return nil
		}

		// Run the solo migration check
		if err := t.checkSoloMigrations.Run(state); err != nil {
			t.logger.Error(err.Error())
		}
	} else {
		/*
		 */
		if !t.useRollingRecords {
			// Run the rewards tree submission check
			if err := t.submitRewardsTree_Stateless.Run(isOnOdao, nil, state.BeaconSlotNumber); err != nil {
				t.logger.Error(err.Error())
			}
		} else {
			// Run the network balance and rewards tree submission check
			if err := t.submitRewardsTree_Rolling.Run(nil); err != nil {
				t.logger.Error(err.Error())
			}
		}
	}

	return nil
}
