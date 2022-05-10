package watchtower

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool/watchtower/collectors"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Config
var minTasksInterval, _ = time.ParseDuration("4m")
var maxTasksInterval, _ = time.ParseDuration("6m")
var taskCooldown, _ = time.ParseDuration("10s")

const (
	MaxConcurrentEth1Requests = 200

	RespondChallengesColor           = color.FgWhite
	ClaimRplRewardsColor             = color.FgGreen
	SubmitRplPriceColor              = color.FgYellow
	SubmitNetworkBalancesColor       = color.FgYellow
	SubmitWithdrawableMinipoolsColor = color.FgBlue
	DissolveTimedOutMinipoolsColor   = color.FgMagenta
	ProcessWithdrawalsColor          = color.FgCyan
	SubmitScrubMinipoolsColor        = color.FgHiGreen
	ErrorColor                       = color.FgRed
	MetricsColor                     = color.FgHiYellow
	SubmitRewardsTreeColor           = color.FgHiCyan
	WarningColor                     = color.FgYellow
	ProcessPenaltiesColor            = color.FgHiMagenta
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

	// Initialize the scrub metrics reporter
	scrubCollector := collectors.NewScrubCollector()

	// Initialize error logger
	errorLog := log.NewColorLogger(ErrorColor)

	// Initialize tasks
	respondChallenges, err := newRespondChallenges(c, log.NewColorLogger(RespondChallengesColor))
	if err != nil {
		return err
	}
	claimRplRewards, err := newClaimRplRewards(c, log.NewColorLogger(ClaimRplRewardsColor))
	if err != nil {
		return err
	}
	submitRplPrice, err := newSubmitRplPrice(c, log.NewColorLogger(SubmitRplPriceColor))
	if err != nil {
		return err
	}
	submitNetworkBalances, err := newSubmitNetworkBalances(c, log.NewColorLogger(SubmitNetworkBalancesColor))
	if err != nil {
		return err
	}
	submitWithdrawableMinipools, err := newSubmitWithdrawableMinipools(c, log.NewColorLogger(SubmitWithdrawableMinipoolsColor))
	if err != nil {
		return err
	}
	dissolveTimedOutMinipools, err := newDissolveTimedOutMinipools(c, log.NewColorLogger(DissolveTimedOutMinipoolsColor))
	if err != nil {
		return err
	}
	processWithdrawals, err := newProcessWithdrawals(c, log.NewColorLogger(ProcessWithdrawalsColor))
	if err != nil {
		return err
	}
	submitScrubMinipools, err := newSubmitScrubMinipools(c, log.NewColorLogger(SubmitScrubMinipoolsColor), scrubCollector)
	if err != nil {
		return err
	}
	submitRewardsTree, err := newSubmitRewardsTree(c, log.NewColorLogger(SubmitRewardsTreeColor), errorLog)
	if err != nil {
		return err
	}
	processPenalties, err := newProcessPenalties(c, log.NewColorLogger(ProcessPenaltiesColor))
	if err != nil {
		return err
	}

	intervalDelta := maxTasksInterval - minTasksInterval
	secondsDelta := intervalDelta.Seconds()

	// Wait group to handle the various threads
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// Run task loop
	isUpdateDeployed := false
	go func() {
		for {
			// Randomize the next interval
			randomSeconds := rand.Intn(int(secondsDelta))
			interval := time.Duration(randomSeconds)*time.Second + minTasksInterval

			// Check the EC status
			err := services.WaitEthClientSynced(c, false) // Force refresh the primary / fallback EC status
			if err != nil {
				errorLog.Println(err)
			} else {
				// Run the challenge check
				if err := respondChallenges.run(); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the oDAO rewards check
				if !isUpdateDeployed {
					// Only run auto-claims during the legacy period
					isUpdateDeployed, err = claimRplRewards.run()
					if err != nil {
						errorLog.Println(err)
					}
					time.Sleep(taskCooldown)
				}

				// Run the rewards tree submission check
				if err := submitRewardsTree.run(); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the price submission check
				if err := submitRplPrice.run(); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the network balance submission check
				if err := submitNetworkBalances.run(); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the withdrawable status submission check
				if err := submitWithdrawableMinipools.run(); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the minipool dissolve check
				if err := dissolveTimedOutMinipools.run(); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the withdrawal processing check
				if err := processWithdrawals.run(); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the minipool scrub check
				if err := submitScrubMinipools.run(); err != nil {
					errorLog.Println(err)
				}
			}
			time.Sleep(taskCooldown)
			if err := processPenalties.run(); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(interval)
		}
		wg.Done()
	}()

	// Run metrics loop
	go func() {
		err := runMetricsServer(c, log.NewColorLogger(MetricsColor), scrubCollector)
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
