package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/rivo/tview"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v2"

	"github.com/dustin/go-humanize"
	cliconfig "github.com/rocket-pool/smartnode/rocketpool-cli/service/config"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/shirou/gopsutil/v3/disk"
)

// Settings
const (
	ExporterContainerSuffix         string = "_exporter"
	ValidatorContainerSuffix        string = "_validator"
	BeaconContainerSuffix           string = "_eth2"
	ExecutionContainerSuffix        string = "_eth1"
	NodeContainerSuffix             string = "_node"
	ApiContainerSuffix              string = "_api"
	WatchtowerContainerSuffix       string = "_watchtower"
	PruneProvisionerContainerSuffix string = "_prune_provisioner"
	clientDataVolumeName            string = "/ethclient"
	dataFolderVolumeName            string = "/.rocketpool/data"

	PruneFreeSpaceRequired           uint64 = 50 * 1024 * 1024 * 1024
	NethermindPruneFreeSpaceRequired uint64 = 250 * 1024 * 1024 * 1024

	// Capture the entire image name, including the custom registry if present.
	// Just ignore the version tag.
	dockerImageRegex string = "(?P<image>.+):.*"

	clearLine string = "\033[2K"
)

// Install the Rocket Pool service
func installService(yes, verbose, noDeps bool, path string) error {
	dataPath := ""

	// Prompt for confirmation
	if !(yes || prompt.Confirm(
		"%s",
		fmt.Sprintf("The Rocket Pool %s service will be installed.\n\n", shared.RocketPoolVersion())+
			color.Green("If you're upgrading, your existing configuration will be backed up and preserved.\nAll of your previous settings will be migrated automatically.\n")+
			"Are you sure you want to continue?",
	)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Attempt to load the config to see if any settings need to be passed along to the install script
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading old configuration: %w", err)
	}
	if !isNew {
		dataPath = cfg.Smartnode.DataPath.Value.(string)
		dataPath, err = homedir.Expand(dataPath)
		if err != nil {
			return fmt.Errorf("error getting data path from old configuration: %w", err)
		}
	}

	// Install service
	err = rp.InstallService(verbose, noDeps, path, dataPath)
	if err != nil {
		return err
	}

	// Print success message & return
	fmt.Println("")
	fmt.Println("The Rocket Pool service was successfully installed!")

	printPatchNotes()

	// Reload the config after installation
	_, isNew, err = rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading new configuration: %w", err)
	}

	// Report next steps
	color.LightBluePrintln("=== Next Steps ===")
	color.LightBluePrintln("Run 'rocketpool service config' to review the settings changes for this update, or to continue setting up your node.")

	// Print the docker permissions notice
	if isNew {
		fmt.Println()
		color.YellowPrintln("NOTE:")
		color.YellowPrintln("Since this is your first time installing Rocket Pool, please start a new shell session by logging out and back in or restarting the machine.")
		color.YellowPrintln("This is necessary for your user account to have permissions to use Docker.")
	}

	return nil

}

// Print the latest patch notes for this release
func printPatchNotes() {

	fmt.Print(shared.Logo())
	fmt.Println()
	fmt.Println()
	color.GreenPrintf("=== Smart Node v%s ===\n", shared.RocketPoolVersion())
	fmt.Println()
	fmt.Println("Changes you should be aware of before starting:")
	fmt.Println()
	fmt.Println("This Smart Node version is compatible with the Saturn 1 upgrade. The upgrade took place on Feb 18, 2026 00:00:00 UTC.")
	fmt.Println("For more information about the biggest Rocket Pool upgrade ever, please see the official documentation: https://docs.rocketpool.net/upgrades/saturn-1/whats-new")
	fmt.Println()
}

