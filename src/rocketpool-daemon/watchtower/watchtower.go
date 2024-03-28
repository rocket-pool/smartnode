package watchtower

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"

	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/rocket-pool/node-manager-core/utils/log"
	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/watchtower/collectors"
)

// Config
const (
	minTasksInterval time.Duration = time.Minute * 4
	maxTasksInterval time.Duration = time.Minute * 6
	taskCooldown     time.Duration = time.Second * 5
)

const (
	MaxConcurrentEth1Requests = 200

	RespondChallengesColor         = color.FgWhite
	ClaimRplRewardsColor           = color.FgGreen
	SubmitRplPriceColor            = color.FgYellow
	SubmitNetworkBalancesColor     = color.FgYellow
	DissolveTimedOutMinipoolsColor = color.FgMagenta
	SubmitScrubMinipoolsColor      = color.FgHiGreen
	ErrorColor                     = color.FgRed
	MetricsColor                   = color.FgHiYellow
	SubmitRewardsTreeColor         = color.FgHiCyan
	WarningColor                   = color.FgYellow
	ProcessPenaltiesColor          = color.FgHiMagenta
	CancelBondsColor               = color.FgGreen
	CheckSoloMigrationsColor       = color.FgCyan
	FinalizeProposalsColor         = color.FgMagenta
	UpdateColor                    = color.FgHiWhite
)

type TaskLoop struct {
	ctx context.Context
	sp  *services.ServiceProvider
	wg  *sync.WaitGroup
}

func NewTaskLoop(sp *services.ServiceProvider, wg *sync.WaitGroup) *TaskLoop {
	return &TaskLoop{
		sp:  sp,
		ctx: sp.GetContext(),
		wg:  wg,
	}
}

