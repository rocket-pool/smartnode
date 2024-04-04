package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/collectors"
)

// Config
const (
	minTasksInterval       time.Duration = time.Minute * 4
	maxTasksInterval       time.Duration = time.Minute * 6
	taskCooldown           time.Duration = time.Second * 5
	metricsShutdownTimeout time.Duration = time.Second * 5
)

type TaskLoop struct {
	logger        *log.Logger
	ctx           context.Context
	sp            *services.ServiceProvider
	wg            *sync.WaitGroup
	metricsServer *http.Server
}

func NewTaskLoop(sp *services.ServiceProvider, wg *sync.WaitGroup) *TaskLoop {
	logger := sp.GetWatchtowerLogger()
	return &TaskLoop{
		sp:     sp,
		logger: logger,
		ctx:    logger.CreateContextWithLogger(sp.GetBaseContext()),
		wg:     wg,
	}
}

// Run daemon
func (t *TaskLoop) Run() error {
	// Get services
	cfg := t.sp.GetConfig()
	rp := t.sp.GetRocketPool()
	bc := t.sp.GetBeaconClient()

	// Wait until node is registered
	if err := t.sp.WaitNodeRegistered(t.ctx, true); err != nil {
		return err
	}

	// Print the current mode
	if cfg.IsNativeMode {
		fmt.Println("Starting watchtower daemon in Native Mode.")
	} else {
		fmt.Println("Starting watchtower daemon in Docker Mode.")
	}

	// Check if rolling records are enabled
	useRollingRecords := cfg.UseRollingRecords.Value
	if useRollingRecords {
		fmt.Println("Rolling records are enabled.")
	} else {
		fmt.Println("Rolling records are disabled.")
	}

	// Initialize the metrics reporters
	scrubCollector := collectors.NewScrubCollector()
	bondReductionCollector := collectors.NewBondReductionCollector()
	soloMigrationCollector := collectors.NewSoloMigrationCollector()

	// Create the state manager
	m, err := state.NewNetworkStateManager(t.ctx, rp, cfg, rp.Client, bc, t.logger.Logger)
	if err != nil {
		return err
	}

	// Initialize tasks
	respondChallenges := NewRespondChallenges(t.sp, t.logger, m)
	submitRplPrice := NewSubmitRplPrice(t.ctx, t.sp, t.logger)
	submitNetworkBalances := NewSubmitNetworkBalances(t.ctx, t.sp, t.logger)
	dissolveTimedOutMinipools := NewDissolveTimedOutMinipools(t.sp, t.logger)
	submitScrubMinipools := NewSubmitScrubMinipools(t.sp, t.logger, scrubCollector)
	var submitRewardsTree_Stateless *SubmitRewardsTree_Stateless
	var submitRewardsTree_Rolling *SubmitRewardsTree_Rolling
	if !useRollingRecords {
		submitRewardsTree_Stateless = NewSubmitRewardsTree_Stateless(t.ctx, t.sp, t.logger, m)
	} else {
		submitRewardsTree_Rolling, err = NewSubmitRewardsTree_Rolling(t.ctx, t.sp, t.logger, m)
		if err != nil {
			return fmt.Errorf("error during rolling rewards tree check: %w", err)
		}
	}
	/*processPenalties, err := newProcessPenalties(c, log.NewColorLogger(ProcessPenaltiesColor), errorLog)
	if err != nil {
		return fmt.Errorf("error during penalties check: %w", err)
	}*/
	generateRewardsTree := NewGenerateRewardsTree(t.ctx, t.sp, t.logger)
	cancelBondReductions := NewCancelBondReductions(t.ctx, t.sp, t.logger, bondReductionCollector)
	checkSoloMigrations := NewCheckSoloMigrations(t.ctx, t.sp, t.logger, soloMigrationCollector)
	finalizePdaoProposals := NewFinalizePdaoProposals(t.sp, t.logger)

	intervalDelta := maxTasksInterval - minTasksInterval
	secondsDelta := intervalDelta.Seconds()

	// Run task loop
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()

		for {
			// Randomize the next interval
			randomSeconds := rand.Intn(int(secondsDelta))
			interval := time.Duration(randomSeconds)*time.Second + minTasksInterval

			// Check the EC status
			err := t.sp.WaitEthClientSynced(t.ctx, false) // Force refresh the primary / fallback EC status
			if err != nil {
				t.logger.Error(err.Error())
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			// Check the BC status
			err = t.sp.WaitBeaconClientSynced(t.ctx, false) // Force refresh the primary / fallback BC status
			if err != nil {
				t.logger.Error(err.Error())
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			// Load contracts
			err = t.sp.RefreshRocketPoolContracts()
			if err != nil {
				t.logger.Error("error loading contract bindings", log.Err(err))
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			// Get the Beacon block
			//latestBlock, err := m.GetLatestFinalizedBeaconBlock()
			latestBlock, err := m.GetLatestBeaconBlock(t.ctx)
			if err != nil {
				t.logger.Error("error getting latest Beacon block", log.Err(err))
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			nodeAddress, hasNodeAddress := t.sp.GetWallet().GetAddress()
			if !hasNodeAddress {
				continue
			}

			// Check if on the Oracle DAO
			isOnOdao, err := isOnOracleDAO(rp, nodeAddress, latestBlock)
			if err != nil {
				t.logger.Error(err.Error())
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			// Run the manual rewards tree generation
			if err := generateRewardsTree.Run(); err != nil {
				t.logger.Error(err.Error())
			}
			if utils.SleepWithCancel(t.ctx, taskCooldown) {
				break
			}

			if isOnOdao {
				// Run the challenge check
				if err := respondChallenges.Run(); err != nil {
					t.logger.Error(err.Error())
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Update the network state
				state, err := updateNetworkState(t.ctx, m, t.logger, latestBlock)
				if err != nil {
					t.logger.Error(err.Error())
					if utils.SleepWithCancel(t.ctx, taskCooldown) {
						break
					}
					continue
				}

				// Run the network balance submission check
				if err := submitNetworkBalances.Run(state); err != nil {
					t.logger.Error(err.Error())
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				if !useRollingRecords {
					// Run the rewards tree submission check
					if err := submitRewardsTree_Stateless.Run(isOnOdao, state, latestBlock.Header.Slot); err != nil {
						t.logger.Error(err.Error())
					}
					if utils.SleepWithCancel(t.ctx, taskCooldown) {
						break
					}
				} else {
					// Run the network balance and rewards tree submission check
					if err := submitRewardsTree_Rolling.Run(state); err != nil {
						t.logger.Error(err.Error())
					}
					if utils.SleepWithCancel(t.ctx, taskCooldown) {
						break
					}
				}

				// Run the price submission check
				if err := submitRplPrice.Run(state); err != nil {
					t.logger.Error(err.Error())
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the minipool dissolve check
				if err := dissolveTimedOutMinipools.Run(state); err != nil {
					t.logger.Error(err.Error())
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the finalize proposals check
				if err := finalizePdaoProposals.Run(state); err != nil {
					t.logger.Error(err.Error())
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the minipool scrub check
				if err := submitScrubMinipools.Run(state); err != nil {
					t.logger.Error(err.Error())
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the bond cancel check
				if err := cancelBondReductions.Run(state); err != nil {
					t.logger.Error(err.Error())
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the solo migration check
				if err := checkSoloMigrations.Run(state); err != nil {
					t.logger.Error(err.Error())
				}
				/*time.Sleep(taskCooldown)

				// Run the fee recipient penalty check
				if err := processPenalties.run(); err != nil {
					errorLog.Println(err)
				}*/
				// DISABLED until MEV-Boost can support it
			} else {
				/*
				 */
				if !useRollingRecords {
					// Run the rewards tree submission check
					if err := submitRewardsTree_Stateless.Run(isOnOdao, nil, latestBlock.Header.Slot); err != nil {
						t.logger.Error(err.Error())
					}
				} else {
					// Run the network balance and rewards tree submission check
					if err := submitRewardsTree_Rolling.Run(nil); err != nil {
						t.logger.Error(err.Error())
					}
				}
			}

			if utils.SleepWithCancel(t.ctx, interval) {
				break
			}
		}
	}()

	// Run metrics loop
	t.metricsServer = runMetricsServer(t.sp, t.logger, scrubCollector, bondReductionCollector, soloMigrationCollector, t.wg)

	return nil
}

func (t *TaskLoop) Stop() {
	if t.metricsServer != nil {
		// Shut down the metrics server
		ctx, cancel := context.WithTimeout(context.Background(), metricsShutdownTimeout)
		defer cancel()
		t.metricsServer.Shutdown(ctx)
	}
}

// Update the latest network state at each cycle
func updateNetworkState(ctx context.Context, m *state.NetworkStateManager, logger *log.Logger, block beacon.BeaconBlock) (*state.NetworkState, error) {
	logger.Info("Getting latest network state... ")
	// Get the state of the network
	state, err := m.GetStateForSlot(ctx, block.Header.Slot)
	if err != nil {
		return nil, fmt.Errorf("error getting network state: %w", err)
	}
	return state, nil
}

// Check if this node is on the Oracle DAO
func isOnOracleDAO(rp *rocketpool.RocketPool, nodeAddress common.Address, block beacon.BeaconBlock) (bool, error) {
	opts := &bind.CallOpts{
		BlockNumber: big.NewInt(0).SetUint64(block.ExecutionBlockNumber),
	}

	member, err := oracle.NewOracleDaoMember(rp, nodeAddress)
	if err != nil {
		return false, fmt.Errorf("error creating oDAO member binding: %w", err)
	}
	err = rp.Query(nil, opts, member.Exists)
	if err != nil {
		return false, fmt.Errorf("error checking if node is in the Oracle DAO for Beacon block %d, EL block %d: %w", block.Header.Slot, block.ExecutionBlockNumber, err)
	}
	return member.Exists.Get(), nil
}
