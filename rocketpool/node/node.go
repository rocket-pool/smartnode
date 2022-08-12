package node

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
	DownloadRewardsTreesColor    = color.FgGreen
	MetricsColor                 = color.FgHiYellow
	ManageFeeRecipientColor      = color.FgHiCyan
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

	// Handle the initial fee recipient file deployment
	err := deployFeeRecipientFile(c)
	if err != nil {
		return err
	}

	// Configure
	configureHTTP()

	// Wait until node is registered
	if err := services.WaitNodeRegistered(c, true); err != nil {
		return err
	}

	// Initialize tasks
	manageFeeRecipient, err := newManageFeeRecipient(c, log.NewColorLogger(ManageFeeRecipientColor))
	if err != nil {
		return err
	}
	claimRplRewards, err := newClaimRplRewards(c, log.NewColorLogger(ClaimRplRewardsColor))
	if err != nil {
		return err
	}
	stakePrelaunchMinipools, err := newStakePrelaunchMinipools(c, log.NewColorLogger(StakePrelaunchMinipoolsColor))
	if err != nil {
		return err
	}
	downloadRewardsTrees, err := newDownloadRewardsTrees(c, log.NewColorLogger(DownloadRewardsTreesColor))
	if err != nil {
		return err
	}

	// Initialize loggers
	errorLog := log.NewColorLogger(ErrorColor)

	// Wait group to handle the various threads
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// Run task loop
	go func() {
		for {
			// Check the EC status
			err := services.WaitEthClientSynced(c, false) // Force refresh the primary / fallback EC status
			if err != nil {
				errorLog.Println(err)
			} else {
				// Check the BC status
				err := services.WaitBeaconClientSynced(c, false) // Force refresh the primary / fallback BC status
				if err != nil {
					errorLog.Println(err)
				} else {
					// Manage the fee recipient for the node
					if err := manageFeeRecipient.run(); err != nil {
						errorLog.Println(err)
					}
					time.Sleep(taskCooldown)

					// Run auto-claims during the legacy period
					if err := claimRplRewards.run(); err != nil {
						errorLog.Println(err)
					}
					time.Sleep(taskCooldown)

					// Run the rewards download check
					if err := downloadRewardsTrees.run(); err != nil {
						errorLog.Println(err)
					}
					time.Sleep(taskCooldown)

					// Run the minipool stake check
					if err := stakePrelaunchMinipools.run(); err != nil {
						errorLog.Println(err)
					}
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

// Copy the default fee recipient file into the proper location
func deployFeeRecipientFile(c *cli.Context) error {

	cfg, err := services.GetConfig(c)
	if err != nil {
		return err
	}

	feeRecipientPath := cfg.Smartnode.GetFeeRecipientFilePath()
	_, err = os.Stat(feeRecipientPath)
	if os.IsNotExist(err) {
		// Fee recipient file didn't exist so copy the default over
		defaultFeeRecipientPath := cfg.Smartnode.GetDefaultFeeRecipientFilePath(true)
		_, err = os.Stat(defaultFeeRecipientPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("default fee recipient file did not exist at %s.", defaultFeeRecipientPath)
		}

		// Make sure the validators dir is created
		validatorsFolder := filepath.Dir(defaultFeeRecipientPath)
		err = os.MkdirAll(validatorsFolder, 0755)
		if err != nil {
			return fmt.Errorf("could not create validators directory: %w", err)
		}

		// Copy the file
		bytes, err := ioutil.ReadFile(defaultFeeRecipientPath)
		if err != nil {
			return fmt.Errorf("could not read default fee recipient file: %w", err)
		}
		err = ioutil.WriteFile(feeRecipientPath, bytes, 0664)
		if err != nil {
			return fmt.Errorf("could not write default fee recipient file to %s: %w", feeRecipientPath, err)
		}
	}

	return nil

}
