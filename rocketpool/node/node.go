package node

import (
	"net/http"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Config
var tasksInterval, _ = time.ParseDuration("5m")
var taskCooldown, _ = time.ParseDuration("10s")

const (
	MaxConcurrentEth1Requests = 200

	ClaimRplRewardsColor         = color.FgGreen
	StakePrelaunchMinipoolsColor = color.FgBlue
	MetricsColor                 = color.FgHiYellow
	ErrorColor                   = color.FgRed
	WarningColor                 = color.FgYellow
)

// Register node command
func RegisterCommands(app *cli.App, name string, aliases []string) {
	app.Commands = append(app.Commands, cli.Command{
		Name:    name,
		Aliases: aliases,
		Usage:   "Run Rocket Pool node activity daemon",
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

	// Get the EC manager
	ec, err := services.GetEthClient(c)
	if err != nil {
		return err
	}

	// Initialize tasks
	claimRplRewards, err := newClaimRplRewards(c, log.NewColorLogger(ClaimRplRewardsColor))
	if err != nil {
		return err
	}
	stakePrelaunchMinipools, err := newStakePrelaunchMinipools(c, log.NewColorLogger(StakePrelaunchMinipoolsColor))
	if err != nil {
		return err
	}

	// Initialize loggers
	errorLog := log.NewColorLogger(ErrorColor)
	warningLog := log.NewColorLogger(WarningColor)

	// Wait group to handle the various threads
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// Run task loop
	go func() {
		for {
			// Check the EC status
			_, log, err := ec.CheckStatus()
			if log != "" {
				warningLog.Println(log)
			}
			if err != nil {
				errorLog.Println(err)
			} else {
				// Run the rewards check
				if err := claimRplRewards.run(); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the minipool stake check
				if err := stakePrelaunchMinipools.run(); err != nil {
					errorLog.Println(err)
				}
			}
			time.Sleep(tasksInterval)
		}
		wg.Done()
	}()

	// Run metrics loop
	go func() {
		err := runMetricsServer(c, log.NewColorLogger(MetricsColor))
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
