package watchtower

import (
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"
	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/watchtower/collectors"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Config
var minTasksInterval, _ = time.ParseDuration("4m")
var maxTasksInterval, _ = time.ParseDuration("6m")
var taskCooldown, _ = time.ParseDuration("5s")

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

// Register watchtower command
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Run Rocket Pool watchtower activity daemon",
		Action: func(c *cli.Context) error {
			return run(c)
		},
	})
}

// Run daemon
func run(c *cli.Context) error {

	// Configure
	configureHTTP()

	// Wait until node is registered
	if err := services.WaitNodeRegistered(c, true); err != nil {
		return err
	}

	// Get services
	cfg, err := services.GetConfig(c)
	if err != nil {
		return err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return err
	}

	// Print the current mode
	if cfg.IsNativeMode {
		fmt.Println("Starting watchtower daemon in Native Mode.")
	} else {
		fmt.Println("Starting watchtower daemon in Docker Mode.")
	}

	// Check if rolling records are enabled
	useRollingRecords := cfg.Smartnode.UseRollingRecords.Value.(bool)
	if useRollingRecords {
		fmt.Println("***NOTE: EXPERIMENTAL ROLLING RECORDS ARE ENABLED, BE ADVISED!***")
	}

	// Initialize the metrics reporters
	scrubCollector := collectors.NewScrubCollector()
	bondReductionCollector := collectors.NewBondReductionCollector()
	soloMigrationCollector := collectors.NewSoloMigrationCollector()

	// Initialize error logger
	errorLog := log.NewColorLogger(ErrorColor)
	updateLog := log.NewColorLogger(UpdateColor)

	// Create the state manager
	m := state.NewNetworkStateManager(rp, cfg.Smartnode.GetStateManagerContracts(), bc, &updateLog)

	// Get the node address
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return fmt.Errorf("error getting node account: %w", err)
	}

	// Initialize tasks
	respondChallenges, err := newRespondChallenges(c, log.NewColorLogger(RespondChallengesColor), m)
	if err != nil {
		return fmt.Errorf("error during respond-to-challenges check: %w", err)
	}
	submitRplPrice, err := newSubmitRplPrice(c, log.NewColorLogger(SubmitRplPriceColor), errorLog)
	if err != nil {
		return fmt.Errorf("error during rpl price check: %w", err)
	}
	submitNetworkBalances, err := newSubmitNetworkBalances(c, log.NewColorLogger(SubmitNetworkBalancesColor), errorLog)
	if err != nil {
		return fmt.Errorf("error during network balances check: %w", err)
	}
	dissolveTimedOutMinipools, err := newDissolveTimedOutMinipools(c, log.NewColorLogger(DissolveTimedOutMinipoolsColor))
	if err != nil {
		return fmt.Errorf("error during timed-out minipools check: %w", err)
	}
	submitScrubMinipools, err := newSubmitScrubMinipools(c, log.NewColorLogger(SubmitScrubMinipoolsColor), errorLog, scrubCollector)
	if err != nil {
		return fmt.Errorf("error during scrub check: %w", err)
	}
	var submitRewardsTree_Stateless *submitRewardsTree_Stateless
	var submitRewardsTree_Rolling *submitRewardsTree_Rolling
	if !useRollingRecords {
		submitRewardsTree_Stateless, err = newSubmitRewardsTree_Stateless(c, log.NewColorLogger(SubmitRewardsTreeColor), errorLog, m)
		if err != nil {
			return fmt.Errorf("error during stateless rewards tree check: %w", err)
		}
	} else {
		submitRewardsTree_Rolling, err = newSubmitRewardsTree_Rolling(c, log.NewColorLogger(SubmitRewardsTreeColor), errorLog, m)
		if err != nil {
			return fmt.Errorf("error during rolling rewards tree check: %w", err)
		}
	}
	/*processPenalties, err := newProcessPenalties(c, log.NewColorLogger(ProcessPenaltiesColor), errorLog)
	if err != nil {
		return fmt.Errorf("error during penalties check: %w", err)
	}*/
	generateRewardsTree, err := newGenerateRewardsTree(c, log.NewColorLogger(SubmitRewardsTreeColor), errorLog)
	if err != nil {
		return fmt.Errorf("error during manual tree generation check: %w", err)
	}
	cancelBondReductions, err := newCancelBondReductions(c, log.NewColorLogger(CancelBondsColor), errorLog, bondReductionCollector)
	if err != nil {
		return fmt.Errorf("error during bond reduction cancel check: %w", err)
	}
	checkSoloMigrations, err := newCheckSoloMigrations(c, log.NewColorLogger(CheckSoloMigrationsColor), errorLog, soloMigrationCollector)
	if err != nil {
		return fmt.Errorf("error during solo migration check: %w", err)
	}
	finalizePdaoProposals, err := newFinalizePdaoProposals(c, log.NewColorLogger(FinalizeProposalsColor))
	if err != nil {
		return fmt.Errorf("error creating finalize-pdao-proposals task: %w", err)
	}

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
			err := services.WaitEthClientSynced(c, false) // Force refresh the primary / fallback EC status
			if err != nil {
				errorLog.Println(err)
				time.Sleep(taskCooldown)
				continue
			}

			// Check the BC status
			err = services.WaitBeaconClientSynced(c, false) // Force refresh the primary / fallback BC status
			if err != nil {
				errorLog.Println(err)
				time.Sleep(taskCooldown)
				continue
			}

			// Get the Beacon block
			//latestBlock, err := m.GetLatestFinalizedBeaconBlock()
			latestBlock, err := m.GetLatestBeaconBlock()
			if err != nil {
				errorLog.Println(fmt.Errorf("error getting latest Beacon block: %w", err))
				time.Sleep(taskCooldown)
				continue
			}

			// Check if on the Oracle DAO
			isOnOdao, err := isOnOracleDAO(rp, nodeAccount.Address, latestBlock)
			if err != nil {
				errorLog.Println(err)
				time.Sleep(taskCooldown)
				continue
			}

			// Run the manual rewards tree generation
			if err := generateRewardsTree.run(); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(taskCooldown)

			if isOnOdao {
				// Run the challenge check
				if err := respondChallenges.run(); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Update the network state
				state, err := updateNetworkState(m, &updateLog, latestBlock)
				if err != nil {
					errorLog.Println(err)
					time.Sleep(taskCooldown)
					continue
				}

				// Run the network balance submission check
				if err := submitNetworkBalances.run(state); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				if !useRollingRecords {
					// Run the rewards tree submission check
					if err := submitRewardsTree_Stateless.Run(isOnOdao, state, latestBlock.Slot); err != nil {
						errorLog.Println(err)
					}
					time.Sleep(taskCooldown)
				} else {
					// Run the network balance and rewards tree submission check
					if err := submitRewardsTree_Rolling.run(state); err != nil {
						errorLog.Println(err)
					}
					time.Sleep(taskCooldown)
				}

				// Run the price submission check
				if err := submitRplPrice.run(state); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the minipool dissolve check
				if err := dissolveTimedOutMinipools.run(state); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the finalize proposals check
				if err := finalizePdaoProposals.run(state); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the minipool scrub check
				if err := submitScrubMinipools.run(state); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the bond cancel check
				if err := cancelBondReductions.run(state); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the solo migration check
				if err := checkSoloMigrations.run(state); err != nil {
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
					if err := submitRewardsTree_Stateless.Run(isOnOdao, nil, latestBlock.Slot); err != nil {
						errorLog.Println(err)
					}
				} else {
					// Run the network balance and rewards tree submission check
					if err := submitRewardsTree_Rolling.run(nil); err != nil {
						errorLog.Println(err)
					}
				}
			}

			time.Sleep(interval)
		}
		wg.Done()
	}()

	// Run metrics loop
	go func() {
		err := runMetricsServer(c, log.NewColorLogger(MetricsColor), scrubCollector, bondReductionCollector, soloMigrationCollector)
		if err != nil {
			errorLog.Println(err)
		}
		wg.Done()
	}()

	// Wait for both threads to stop
	wg.Wait()
	return nil
}

