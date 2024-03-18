package node

import (
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/rocketpool/node/collectors"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/alerting"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore/lighthouse"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore/nimbus"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore/prysm"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore/teku"
	"github.com/rocket-pool/smartnode/shared/utils/log"
)

// Config
var tasksInterval, _ = time.ParseDuration("5m")
var taskCooldown, _ = time.ParseDuration("10s")
var totalEffectiveStakeCooldown, _ = time.ParseDuration("1h")

const (
	MaxConcurrentEth1Requests = 200

	StakePrelaunchMinipoolsColor = color.FgBlue
	DownloadRewardsTreesColor    = color.FgGreen
	MetricsColor                 = color.FgHiYellow
	ManageFeeRecipientColor      = color.FgHiCyan
	PromoteMinipoolsColor        = color.FgMagenta
	ReduceBondAmountColor        = color.FgHiBlue
	DefendPdaoPropsColor         = color.FgYellow
	VerifyPdaoPropsColor         = color.FgYellow
	DistributeMinipoolsColor     = color.FgHiGreen
	ErrorColor                   = color.FgRed
	WarningColor                 = color.FgYellow
	UpdateColor                  = color.FgHiWhite
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
	err := deployDefaultFeeRecipientFile(c)
	if err != nil {
		return err
	}

	// Clean up old fee recipient files
	err = removeLegacyFeeRecipientFiles(c)
	if err != nil {
		return err
	}

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
		fmt.Println("Starting node daemon in Native Mode.")
	} else {
		fmt.Println("Starting node daemon in Docker Mode.")
	}

	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return fmt.Errorf("error getting node account: %w", err)
	}

	// Initialize loggers
	errorLog := log.NewColorLogger(ErrorColor)
	updateLog := log.NewColorLogger(UpdateColor)

	// Create the state manager
	m, err := state.NewNetworkStateManager(rp, cfg, rp.Client, bc, &updateLog)
	if err != nil {
		return err
	}
	stateLocker := collectors.NewStateLocker()

	// Initialize tasks
	manageFeeRecipient, err := newManageFeeRecipient(c, log.NewColorLogger(ManageFeeRecipientColor))
	if err != nil {
		return err
	}
	distributeMinipools, err := newDistributeMinipools(c, log.NewColorLogger(DistributeMinipoolsColor))
	if err != nil {
		return err
	}
	stakePrelaunchMinipools, err := newStakePrelaunchMinipools(c, log.NewColorLogger(StakePrelaunchMinipoolsColor))
	if err != nil {
		return err
	}
	promoteMinipools, err := newPromoteMinipools(c, log.NewColorLogger(PromoteMinipoolsColor))
	if err != nil {
		return err
	}
	downloadRewardsTrees, err := newDownloadRewardsTrees(c, log.NewColorLogger(DownloadRewardsTreesColor))
	if err != nil {
		return err
	}
	reduceBonds, err := newReduceBonds(c, log.NewColorLogger(ReduceBondAmountColor))
	if err != nil {
		return err
	}
	defendPdaoProps, err := newDefendPdaoProps(c, log.NewColorLogger(DefendPdaoPropsColor))
	if err != nil {
		return err
	}
	var verifyPdaoProps *verifyPdaoProps
	// Make sure the user opted into this duty
	verifyEnabled := cfg.Smartnode.VerifyProposals.Value.(bool)
	if verifyEnabled {
		verifyPdaoProps, err = newVerifyPdaoProps(c, log.NewColorLogger(VerifyPdaoPropsColor))
		if err != nil {
			return err
		}
	}

	// Wait group to handle the various threads
	wg := new(sync.WaitGroup)
	wg.Add(2)

	// Timestamp for caching total effective RPL stake
	lastTotalEffectiveStakeTime := time.Unix(0, 0)

	// Run task loop
	isHoustonDeployedMasterFlag := false
	go func() {
		// we assume clients are synced on startup so that we don't send unnecessary alerts
		wasExecutionClientSynced := true
		wasBeaconClientSynced := true
		for {
			// Check the EC status
			err := services.WaitEthClientSynced(c, false) // Force refresh the primary / fallback EC status
			if err != nil {
				wasExecutionClientSynced = false
				errorLog.Printlnf("Execution client not synced: %s. Waiting for sync...", err.Error())
				time.Sleep(taskCooldown)
				continue
			}

			if !wasExecutionClientSynced {
				updateLog.Println("Execution client is now synced.")
				wasExecutionClientSynced = true
				alerting.AlertExecutionClientSyncComplete(cfg)
			}

			// Check the BC status
			err = services.WaitBeaconClientSynced(c, false) // Force refresh the primary / fallback BC status
			if err != nil {
				// NOTE: if not synced, it returns an error - so there isn't necessarily an underlying issue
				wasBeaconClientSynced = false
				errorLog.Printlnf("Beacon client not synced: %s. Waiting for sync...", err.Error())
				time.Sleep(taskCooldown)
				continue
			}

			if !wasBeaconClientSynced {
				updateLog.Println("Beacon client is now synced.")
				wasBeaconClientSynced = true
				alerting.AlertBeaconClientSyncComplete(cfg)
			}

			// Update the network state
			updateTotalEffectiveStake := false
			if time.Since(lastTotalEffectiveStakeTime) > totalEffectiveStakeCooldown {
				updateTotalEffectiveStake = true
				lastTotalEffectiveStakeTime = time.Now() // Even if the call below errors out, this will prevent contant errors related to this flag
			}
			state, totalEffectiveStake, err := updateNetworkState(m, &updateLog, nodeAccount.Address, updateTotalEffectiveStake)
			if err != nil {
				errorLog.Println(err)
				time.Sleep(taskCooldown)
				continue
			}
			stateLocker.UpdateState(state, totalEffectiveStake)

			// Check for Houston
			if !isHoustonDeployedMasterFlag && state.IsHoustonDeployed {
				printHoustonMessage(&updateLog)
				isHoustonDeployedMasterFlag = true
			}

			// Manage the fee recipient for the node
			if err := manageFeeRecipient.run(state); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(taskCooldown)

			// Run the rewards download check
			if err := downloadRewardsTrees.run(state); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(taskCooldown)

			if state.IsHoustonDeployed {
				// Run the pDAO proposal defender
				if err := defendPdaoProps.run(state); err != nil {
					errorLog.Println(err)
				}
				time.Sleep(taskCooldown)

				// Run the pDAO proposal verifier
				if verifyPdaoProps != nil {
					if err := verifyPdaoProps.run(state); err != nil {
						errorLog.Println(err)
					}
					time.Sleep(taskCooldown)
				}
			}

			// Run the minipool stake check
			if err := stakePrelaunchMinipools.run(state); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(taskCooldown)

			// Run the balance distribution check
			if err := distributeMinipools.run(state); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(taskCooldown)

			// Run the reduce bond check
			if err := reduceBonds.run(state); err != nil {
				errorLog.Println(err)
			}
			time.Sleep(taskCooldown)

			// Run the minipool promotion check
			if err := promoteMinipools.run(state); err != nil {
				errorLog.Println(err)
			}

			time.Sleep(tasksInterval)
		}
		wg.Done()
	}()

	// Run metrics loop
	go func() {
		err := runMetricsServer(c, log.NewColorLogger(MetricsColor), stateLocker)
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

// Copy the default fee recipient file into the proper location
func deployDefaultFeeRecipientFile(c *cli.Context) error {

	cfg, err := services.GetConfig(c)
	if err != nil {
		return err
	}

	feeRecipientPath := cfg.Smartnode.GetFeeRecipientFilePath()
	_, err = os.Stat(feeRecipientPath)
	if os.IsNotExist(err) {
		// Make sure the validators dir is created
		validatorsFolder := filepath.Dir(feeRecipientPath)
		err = os.MkdirAll(validatorsFolder, 0755)
		if err != nil {
			return fmt.Errorf("could not create validators directory: %w", err)
		}

		// Create the file
		var defaultFeeRecipientFileContents string
		if cfg.IsNativeMode {
			// Native mode needs an environment variable definition
			defaultFeeRecipientFileContents = fmt.Sprintf("FEE_RECIPIENT=%s", cfg.Smartnode.GetRethAddress().Hex())
		} else {
			// Docker and Hybrid just need the address itself
			defaultFeeRecipientFileContents = cfg.Smartnode.GetRethAddress().Hex()
		}
		err := os.WriteFile(feeRecipientPath, []byte(defaultFeeRecipientFileContents), 0664)
		if err != nil {
			return fmt.Errorf("could not write default fee recipient file to %s: %w", feeRecipientPath, err)
		}
	} else if err != nil {
		return fmt.Errorf("Error checking fee recipient file status: %w", err)
	}

	return nil

}

// Remove the old fee recipient files that were created in v1.5.0
func removeLegacyFeeRecipientFiles(c *cli.Context) error {

	legacyFeeRecipientFile := "rp-fee-recipient.txt"

	cfg, err := services.GetConfig(c)
	if err != nil {
		return err
	}

	validatorsFolder := cfg.Smartnode.GetValidatorKeychainPath()

	// Remove the legacy files
	keystoreDirs := []string{lighthouse.KeystoreDir, nimbus.KeystoreDir, prysm.KeystoreDir, teku.KeystoreDir}
	for _, keystoreDir := range keystoreDirs {
		oldFile := filepath.Join(validatorsFolder, keystoreDir, legacyFeeRecipientFile)
		_, err = os.Stat(oldFile)
		if !os.IsNotExist(err) {
			err = os.Remove(oldFile)
			if err != nil {
				fmt.Printf("NOTE: Couldn't remove old fee recipient file (%s): %s\nThis file is no longer used, you may remove it manually if you wish.\n", oldFile, err.Error())
			}
		}
	}

	return nil

}

// Update the latest network state at each cycle
func updateNetworkState(m *state.NetworkStateManager, log *log.ColorLogger, nodeAddress common.Address, calculateTotalEffectiveStake bool) (*state.NetworkState, *big.Int, error) {
	// Get the state of the network
	state, totalEffectiveStake, err := m.GetHeadStateForNode(nodeAddress, calculateTotalEffectiveStake)
	if err != nil {
		return nil, nil, fmt.Errorf("error updating network state: %w", err)
	}
	return state, totalEffectiveStake, nil
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