// Install the Rocket Pool update tracker for the metrics dashboard
func installUpdateTracker(yes, verbose bool) error {

	// Prompt for confirmation
	if !(yes || prompt.Confirm(
		"This will add the ability to display any available Operating System updates or new Rocket Pool versions on the metrics dashboard. "+
			"Are you sure you want to install the update tracker?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Install service
	err := rp.InstallUpdateTracker(verbose)
	if err != nil {
		return err
	}

	// Print success message & return
	fmt.Println("")
	fmt.Println("The Rocket Pool update tracker service was successfully installed!")
	fmt.Println("")
	color.YellowPrintln("NOTE:")
	color.YellowPrintln("Please restart the Smart Node stack to enable update tracking on the metrics dashboard.")
	fmt.Println("")
	return nil

}

// View the Rocket Pool service status
func serviceStatus(composeFiles []string) error {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = cliutils.PrintNetwork(cfg.GetNetwork(), isNew)
	if err != nil {
		return err
	}

	// Print service status
	return rp.PrintServiceStatus(composeFiles)

}

func configureServicePrecheck() (isNew bool, cfg, oldCfg *config.RocketPoolConfig, err error) {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Load the config, checking to see if it's new (hasn't been installed before)
	cfg, isNew, err = rp.LoadConfig()
	if err != nil {
		return false, nil, nil, fmt.Errorf("error loading user settings: %w", err)
	}

	// Check if this is a new install
	isUpdate, err := rp.IsFirstRun()
	if err != nil {
		return false, nil, nil, fmt.Errorf("error checking for first-run status: %w", err)
	}

	// For upgrades, move the config to the old one and create a new upgraded copy
	if isUpdate {
		oldCfg = cfg
		cfg = cfg.CreateCopy()
		err = cfg.UpdateDefaults()
		if err != nil {
			return false, nil, nil, fmt.Errorf("error upgrading configuration with the latest parameters: %w", err)
		}
	}

	cfg.ConfirmUpdateSuggestedSettings()

	return isNew, cfg, oldCfg, nil
}

// This function is the exception to the rule-
// we pass cli.Context here and here only because
// otherwise it's very difficult to set config values by CLI flag.
func configureServiceHeadless(c *cli.Command) error {
	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	_, cfg, _, err := configureServicePrecheck()
	if err != nil {
		return err
	}

	// Root params
	for _, param := range cfg.GetParameters() {
		err := updateConfigParamFromCliArg(c, "", param, cfg)
		if err != nil {
			return fmt.Errorf("error updating config from provided arguments: %w", err)
		}
	}

	// Subconfigs
	for sectionName, subconfig := range cfg.GetSubconfigs() {
		for _, param := range subconfig.GetParameters() {
			err := updateConfigParamFromCliArg(c, sectionName, param, cfg)
			if err != nil {
				return fmt.Errorf("error updating config from provided arguments: %w", err)
			}
		}
	}

	return nil
}

// Configure the service
func configureService(configPath string, isNative, yes bool, composeFiles []string) error {
	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	isNew, cfg, oldCfg, err := configureServicePrecheck()
	if err != nil {
		return err
	}

	isUpdate := !isNew && oldCfg != nil

	app := tview.NewApplication()
	md := cliconfig.NewMainDisplay(app, oldCfg, cfg, isNew, isUpdate, isNative)
	err = app.Run()
	if err != nil {
		return err
	}

	// Deal with saving the config and printing the changes
	if md.ShouldSave {
		// Save the config
		err = rp.SaveConfig(md.Config)
		if err != nil {
			return fmt.Errorf("error saving config: %w", err)
		}
		fmt.Println("Your changes have been saved!")

		// Exit immediately if we're in native mode
		if isNative {
			fmt.Println("Please restart your daemon service for them to take effect.")
			return nil
		}

		// Handle network changes
		prefix := fmt.Sprint(md.PreviousConfig.Smartnode.ProjectName.Value)
		if md.ChangeNetworks {
			// Remove the checkpoint sync provider
			md.Config.ConsensusCommon.CheckpointSyncProvider.Value = ""
			err = rp.SaveConfig(md.Config)
			if err != nil {
				return fmt.Errorf("error saving config: %w", err)
			}

			color.YellowPrintln("WARNING: You have requested to change networks.")
			fmt.Println()
			color.YellowPrintln("All of your existing chain data, your node wallet, and your validator keys will be removed. If you had a Checkpoint Sync URL provided for your Consensus client, it will be removed and you will need to specify a different one that supports the new network.")
			fmt.Println()
			color.YellowPrintln("Please confirm you have backed up everything you want to keep, because it will be deleted if you answer `y` to the prompt below.")
			fmt.Println()

			if !prompt.Confirm("Would you like the Smart Node to automatically switch networks for you? This will destroy and rebuild your `data` folder and all of Rocket Pool's Docker containers.") {
				fmt.Println("To change networks manually, please follow the steps laid out in the Node Operator's guide (https://docs.rocketpool.net/node-staking/config-docker#choosing-a-network).")
				return nil
			}

			err = changeNetworks(rp, fmt.Sprintf("%s%s", prefix, ApiContainerSuffix), composeFiles)
			if err != nil {
				color.RedPrintln(err.Error())
				fmt.Println("The Smart Node could not automatically change networks for you, so you will have to run the steps manually. Please follow the steps laid out in the Node Operator's guide (https://docs.rocketpool.net/node-staking/mainnet.html).")
			}
			return nil
		}

		// Query for service start if this is a new installation
		if isNew {
			if !prompt.Confirm("Would you like to start the Smart Node services automatically now?") {
				fmt.Println("Please run `rocketpool service start` when you are ready to launch.")
				return nil
			}
			return startService(startServiceParams{
				yes:                    yes,
				ignoreConfigSuggestion: true,
				composeFiles:           composeFiles,
			})
		}

		// Query for service start if this is old and there are containers to change
		if len(md.ContainersToRestart) > 0 {
			fmt.Println("The following containers must be restarted for the changes to take effect:")
			for _, container := range md.ContainersToRestart {
				fmt.Printf("\t%s_%s\n", prefix, container)
			}
			if !prompt.Confirm("Would you like to restart them automatically now?") {
				fmt.Println("Please run `rocketpool service start` when you are ready to apply the changes.")
				return nil
			}

			// Let's reduce potential downtime by pulling the new containers before restarting
			fmt.Println("Pulling potential new container images...")
			err = rp.PullComposeImages(composeFiles)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: couldn't pull new images for updated containers: %s\n", err.Error())
			}

			fmt.Println()
			for _, container := range md.ContainersToRestart {
				fullName := fmt.Sprintf("%s_%s", prefix, container)
				fmt.Printf("Stopping %s... ", fullName)
				rp.StopContainer(fullName)
				fmt.Print("done!\n")
			}

			fmt.Println()
			fmt.Println("Applying changes and restarting containers...")
			return startService(startServiceParams{
				yes:                    yes,
				ignoreConfigSuggestion: true,
				composeFiles:           composeFiles,
			})
		}
	} else {
		fmt.Println("Your changes have not been saved. Your Smart Node configuration is the same as it was before.")
		return nil
	}

	return err
}