// Run daemon
func (t *TaskLoop) Run() error {
	// Get services
	cfg := t.sp.GetConfig()
	rp := t.sp.GetRocketPool()
	bc := t.sp.GetBeaconClient()

	// Wait until node is registered
	if err := t.sp.WaitNodeRegistered(true); err != nil {
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

	// Initialize error logger
	errorLog := log.NewColorLogger(ErrorColor)
	updateLog := log.NewColorLogger(UpdateColor)

	// Create the state manager
	m, err := state.NewNetworkStateManager(t.ctx, rp, cfg, rp.Client, bc, &updateLog)
	if err != nil {
		return err
	}

	// Initialize tasks
	respondChallenges := NewRespondChallenges(t.sp, log.NewColorLogger(RespondChallengesColor), m)
	submitRplPrice := NewSubmitRplPrice(t.ctx, t.sp, log.NewColorLogger(SubmitRplPriceColor), errorLog)
	submitNetworkBalances := NewSubmitNetworkBalances(t.ctx, t.sp, log.NewColorLogger(SubmitNetworkBalancesColor), errorLog)
	dissolveTimedOutMinipools := NewDissolveTimedOutMinipools(t.sp, log.NewColorLogger(DissolveTimedOutMinipoolsColor))
	submitScrubMinipools := NewSubmitScrubMinipools(t.sp, log.NewColorLogger(SubmitScrubMinipoolsColor), errorLog, scrubCollector)
	var submitRewardsTree_Stateless *SubmitRewardsTree_Stateless
	var submitRewardsTree_Rolling *SubmitRewardsTree_Rolling
	if !useRollingRecords {
		submitRewardsTree_Stateless = NewSubmitRewardsTree_Stateless(t.ctx, t.sp, log.NewColorLogger(SubmitRewardsTreeColor), errorLog, m)
	} else {
		submitRewardsTree_Rolling, err = NewSubmitRewardsTree_Rolling(t.ctx, t.sp, log.NewColorLogger(SubmitRewardsTreeColor), errorLog, m)
		if err != nil {
			return fmt.Errorf("error during rolling rewards tree check: %w", err)
		}
	}
	/*processPenalties, err := newProcessPenalties(c, log.NewColorLogger(ProcessPenaltiesColor), errorLog)
	if err != nil {
		return fmt.Errorf("error during penalties check: %w", err)
	}*/
	generateRewardsTree := NewGenerateRewardsTree(t.ctx, t.sp, log.NewColorLogger(SubmitRewardsTreeColor), errorLog)
	cancelBondReductions := NewCancelBondReductions(t.sp, log.NewColorLogger(CancelBondsColor), errorLog, bondReductionCollector)
	checkSoloMigrations := NewCheckSoloMigrations(t.sp, log.NewColorLogger(CheckSoloMigrationsColor), errorLog, soloMigrationCollector)
	finalizePdaoProposals := NewFinalizePdaoProposals(t.sp, log.NewColorLogger(FinalizeProposalsColor))

	intervalDelta := maxTasksInterval - minTasksInterval
	secondsDelta := intervalDelta.Seconds()

	// Wait group to handle the various threads
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// Run task loop
	go func() {
		for {
			// Randomize the next interval
			randomSeconds := rand.Intn(int(secondsDelta))
			interval := time.Duration(randomSeconds)*time.Second + minTasksInterval

			// Check the EC status
			err := t.sp.WaitEthClientSynced(false) // Force refresh the primary / fallback EC status
			if err != nil {
				errorLog.Println(err)
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			// Check the BC status
			err = t.sp.WaitBeaconClientSynced(false) // Force refresh the primary / fallback BC status
			if err != nil {
				errorLog.Println(err)
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			// Load contracts
			err = t.sp.RefreshRocketPoolContracts()
			if err != nil {
				errorLog.Println(fmt.Sprintf("error loading contract bindings: %s", err.Error()))
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			// Get the Beacon block
			//latestBlock, err := m.GetLatestFinalizedBeaconBlock()
			latestBlock, err := m.GetLatestBeaconBlock(t.ctx)
			if err != nil {
				errorLog.Println(fmt.Errorf("error getting latest Beacon block: %w", err))
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
				errorLog.Println(err)
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}
				continue
			}

			// Run the manual rewards tree generation
			if err := generateRewardsTree.Run(); err != nil {
				errorLog.Println(err)
			}
			if utils.SleepWithCancel(t.ctx, taskCooldown) {
				break
			}

			if isOnOdao {
				// Run the challenge check
				if err := respondChallenges.Run(); err != nil {
					errorLog.Println(err)
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Update the network state
				state, err := updateNetworkState(t.ctx, m, &updateLog, latestBlock)
				if err != nil {
					errorLog.Println(err)
					if utils.SleepWithCancel(t.ctx, taskCooldown) {
						break
					}
					continue
				}

				// Run the network balance submission check
				if err := submitNetworkBalances.Run(state); err != nil {
					errorLog.Println(err)
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				if !useRollingRecords {
					// Run the rewards tree submission check
					if err := submitRewardsTree_Stateless.Run(isOnOdao, state, latestBlock.Header.Slot); err != nil {
						errorLog.Println(err)
					}
					if utils.SleepWithCancel(t.ctx, taskCooldown) {
						break
					}
				} else {
					// Run the network balance and rewards tree submission check
					if err := submitRewardsTree_Rolling.Run(state); err != nil {
						errorLog.Println(err)
					}
					if utils.SleepWithCancel(t.ctx, taskCooldown) {
						break
					}
				}

				// Run the price submission check
				if err := submitRplPrice.Run(state); err != nil {
					errorLog.Println(err)
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the minipool dissolve check
				if err := dissolveTimedOutMinipools.Run(state); err != nil {
					errorLog.Println(err)
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the finalize proposals check
				if err := finalizePdaoProposals.Run(state); err != nil {
					errorLog.Println(err)
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the minipool scrub check
				if err := submitScrubMinipools.Run(state); err != nil {
					errorLog.Println(err)
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the bond cancel check
				if err := cancelBondReductions.Run(state); err != nil {
					errorLog.Println(err)
				}
				if utils.SleepWithCancel(t.ctx, taskCooldown) {
					break
				}

				// Run the solo migration check
				if err := checkSoloMigrations.Run(state); err != nil {
					errorLog.Println(err)
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
						errorLog.Println(err)
					}
				} else {
					// Run the network balance and rewards tree submission check
					if err := submitRewardsTree_Rolling.Run(nil); err != nil {
						errorLog.Println(err)
					}
				}
			}

			if utils.SleepWithCancel(t.ctx, interval) {
				break
			}
		}
	}()

	// Run metrics loop
	go func() {
		err := runMetricsServer(t.sp, log.NewColorLogger(MetricsColor), scrubCollector, bondReductionCollector, soloMigrationCollector)
		if err != nil {
			errorLog.Println(err)
		}
		wg.Done()
	}()

	// Wait for both threads to stop
	wg.Wait()
	return nil
}

// Update the latest network state at each cycle
func updateNetworkState(ctx context.Context, m *state.NetworkStateManager, log *log.ColorLogger, block beacon.BeaconBlock) (*state.NetworkState, error) {
	log.Print("Getting latest network state... ")
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