// Configure HTTP transport settings
func configureHTTP() {

	// The daemon makes a large number of concurrent RPC requests to the Eth1 client
	// The HTTP transport is set to cache connections for future re-use equal to the maximum expected number of concurrent requests
	// This prevents issues related to memory consumption and address allowance from repeatedly opening and closing connections
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = MaxConcurrentEth1Requests

}

// Update the latest network state at each cycle
func updateNetworkState(m *state.NetworkStateManager, log *log.ColorLogger, block beacon.BeaconBlock) (*state.NetworkState, error) {
	log.Print("Getting latest network state... ")
	// Get the state of the network
	state, err := m.GetStateForSlot(block.Slot)
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

	nodeTrusted, err := trustednode.GetMemberExists(rp, nodeAddress, opts)
	if err != nil {
		return false, fmt.Errorf("error checking if node is in the Oracle DAO for Beacon block %d, EL block %d: %w", block.Slot, block.ExecutionBlockNumber, err)
	}
	return nodeTrusted, nil
}

// Check if Houston has been deployed yet
func printHoustonMessage(log *log.ColorLogger) {
	log.Println(`
*       .
*      / \
*     |.'.|
*     |'.'|
*   ,'|   |'.
*  |,-'-|-'-.|
*   __|_| |         _        _      _____           _
*  | ___ \|        | |      | |    | ___ \         | |
*  | |_/ /|__   ___| | _____| |_   | |_/ /__   ___ | |
*  |    // _ \ / __| |/ / _ \ __|  |  __/ _ \ / _ \| |
*  | |\ \ (_) | (__|   <  __/ |_   | | | (_) | (_) | |
*  \_| \_\___/ \___|_|\_\___|\__|  \_|  \___/ \___/|_|
* +---------------------------------------------------+
* |    DECENTRALISED STAKING PROTOCOL FOR ETHEREUM    |
* +---------------------------------------------------+
*
* =============== Houston has launched! ===============
`)
}
