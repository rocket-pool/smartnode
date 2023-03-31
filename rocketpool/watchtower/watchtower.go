package watchtower

import (
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"os"
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
	"github.com/rocket-pool/smartnode/shared/services/beacon/client"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Config
var minTasksInterval, _ = time.ParseDuration("30s")
var maxTasksInterval, _ = time.ParseDuration("1m")
var taskCooldown, _ = time.ParseDuration("2s")

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

	// Initialize the metrics reporters
	scrubCollector := collectors.NewScrubCollector()
	bondReductionCollector := collectors.NewBondReductionCollector()
	soloMigrationCollector := collectors.NewSoloMigrationCollector()

	// Initialize error logger
	errorLog := log.NewColorLogger(ErrorColor)
	updateLog := log.NewColorLogger(UpdateColor)

	var bc beacon.Client
	// Override the beacon client, if requested
	if beaconOverride := os.Getenv("TREEGEN_BEACON_CLIENT_ENDPOINT"); beaconOverride != "" {
		updateLog.Printlnf("Overriding the Beacon Node URL to %s", beaconOverride)
		bc = client.NewStandardHttpClient(beaconOverride)
	} else {
		var err error

		bc, err = services.GetBeaconClient(c)
		if err != nil {
			return err
		}
	}

	// Create the state manager
	m, err := state.NewNetworkStateManager(rp, cfg, rp.Client, bc, &updateLog)
	if err != nil {
		return err
	}

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
	submitRewardsTree, err := newSubmitRewardsTree(c, log.NewColorLogger(SubmitRewardsTreeColor), errorLog, m)
	if err != nil {
		return fmt.Errorf("error during rewards tree check: %w", err)
	}
	/*processPenalties, err := newProcessPenalties(c, log.NewColorLogger(ProcessPenaltiesColor), errorLog)
	if err != nil {
		return fmt.Errorf("error during penalties check: %w", err)
	}*/
	generateRewardsTree, err := newGenerateRewardsTree(c, log.NewColorLogger(SubmitRewardsTreeColor), errorLog, m)
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

	intervalDelta := maxTasksInterval - minTasksInterval
	secondsDelta := intervalDelta.Seconds()

	// Wait group to handle the various threads
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// Run task loop
	isAtlasDeployedMasterFlag := false
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
				// Update the network state
				state, err := updateNetworkState(m, &updateLog, latestBlock)
				if err != nil {
					errorLog.Println(err)
					time.Sleep(taskCooldown)
					continue
				}

				// Check for Atlas
				if !isAtlasDeployedMasterFlag && state.IsAtlasDeployed {
					printAtlasMessage(&updateLog)
					isAtlasDeployedMasterFlag = true
				}

				// Run the rewards tree submission check
				if err := submitRewardsTree.run(isOnOdao, state, latestBlock.Slot, isAtlasDeployedMasterFlag); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the challenge check
				if err := respondChallenges.run(isAtlasDeployedMasterFlag); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the price submission check
				if err := submitRplPrice.run(state, isAtlasDeployedMasterFlag); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the network balance submission check
				if err := submitNetworkBalances.run(state, isAtlasDeployedMasterFlag); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the minipool dissolve check
				if err := dissolveTimedOutMinipools.run(state, isAtlasDeployedMasterFlag); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the minipool scrub check
				if err := submitScrubMinipools.run(state, isAtlasDeployedMasterFlag); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the bond cancel check
				if err := cancelBondReductions.run(state, isAtlasDeployedMasterFlag); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the solo migration check
				if err := checkSoloMigrations.run(state, isAtlasDeployedMasterFlag); err != nil {
					errorLog.Println(err)
				}
				/*time.Sleep(taskCooldown)

				// Run the fee recipient penalty check
				if err := processPenalties.run(); err != nil {
					errorLog.Println(err)
				}*/
				// DISABLED until MEV-Boost can support it
			} else {
				// Check for Atlas
				isAtlasDeployed, err := state.IsAtlasDeployed(rp, &bind.CallOpts{
					BlockNumber: big.NewInt(0).SetUint64(latestBlock.ExecutionBlockNumber),
				})
				if err != nil {
					errorLog.Println(fmt.Errorf("error checking if Atlas is deployed: %w", err))
					time.Sleep(taskCooldown)
					continue
				}

				// Run the rewards tree submission check
				if err := submitRewardsTree.run(isOnOdao, nil, latestBlock.Slot, isAtlasDeployed); err != nil {
					errorLog.Println(err)
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

	// The watchtower daemon makes a large number of concurrent RPC requests to the Eth1 client
	// The HTTP transport is set to cache connections for future re-use equal to the maximum expected number of concurrent requests
	// This prevents issues related to memory consumption and address allowance from repeatedly opening and closing connections
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = MaxConcurrentEth1Requests

}

// Check if Atlas has been deployed yet
func printAtlasMessage(log *log.ColorLogger) {
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
* ================ Atlas has launched! ================
`)
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