// Updates a config parameter from a CLI flag
func updateConfigParamFromCliArg(c *cli.Command, sectionName string, param *cfgtypes.Parameter, cfg *config.RocketPoolConfig) error {

	var paramName string
	if sectionName == "" {
		paramName = param.ID
	} else {
		paramName = fmt.Sprintf("%s-%s", sectionName, param.ID)
	}

	if c.IsSet(paramName) {
		switch param.Type {
		case cfgtypes.ParameterType_Bool:
			param.Value = c.Bool(paramName)
		case cfgtypes.ParameterType_Int:
			param.Value = c.Int(paramName)
		case cfgtypes.ParameterType_Float:
			param.Value = c.Float64(paramName)
		case cfgtypes.ParameterType_String:
			setting := c.String(paramName)
			if param.MaxLength > 0 && len(setting) > param.MaxLength {
				return fmt.Errorf("error setting value for %s: [%s] is too long (max length %d)", paramName, setting, param.MaxLength)
			}
			param.Value = c.String(paramName)
		case cfgtypes.ParameterType_Uint:
			param.Value = c.Uint(paramName)
		case cfgtypes.ParameterType_Uint16:
			param.Value = uint16(c.Uint(paramName))
		case cfgtypes.ParameterType_Choice:
			selection := c.String(paramName)
			found := false
			for _, option := range param.Options {
				if fmt.Sprint(option.Value) == selection {
					param.Value = option.Value
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("error setting value for %s: [%s] is not one of the valid options", paramName, selection)
			}
		}
	}

	return nil

}

// Handle a network change by terminating the service, deleting everything, and starting over
func changeNetworks(rp *rocketpool.Client, apiContainerName string, composeFiles []string) error {

	// Stop all of the containers
	fmt.Println("Stopping containers... ")
	err := rp.PauseService(composeFiles)
	if err != nil {
		return fmt.Errorf("error stopping service: %w", err)
	}
	fmt.Println("done")

	// Restart the API container
	fmt.Println("Starting API container... ")
	output, err := rp.StartContainer(apiContainerName)
	if err != nil {
		return fmt.Errorf("error starting API container: %w", err)
	}
	if output != apiContainerName {
		return fmt.Errorf("starting API container had unexpected output: %s", output)
	}
	fmt.Println("done")

	// Get the path of the user's data folder
	fmt.Println("Retrieving data folder path... ")
	volumePath, err := rp.GetClientVolumeSource(apiContainerName, dataFolderVolumeName)
	if err != nil {
		return fmt.Errorf("error getting data folder path: %w", err)
	}
	fmt.Printf("done, data folder = %s\n", volumePath)

	// Delete the data folder
	fmt.Println("Removing data folder... ")
	_, err = rp.TerminateDataFolder()
	if err != nil {
		return err
	}
	fmt.Println("done")

	// Terminate the current setup
	fmt.Println("Removing old installation... ")
	err = rp.StopService(composeFiles)
	if err != nil {
		return fmt.Errorf("error terminating old installation: %w", err)
	}
	fmt.Println("done")

	// Create new validator folder
	fmt.Println("Recreating data folder... ")
	err = os.MkdirAll(filepath.Join(volumePath, "validators"), 0775)
	if err != nil {
		return fmt.Errorf("error recreating data folder: %w", err)
	}

	// Start the service
	fmt.Println("Starting Rocket Pool... ")
	err = rp.StartService(composeFiles)
	if err != nil {
		return fmt.Errorf("error starting service: %w", err)
	}
	fmt.Println("done")

	return nil

}

type startServiceParams struct {
	yes bool // Whether to automatically confirm prompts
	// N.B.: This should ALYWAYS be false unless --ignore-slash-timer is set!
	ignoreSlashTimer       bool     // Whether to ignore the slash timer
	ignoreConfigSuggestion bool     // Whether to skip suggesting the user run config first
	composeFiles           []string // The compose files to start the service with
}

// Start the Rocket Pool service
func startService(params startServiceParams) error {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Update the Prometheus template with the assigned ports
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading user settings: %w", err)
	}

	// Force all Docker or all Hybrid
	if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local && cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_External {
		color.RedPrintln("You are using a locally-managed Execution client and an externally-managed Consensus client.")
		color.RedPrintln("This configuration is not compatible with The Merge; please select either locally-managed or externally-managed for both the EC and CC.")
	} else if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_External && cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		color.RedPrintln("You are using an externally-managed Execution client and a locally-managed Consensus client.")
		color.RedPrintln("This configuration is not compatible with The Merge; please select either locally-managed or externally-managed for both the EC and CC.")
	}

	if isNew {
		return fmt.Errorf("No configuration detected. Please run `rocketpool service config` to set up your Smart Node before running it.")
	}

	// Check if this is a new install
	isUpdate, err := rp.IsFirstRun()
	if err != nil {
		return fmt.Errorf("error checking for first-run status: %w", err)
	}
	if isUpdate && !params.ignoreConfigSuggestion {
		if params.yes || prompt.Confirm("Smart Node upgrade detected - starting will overwrite certain settings with the latest defaults (such as container versions).\nYou may want to run `service config` first to see what's changed.\n\nWould you like to continue starting the service?") {
			err = cfg.UpdateDefaults()
			if err != nil {
				return fmt.Errorf("error upgrading configuration with the latest parameters: %w", err)
			}
			rp.SaveConfig(cfg)
			color.GreenPrintln("Updated settings successfully.")
		} else {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Update the Prometheus template with the assigned ports
	metricsEnabled := cfg.EnableMetrics.Value.(bool)
	if metricsEnabled {
		err := rp.UpdatePrometheusConfiguration(cfg)
		if err != nil {
			return err
		}
	}

	// Update the Alertmanager configuration files even if metrics is disabled; as smartnode sends some alerts directly
	alertingEnabled := cfg.Alertmanager.EnableAlerting.Value.(bool)
	if alertingEnabled {
		err = cfg.Alertmanager.UpdateConfigurationFiles(rp.ConfigPath())
		if err != nil {
			return err
		}
	}

	// Validate the config
	errors := cfg.Validate()
	if len(errors) > 0 {
		color.RedPrintln("Your configuration encountered errors. You must correct the following in order to start Rocket Pool:")
		fmt.Println()
		for _, err := range errors {
			color.RedPrintf("%s\n", err)
			fmt.Println()
		}
		return nil
	}

	if !params.ignoreSlashTimer {
		// Do the client swap check
		err := checkForValidatorChange(rp, cfg)
		if err != nil {
			color.YellowPrintln("WARNING: couldn't verify that the validator container can be safely restarted:")
			color.YellowPrintf("\t%s\n", err.Error())
			color.YellowPrintln("If you are changing to a different ETH2 client, it may resubmit an attestation you have already submitted.")
			color.YellowPrintln("This will slash your validator!")
			color.YellowPrintln("To prevent slashing, you must wait 15 minutes from the time you stopped the clients before starting them again.")
			fmt.Println()
			color.YellowPrintln("**If you did NOT change clients, you can safely ignore this warning.**")
			fmt.Println()
			if !prompt.ConfirmYellow("Press y when you understand the above warning, have waited, and are ready to start Rocket Pool:") {
				fmt.Println("Cancelled.")
				return nil
			}
		}
	} else {
		color.YellowPrintln("Ignoring anti-slashing safety delay.")
	}

	// Write a note on doppelganger protection
	doppelgangerEnabled, err := cfg.IsDoppelgangerEnabled()
	if err != nil {
		color.YellowPrintf("Couldn't check if you have Doppelganger Protection enabled: %s\n", err.Error())
		color.YellowPrintln("If you do, your validator will miss up to 3 attestations when it starts.")
		color.YellowPrintln("This is *intentional* and does not indicate a problem with your node.")
	} else if doppelgangerEnabled {
		color.YellowPrintln("NOTE: You currently have Doppelganger Protection enabled.")
		color.YellowPrintln("Your validator will miss up to 3 attestations when it starts.")
		color.YellowPrintln("This is *intentional* and does not indicate a problem with your node.")
	}
	fmt.Println()

	// Start service
	err = rp.StartService(params.composeFiles)
	if err != nil {
		return err
	}

	// Remove the upgrade flag if it's there
	return rp.RemoveUpgradeFlagFile()

}

func checkForValidatorChange(rp *rocketpool.Client, cfg *config.RocketPoolConfig) error {

	// Get the container prefix
	prefix, err := rp.GetContainerPrefix()
	if err != nil {
		return fmt.Errorf("Error getting validator container prefix: %w", err)
	}

	// Get the current validator client
	currentValidatorImageString, err := rp.GetDockerImage(prefix + ValidatorContainerSuffix)
	if err != nil {
		return fmt.Errorf("Error getting current validator image: %w", err)
	}

	currentValidatorName, err := getDockerImageName(currentValidatorImageString)
	if err != nil {
		return fmt.Errorf("Error getting current validator image name: %w", err)
	}

	// Get the new validator client according to the settings file
	selectedConsensusClientConfig, err := cfg.GetSelectedConsensusClientConfig()
	if err != nil {
		return fmt.Errorf("Error getting selected consensus client config: %w", err)
	}
	pendingValidatorName, err := getDockerImageName(selectedConsensusClientConfig.GetValidatorImage())
	if err != nil {
		return fmt.Errorf("Error getting pending validator image name: %w", err)
	}

	// Compare the clients and warn if necessary
	if currentValidatorName == pendingValidatorName {
		fmt.Printf("Validator client [%s] was previously used - no slashing prevention delay necessary.\n", currentValidatorName)
	} else if currentValidatorName == "" {
		fmt.Println("This is the first time starting Rocket Pool - no slashing prevention delay necessary.")
	} else {

		consensusClient, _ := cfg.GetSelectedConsensusClient()
		// Warn about Lodestar
		if consensusClient == cfgtypes.ConsensusClient_Lodestar {
			color.YellowPrintln("NOTE:")
			color.YellowPrintln("If this is your first time running Lodestar and you have existing minipools, you must run `rocketpool wallet rebuild` after the Smart Node starts to generate the validator keys for it.")
			color.YellowPrintln("If you have run it before or you don't have any minipools, you can ignore this message.")
			fmt.Println()
		}

		// Get the time that the container responsible for validator duties exited
		validatorDutyContainerName, err := getContainerNameForValidatorDuties(currentValidatorName, rp)
		if err != nil {
			return fmt.Errorf("Error getting validator container name: %w", err)
		}
		validatorFinishTime, err := rp.GetDockerContainerShutdownTime(validatorDutyContainerName)
		if err != nil {
			return fmt.Errorf("Error getting validator shutdown time: %w", err)
		}

		// If it hasn't exited yet, shut it down
		zeroTime := time.Time{}
		status, err := rp.GetDockerStatus(validatorDutyContainerName)
		if err != nil {
			return fmt.Errorf("Error getting container [%s] status: %w", validatorDutyContainerName, err)
		}
		if validatorFinishTime == zeroTime || status == "running" {
			color.YellowPrintln("Validator is currently running, stopping it...")
			response, err := rp.StopContainer(validatorDutyContainerName)
			validatorFinishTime = time.Now()
			if err != nil {
				return fmt.Errorf("Error stopping container [%s]: %w", validatorDutyContainerName, err)
			}
			if response != validatorDutyContainerName {
				return fmt.Errorf("Unexpected response when stopping container [%s]: %s", validatorDutyContainerName, response)
			}
		}

		// Print the warning and start the time lockout
		safeStartTime := validatorFinishTime.Add(15 * time.Minute)
		remainingTime := time.Until(safeStartTime)
		if remainingTime <= 0 {
			fmt.Printf("The validator has been offline for %s, which is long enough to prevent slashing.\n", time.Since(validatorFinishTime))
			fmt.Println("The new client can be safely started.")
		} else {
			color.RedPrintln("=== WARNING ===")
			color.RedPrintf("You have changed your validator client from %s to %s. Only %s has elapsed since you stopped %s.\n", currentValidatorName, pendingValidatorName, time.Since(validatorFinishTime), currentValidatorName)
			color.RedPrintf("If you were actively validating while using %s, starting %s without waiting will cause your validators to be slashed due to duplicate attestations!", currentValidatorName, pendingValidatorName)
			color.RedPrintln("To prevent slashing, Rocket Pool will delay activating the new client for 15 minutes.")
			color.RedPrintln("See the documentation for a more detailed explanation: https://docs.rocketpool.net/node-staking/maintenance/node-migration.html#slashing-and-the-slashing-database")
			color.RedPrintln("If you have read the documentation, understand the risks, and want to bypass this cooldown, run `rocketpool service start --ignore-slash-timer`.")
			fmt.Println()

			// Wait for 15 minutes
			for remainingTime > 0 {
				fmt.Printf("Remaining time: %s", remainingTime)
				time.Sleep(1 * time.Second)
				remainingTime = time.Until(safeStartTime)
				fmt.Printf("%s\r", clearLine)
			}

			fmt.Println("You may now safely start the validator without fear of being slashed.")
		}
	}

	return nil
}

// Get the name of the container responsible for validator duties based on the client name
func getContainerNameForValidatorDuties(CurrentValidatorClientName string, rp *rocketpool.Client) (string, error) {

	prefix, err := rp.GetContainerPrefix()
	if err != nil {
		return "", err
	}

	return prefix + ValidatorContainerSuffix, nil

}

// Get the time that the container responsible for validator duties exited
func getValidatorFinishTime(CurrentValidatorClientName string, rp *rocketpool.Client) (time.Time, error) {

	prefix, err := rp.GetContainerPrefix()
	if err != nil {
		return time.Time{}, err
	}

	var validatorFinishTime time.Time
	if CurrentValidatorClientName == "nimbus" {
		validatorFinishTime, err = rp.GetDockerContainerShutdownTime(prefix + BeaconContainerSuffix)
	} else {
		validatorFinishTime, err = rp.GetDockerContainerShutdownTime(prefix + ValidatorContainerSuffix)
	}

	return validatorFinishTime, err

}

// Extract the image name from a Docker image string
func getDockerImageName(imageString string) (string, error) {

	// Return the empty string if the validator didn't exist (probably because this is the first time starting it up)
	if imageString == "" {
		return "", nil
	}

	reg := regexp.MustCompile(dockerImageRegex)
	matches := reg.FindStringSubmatch(imageString)
	if matches == nil {
		return "", fmt.Errorf("Couldn't parse the Docker image string [%s]", imageString)
	}
	imageIndex := reg.SubexpIndex("image")
	if imageIndex == -1 {
		return "", fmt.Errorf("Image name not found in Docker image [%s]", imageString)
	}

	imageName := matches[imageIndex]
	return imageName, nil
}

// Prepares the execution client for pruning
func pruneExecutionClient(yes bool) error {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return err
	}
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
	}

	// Sanity checks
	if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_External {
		fmt.Println("You are using an externally managed Execution client.\nThe Smart Node cannot prune it for you.")
		return nil
	}
	if cfg.IsNativeMode {
		fmt.Println("You are using Native Mode.\nThe Smart Node cannot prune your Execution client for you, you'll have to do it manually.")
	}
	selectedEc := cfg.ExecutionClient.Value.(cfgtypes.ExecutionClient)

	// Don't prune if the EC is in archive mode
	if cfg.ExecutionCommon.PruningMode.Value == cfgtypes.PruningMode_Archive {
		fmt.Println("Your Execution Client is being used as an archive node.\nArchive nodes should not be pruned. Aborting.")
		return nil
	}

	// Print the appropriate warnings before pruning
	if selectedEc == cfgtypes.ExecutionClient_Geth {
		color.YellowPrintln("Geth has a new feature that renders pruning obsolete. However, as this is a new feature you may have to resync with `rocketpool service resync-eth1` before this takes effect.")
		fmt.Println("This will shut down your main execution client and prune its database, freeing up disk space.")
		if cfg.UseFallbackClients.Value == false {
			color.RedPrintln("You do not have a fallback execution client configured.")
			color.RedPrintln("Your node will no longer be able to perform any validation duties (attesting or proposing blocks) until pruning is done.")
			color.RedPrintln("Please configure a fallback client with `rocketpool service config` before running this.")
		} else {
			fmt.Println("You have fallback clients enabled. Rocket Pool (and your consensus client) will use that while the main client is pruning.")
		}
		fmt.Println("Once pruning is complete, your execution client will restart automatically.")
	} else {
		fmt.Println("This will request your main execution client to prune its database, freeing up disk space. This is a resource intensive operation and may lead to an increase in missed attestations until it finishes.")
	}
	fmt.Println()

	// Get the container prefix
	prefix, err := rp.GetContainerPrefix()
	if err != nil {
		return fmt.Errorf("Error getting container prefix: %w", err)
	}

	// Prompt for confirmation
	if !(yes || prompt.Confirm("Are you sure you want to prune your main execution client?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get the execution container name
	executionContainerName := prefix + ExecutionContainerSuffix

	// Check for enough free space
	volumePath, err := rp.GetClientVolumeSource(executionContainerName, clientDataVolumeName)
	if err != nil {
		return fmt.Errorf("Error getting execution volume source path: %w", err)
	}
	partitions, err := disk.Partitions(true)
	if err != nil {
		return fmt.Errorf("Error getting partition list: %w", err)
	}

	longestPath := 0
	bestPartition := disk.PartitionStat{}
	for _, partition := range partitions {
		if strings.HasPrefix(volumePath, partition.Mountpoint) && len(partition.Mountpoint) > longestPath {
			bestPartition = partition
			longestPath = len(partition.Mountpoint)
		}
	}

	diskUsage, err := disk.Usage(bestPartition.Mountpoint)
	if err != nil {
		return fmt.Errorf("Error getting free disk space available: %w", err)
	}
	freeSpaceHuman := humanize.IBytes(diskUsage.Free)
	pruneFreeSpaceRequired := PruneFreeSpaceRequired
	if cfg.GetNetwork() == cfgtypes.Network_Mainnet && selectedEc == cfgtypes.ExecutionClient_Nethermind {
		pruneFreeSpaceRequired = NethermindPruneFreeSpaceRequired
	}
	if diskUsage.Free < pruneFreeSpaceRequired {
		return fmt.Errorf("Your disk must have %s GiB free to prune, but it only has %s free. Please free some space before pruning.", humanize.IBytes(pruneFreeSpaceRequired), freeSpaceHuman)
	}

	fmt.Printf("Your disk has %s free, which is enough to prune.\n", freeSpaceHuman)

	if selectedEc == cfgtypes.ExecutionClient_Nethermind {
		// Restarting NM is not needed anymore
		err = rp.RunNethermindPruneStarter(executionContainerName)
		if err != nil {
			return fmt.Errorf("Error starting Nethermind prune starter: %w", err)
		}
		return nil
	}
	fmt.Printf("Stopping %s...\n", executionContainerName)
	result, err := rp.StopContainer(executionContainerName)
	if err != nil {
		return fmt.Errorf("Error stopping main execution container: %w", err)
	}
	if result != executionContainerName {
		return fmt.Errorf("Unexpected output while stopping main execution container: %s", result)
	}

	// Get the ETH1 volume name
	volume, err := rp.GetClientVolumeName(executionContainerName, clientDataVolumeName)
	if err != nil {
		return fmt.Errorf("Error getting execution client volume name: %w", err)
	}

	// Run the prune provisioner
	fmt.Printf("Provisioning pruning on volume %s...\n", volume)
	err = rp.RunPruneProvisioner(prefix+PruneProvisionerContainerSuffix, volume)
	if err != nil {
		return fmt.Errorf("Error running prune provisioner: %w", err)
	}

	// Restart ETH1
	fmt.Printf("Restarting %s...\n", executionContainerName)
	result, err = rp.StartContainer(executionContainerName)
	if err != nil {
		return fmt.Errorf("Error starting main execution client: %w", err)
	}
	if result != executionContainerName {
		return fmt.Errorf("Unexpected output while starting main execution client: %s", result)
	}

	fmt.Println()
	fmt.Println("Done! Your main execution client is now pruning. You can follow its progress with `rocketpool service logs eth1`.")
	fmt.Println("Once it's done, it will restart automatically and resume normal operation.")

	color.YellowPrintln("NOTE: While pruning, you **cannot** interrupt the client (e.g. by restarting) or you risk corrupting the database!")
	color.YellowPrintln("You must let it run to completion!")

	return nil

}

// Stops Smart Node stack containers, prunes docker, and restarts the Smart Node stack.
func resetDocker(yes, all bool, composeFiles []string) error {

	fmt.Println("Once cleanup is complete, Rocket Pool will restart automatically.")
	fmt.Println()

	// Stop...
	// NOTE: pauseService prompts for confirmation, so we don't need to do it here
	confirmed, err := pauseService(yes, composeFiles)
	if err != nil {
		return err
	}

	if !confirmed {
		// if the user cancelled the pause, then we cancel the rest of the operation here:
		return nil
	}

	// Prune images...
	err = pruneDocker(all, composeFiles)
	if err != nil {
		return fmt.Errorf("error pruning Docker: %s", err)
	}

	// Restart...
	// NOTE: startService does some other sanity checks and messages that we leverage here:
	fmt.Println("Restarting Rocket Pool...")
	err = startService(startServiceParams{
		yes:                    yes,
		ignoreConfigSuggestion: true,
		composeFiles:           composeFiles,
	})
	if err != nil {
		return fmt.Errorf("error starting Rocket Pool: %s", err)
	}
	return nil
}

func pruneDocker(deleteAllImages bool, composeFiles []string) error {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// NOTE: we deliberately avoid using `docker system prune -a` and delete all
	//   images manually so that we can preserve the current smartnode-stack
	//   images, _unless_ the user specified --all option
	if !deleteAllImages {
		ourImages, err := rp.GetComposeImages(composeFiles)
		if err != nil {
			return fmt.Errorf("error getting compose images: %w", err)
		}

		ourImagesMap := make(map[string]struct{})
		for _, image := range ourImages {
			ourImagesMap[image] = struct{}{}
		}

		allImages, err := rp.GetAllDockerImages()
		if err != nil {
			return fmt.Errorf("error getting all docker images: %w", err)
		}

		fmt.Println("Deleting images not used by the Rocket Pool Smart Node...")
		for _, image := range allImages {
			if _, ok := ourImagesMap[image.TagString()]; !ok {
				fmt.Printf("Deleting %s...\n", image.String())
				_, err = rp.DeleteDockerImage(image.ID)
				if err != nil {
					// safe to ignore and print to user, since it may just be an image referenced by a running container that is managed outside of the smartnode's compose stack
					fmt.Printf("Error deleting image %s: %s\n", image.String(), err.Error())
				}
				continue
			}

			fmt.Printf("Skipping image used by Smart Node stack: %s\n", image.String())
		}
	}

	// now we can run docker system prune (potentially without --all) to remove
	// all stopped containers and networks:
	fmt.Println("Pruning Docker system...")
	err := rp.DockerSystemPrune(deleteAllImages)
	if err != nil {
		return fmt.Errorf("error pruning Docker system: %w", err)
	}

	return nil
}

// Pause the Rocket Pool service. Returns whether the action proceeded (was confirmed by user and no error occurred before starting it)
func pauseService(yes bool, composeFiles []string) (bool, error) {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Get the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return false, err
	}

	// Write a note on doppelganger protection
	doppelgangerEnabled, err := cfg.IsDoppelgangerEnabled()
	if err != nil {
		color.YellowPrintf("Couldn't check if you have Doppelganger Protection enabled: %s\n", err.Error())
		color.YellowPrintln("If you do, stopping your validator will cause it to miss up to 3 attestations when it next starts.")
		color.YellowPrintln("This is *intentional* and does not indicate a problem with your node.")
	} else if doppelgangerEnabled {
		color.YellowPrintln("NOTE: You currently have Doppelganger Protection enabled.")
		color.YellowPrintln("If you stop your validator, it will miss up to 3 attestations when it next starts.")
		color.YellowPrintln("This is *intentional* and does not indicate a problem with your node.")
	}
	fmt.Println()

	// Prompt for confirmation
	if !(yes || prompt.Confirm("Are you sure you want to pause the Rocket Pool service? Any staking minipools and megapool validators will be penalized!")) {
		fmt.Println("Cancelled.")
		return false, nil
	}

	// Pause service
	err = rp.PauseService(composeFiles)
	return true, err

}

// Terminate the Rocket Pool service
func terminateService(yes bool, composeFiles []string, configPath string) error {

	// Prompt for confirmation
	if !(yes || prompt.ConfirmRed("WARNING: Are you sure you want to terminate the Rocket Pool service? Any staking minipools will be penalized, your ETH1 and ETH2 chain databases will be deleted, you will lose ALL of your sync progress, and you will lose your Prometheus metrics database!\nAfter doing this, you will have to **reinstall** the Smart Node uses `rocketpool service install -d` in order to use it again.")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Stop service
	return rp.TerminateService(composeFiles, configPath)

}

// View the Rocket Pool service logs
func serviceLogs(tail string, composeFiles []string, aliasedNames ...string) error {

	// Handle name aliasing
	serviceNames := []string{}
	for _, name := range aliasedNames {
		trueName := name
		switch name {
		case "ec", "el", "execution":
			trueName = "eth1"
		case "cc", "cl", "bc", "bn", "beacon", "consensus":
			trueName = "eth2"
		case "vc":
			trueName = "validator"
		}
		serviceNames = append(serviceNames, trueName)
	}

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Print service logs
	return rp.PrintServiceLogs(composeFiles, tail, serviceNames...)

}

// View the Rocket Pool service compose config
func serviceCompose(composeFiles []string) error {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Print service compose config
	return rp.PrintServiceCompose(composeFiles)

}

// View the Rocket Pool service version information
func serviceVersion() error {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = cliutils.PrintNetwork(cfg.GetNetwork(), isNew)
	if err != nil {
		return err
	}

	// Get RP service version
	serviceVersion, err := rp.GetServiceVersion()
	if err != nil {
		return err
	}

	// Handle native mode
	if cfg.IsNativeMode {
		fmt.Printf("Rocket Pool client version: %s\n", shared.RocketPoolVersion())
		fmt.Printf("Rocket Pool service version: %s\n", serviceVersion)
		fmt.Println("Configured for Native Mode")
		return nil
	}

	// Get the execution client string
	var eth1ClientString string
	eth1ClientMode := cfg.ExecutionClientMode.Value.(cfgtypes.Mode)
	switch eth1ClientMode {
	case cfgtypes.Mode_Local:
		eth1Client := cfg.ExecutionClient.Value.(cfgtypes.ExecutionClient)
		format := "%s (Locally managed)\n\tImage: %s"
		switch eth1Client {
		case cfgtypes.ExecutionClient_Geth:
			eth1ClientString = fmt.Sprintf(format, "Geth", cfg.Geth.ContainerTag.Value.(string))
		case cfgtypes.ExecutionClient_Nethermind:
			eth1ClientString = fmt.Sprintf(format, "Nethermind", cfg.Nethermind.ContainerTag.Value.(string))
		case cfgtypes.ExecutionClient_Besu:
			eth1ClientString = fmt.Sprintf(format, "Besu", cfg.Besu.ContainerTag.Value.(string))
		case cfgtypes.ExecutionClient_Reth:
			eth1ClientString = fmt.Sprintf(format, "Reth", cfg.Reth.ContainerTag.Value.(string))
		default:
			return fmt.Errorf("unknown local execution client [%v]", eth1Client)
		}

	case cfgtypes.Mode_External:
		eth1ClientString = "Externally managed"

	default:
		return fmt.Errorf("unknown execution client mode [%v]", eth1ClientMode)
	}

	// Get the consensus client string
	var eth2ClientString string
	eth2ClientMode := cfg.ConsensusClientMode.Value.(cfgtypes.Mode)
	switch eth2ClientMode {
	case cfgtypes.Mode_Local:
		eth2Client := cfg.ConsensusClient.Value.(cfgtypes.ConsensusClient)
		format := "%s (Locally managed)\n\tImage: %s"
		switch eth2Client {
		case cfgtypes.ConsensusClient_Lighthouse:
			eth2ClientString = fmt.Sprintf(format, "Lighthouse", cfg.Lighthouse.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Lodestar:
			eth2ClientString = fmt.Sprintf(format, "Lodestar", cfg.Lodestar.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Nimbus:
			eth2ClientString = fmt.Sprintf(format+"\n\tVC image: %s", "Nimbus", cfg.Nimbus.BnContainerTag.Value.(string), cfg.Nimbus.VcContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Prysm:
			eth2ClientString = fmt.Sprintf(format+"\n\tVC image: %s", "Prysm", cfg.Prysm.BnContainerTag.Value.(string), cfg.Prysm.VcContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Teku:
			eth2ClientString = fmt.Sprintf(format, "Teku", cfg.Teku.ContainerTag.Value.(string))
		default:
			return fmt.Errorf("unknown local consensus client [%v]", eth2Client)
		}

	case cfgtypes.Mode_External:
		eth2Client := cfg.ExternalConsensusClient.Value.(cfgtypes.ConsensusClient)
		format := "%s (Externally managed)\n\tVC Image: %s"
		switch eth2Client {
		case cfgtypes.ConsensusClient_Lighthouse:
			eth2ClientString = fmt.Sprintf(format, "Lighthouse", cfg.ExternalLighthouse.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Lodestar:
			eth2ClientString = fmt.Sprintf(format, "Lodestar", cfg.ExternalLodestar.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Nimbus:
			eth2ClientString = fmt.Sprintf(format, "Nimbus", cfg.ExternalNimbus.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Prysm:
			eth2ClientString = fmt.Sprintf(format, "Prysm", cfg.ExternalPrysm.ContainerTag.Value.(string))
		case cfgtypes.ConsensusClient_Teku:
			eth2ClientString = fmt.Sprintf(format, "Teku", cfg.ExternalTeku.ContainerTag.Value.(string))
		default:
			return fmt.Errorf("unknown external consensus client [%v]", eth2Client)
		}

	default:
		return fmt.Errorf("unknown consensus client mode [%v]", eth2ClientMode)
	}

	var mevBoostString string
	if cfg.EnableMevBoost.Value.(bool) {
		if cfg.MevBoost.Mode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
			mevBoostString = fmt.Sprintf("Enabled (Local Mode)\n\tImage: %s", cfg.MevBoost.ContainerTag.Value.(string))
		} else {
			mevBoostString = "Enabled (External Mode)"
		}
	} else {
		mevBoostString = "Disabled"
	}

	var commitBoostString string
	if cfg.EnableCommitBoost.Value.(bool) {
		if cfg.CommitBoost.Mode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
			commitBoostString = fmt.Sprintf("Enabled (Local Mode)\n\tImage: %s", cfg.CommitBoost.ContainerTag.Value.(string))
		} else {
			commitBoostString = "Enabled (External Mode)"
		}
	} else {
		commitBoostString = "Disabled"
	}

	// Print version info
	fmt.Printf("Rocket Pool client version: %s\n", shared.RocketPoolVersion())
	fmt.Printf("Rocket Pool service version: %s\n", serviceVersion)
	fmt.Printf("Selected Eth 1.0 client: %s\n", eth1ClientString)
	fmt.Printf("Selected Eth 2.0 client: %s\n", eth2ClientString)
	fmt.Printf("MEV-Boost client: %s\n", mevBoostString)
	fmt.Printf("Commit-Boost client: %s\n", commitBoostString)
	return nil

}

// Destroy and resync the eth1 client from scratch
func resyncEth1(yes bool, composeFiles []string) error {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Get the config
	_, isNew, err := rp.LoadConfig()
	if err != nil {
		return err
	}
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
	}

	fmt.Println("This will delete the chain data of your primary ETH1 client and resync it from scratch.")
	color.YellowPrintln("You should only do this if your ETH1 client has failed and can no longer start or sync properly.")
	color.YellowPrintln("This is meant to be a last resort.")

	// Get the container prefix
	prefix, err := rp.GetContainerPrefix()
	if err != nil {
		return fmt.Errorf("Error getting container prefix: %w", err)
	}

	// Prompt for confirmation
	if !(yes || prompt.ConfirmRed("Are you SURE you want to delete and resync your main ETH1 client from scratch? This cannot be undone!")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stop ETH1
	executionContainerName := prefix + ExecutionContainerSuffix
	fmt.Printf("Stopping %s...\n", executionContainerName)
	result, err := rp.StopContainer(executionContainerName)
	if err != nil {
		color.YellowPrintf("WARNING: Stopping main ETH1 container failed: %s\n", err.Error())
	}
	if result != executionContainerName {
		color.YellowPrintf("WARNING: Unexpected output while stopping main ETH1 container: %s\n", result)
	}

	// Get ETH1 volume name
	volume, err := rp.GetClientVolumeName(executionContainerName, clientDataVolumeName)
	if err != nil {
		return fmt.Errorf("Error getting ETH1 volume name: %w", err)
	}

	// Remove ETH1
	fmt.Printf("Deleting %s...\n", executionContainerName)
	result, err = rp.RemoveContainer(executionContainerName)
	if err != nil {
		return fmt.Errorf("Error deleting main ETH1 container: %w", err)
	}
	if result != executionContainerName {
		return fmt.Errorf("Unexpected output while deleting main ETH1 container: %s", result)
	}

	// Delete the ETH1 volume
	fmt.Printf("Deleting volume %s...\n", volume)
	result, err = rp.DeleteVolume(volume)
	if err != nil {
		return fmt.Errorf("Error deleting volume: %w", err)
	}
	if result != volume {
		return fmt.Errorf("Unexpected output while deleting volume: %s", result)
	}

	// Restart Rocket Pool
	fmt.Printf("Rebuilding %s and restarting Rocket Pool...\n", executionContainerName)
	err = startService(startServiceParams{
		yes:                    yes,
		ignoreConfigSuggestion: true,
		composeFiles:           composeFiles,
	})
	if err != nil {
		return fmt.Errorf("Error starting Rocket Pool: %s", err)
	}

	fmt.Printf("\nDone! Your main ETH1 client is now resyncing. You can follow its progress with `rocketpool service logs eth1`.\n")

	return nil

}

// Destroy and resync the eth2 client from scratch
func resyncEth2(yes bool, composeFiles []string) error {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	// Get the merged config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return err
	}
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
	}

	fmt.Println("This will delete the chain data of your ETH2 client and resync it from scratch.")
	color.YellowPrintln("You should only do this if your ETH2 client has failed and can no longer start or sync properly.")
	color.YellowPrintln("This is meant to be a last resort.")
	fmt.Println()

	// Get the parameters that the selected client doesn't support
	var unsupportedParams []string
	var clientName string
	eth2ClientMode := cfg.ConsensusClientMode.Value.(cfgtypes.Mode)
	switch eth2ClientMode {
	case cfgtypes.Mode_Local:
		selectedClientConfig, err := cfg.GetSelectedConsensusClientConfig()
		if err != nil {
			return fmt.Errorf("error getting selected consensus client config: %w", err)
		}
		unsupportedParams = selectedClientConfig.(cfgtypes.LocalConsensusConfig).GetUnsupportedCommonParams()
		clientName = selectedClientConfig.GetName()

	case cfgtypes.Mode_External:
		fmt.Println("You use an externally-managed Consensus client. Rocket Pool cannot resync it for you.")
		return nil

	default:
		return fmt.Errorf("unknown consensus client mode [%v]", eth2ClientMode)
	}

	// Check if the selected client supports checkpoint sync
	supportsCheckpointSync := true
	for _, param := range unsupportedParams {
		if param == config.CheckpointSyncUrlID {
			supportsCheckpointSync = false
		}
	}
	if !supportsCheckpointSync {
		color.RedPrintln("Your ETH2 client (%s) does not support checkpoint sync.", clientName)
		color.RedPrintln("If you have active validators, they", color.Bold("will be considered offline and will leak ETH"), "while the client is syncing.")
		fmt.Println()
	} else {
		// Get the current checkpoint sync URL
		checkpointSyncUrl := cfg.ConsensusCommon.CheckpointSyncProvider.Value.(string)
		if checkpointSyncUrl == "" {
			color.RedPrintln("You do not have a checkpoint sync provider configured.")
			color.RedPrintln("If you have active validators, they", color.Bold("will be considered offline and will leak ETH"), "while the client is syncing.")
			color.RedPrintln("We strongly recommend you configure a checkpoint sync provider with `rocketpool service config` so it syncs instantly before running this.")
			fmt.Println()
		} else {
			fmt.Printf("You have a checkpoint sync provider configured (%s).\n", checkpointSyncUrl)
			fmt.Println("Your ETH2 client will use it to sync to the head of the Beacon Chain instantly after being rebuilt.")
		}
	}

	// Prompt for confirmation
	if !(yes || prompt.ConfirmRed("Are you SURE you want to delete and resync your ETH2 client from scratch? This cannot be undone!")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get the container prefix
	prefix, err := rp.GetContainerPrefix()
	if err != nil {
		return fmt.Errorf("Error getting container prefix: %w", err)
	}

	beaconContainerName := prefix + BeaconContainerSuffix

	// Find running containers using the ETH2 volume
	containers, err := rp.GetContainersByPrefix(prefix)
	if err != nil {
		return fmt.Errorf("Error getting containers by prefix: %w", err)
	}

	// Get ETH2 volume name
	volume, err := rp.GetClientVolumeName(beaconContainerName, clientDataVolumeName)
	if err != nil {
		return fmt.Errorf("Error getting ETH2 volume name: %w", err)
	}

	// Stop and delete the containers if they are running
	for _, container := range containers {

		// Ignore containers that don't have the ETH2 volume
		if !container.HasVolume(volume) {
			continue
		}

		fmt.Println(container.Names, container.State)
		if container.State != "exited" {
			fmt.Printf("Stopping %s...\n", container.Names)
			result, err := rp.StopContainer(container.Names)
			if err != nil {
				color.YellowPrintf("WARNING: Stopping container %s failed: %s\n", container.Names, err.Error())
			}
			if result != container.Names {
				color.YellowPrintf("WARNING: Unexpected output while stopping container %s: %s\n", container.Names, result)
			}
		}

		fmt.Printf("Deleting %s...\n", container.Names)
		result, err := rp.RemoveContainer(container.Names)
		if err != nil {
			color.YellowPrintf("WARNING: Deleting container %s failed: %s\n", container.Names, err.Error())
		}
		if result != container.Names {
			color.YellowPrintf("WARNING: Unexpected output while deleting container %s: %s\n", container.Names, result)
		}
	}

	// Delete the ETH2 volume
	fmt.Printf("Deleting volume %s...\n", volume)
	result, err := rp.DeleteVolume(volume)
	if err != nil {
		return fmt.Errorf("Error deleting volume: %w", err)
	}
	if result != volume {
		return fmt.Errorf("Unexpected output while deleting volume: %s", result)
	}

	// Restart Rocket Pool
	fmt.Printf("Rebuilding %s and restarting Rocket Pool...\n", beaconContainerName)
	err = startService(startServiceParams{
		yes:                    yes,
		ignoreConfigSuggestion: true,
		composeFiles:           composeFiles,
	})
	if err != nil {
		return fmt.Errorf("Error starting Rocket Pool: %s", err)
	}

	fmt.Printf("\nDone! Your ETH2 client is now resyncing. You can follow its progress with `rocketpool service logs eth2`.\n")

	return nil

}

// Generate a YAML file that shows the current configuration schema, including all of the parameters and their descriptions
func getConfigYaml() error {
	cfg := config.NewRocketPoolConfig("", false)
	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error serializing configuration schema: %w", err)
	}

	fmt.Println(string(bytes))
	return nil
}

// Get the amount of space used by a Docker volume
func getVolumeSpaceUsed(rp *rocketpool.Client, volume string) (uint64, error) {
	size, err := rp.GetVolumeSize(volume)
	if err != nil {
		return 0, fmt.Errorf("error getting execution client volume name: %w", err)
	}
	volumeBytes, err := humanize.ParseBytes(size)
	if err != nil {
		return 0, fmt.Errorf("couldn't parse size of EC volume (%s): %w", size, err)
	}
	return volumeBytes, nil
}

// Get the amount of free space available in the target dir
func getPartitionFreeSpace(rp *rocketpool.Client, targetDir string) (uint64, error) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return 0, fmt.Errorf("error getting partition list: %w", err)
	}
	longestPath := 0
	bestPartition := disk.PartitionStat{}
	for _, partition := range partitions {
		if strings.HasPrefix(targetDir, partition.Mountpoint) && len(partition.Mountpoint) > longestPath {
			bestPartition = partition
			longestPath = len(partition.Mountpoint)
		}
	}
	diskUsage, err := disk.Usage(bestPartition.Mountpoint)
	if err != nil {
		return 0, fmt.Errorf("error getting free disk space available: %w", err)
	}
	return diskUsage.Free, nil
}
