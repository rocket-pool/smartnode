package service

import (
	"fmt"
	"regexp"
	"time"

	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/shared/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/urfave/cli/v2"
)

// Start the Rocket Pool service
func startService(c *cli.Context, ignoreConfigSuggestion bool) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Update the Prometheus template with the assigned ports
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading user settings: %w", err)
	}

	// Force all Docker or all Hybrid
	if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local && cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_External {
		fmt.Printf("%sYou are using a locally-managed Execution client and an externally-managed Consensus client.\nThis configuration is not compatible with The Merge; please select either locally-managed or externally-managed for both the EC and CC.%s\n", terminal.ColorRed, terminal.ColorReset)
	} else if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_External && cfg.ConsensusClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_Local {
		fmt.Printf("%sYou are using an externally-managed Execution client and a locally-managed Consensus client.\nThis configuration is not compatible with The Merge; please select either locally-managed or externally-managed for both the EC and CC.%s\n", terminal.ColorRed, terminal.ColorReset)
	}

	if isNew {
		return fmt.Errorf("No configuration detected. Please run `rocketpool service config` to set up your Smartnode before running it.")
	}

	// Check if this is a new install
	isUpdate, err := rp.IsFirstRun()
	if err != nil {
		return fmt.Errorf("error checking for first-run status: %w", err)
	}
	if isUpdate && !ignoreConfigSuggestion {
		if c.Bool("yes") || utils.Confirm("Smartnode upgrade detected - starting will overwrite certain settings with the latest defaults (such as container versions).\nYou may want to run `service config` first to see what's changed.\n\nWould you like to continue starting the service?") {
			err = cfg.UpdateDefaults()
			if err != nil {
				return fmt.Errorf("error upgrading configuration with the latest parameters: %w", err)
			}
			rp.SaveConfig(cfg)
			fmt.Printf("%sUpdated settings successfully.%s\n", terminal.ColorGreen, terminal.ColorReset)
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

	// Validate the config
	errors := cfg.Validate()
	if len(errors) > 0 {
		fmt.Printf("%sYour configuration encountered errors. You must correct the following in order to start Rocket Pool:\n\n", terminal.ColorRed)
		for _, err := range errors {
			fmt.Printf("%s\n\n", err)
		}
		fmt.Println(terminal.ColorReset)
		return nil
	}

	if !c.Bool(ignoreSlashTimerFlag.Name) {
		// Do the client swap check
		err := checkForValidatorChange(rp, cfg)
		if err != nil {
			fmt.Printf("%sWARNING: couldn't verify that the validator container can be safely restarted:\n\t%s\n", terminal.ColorYellow, err.Error())
			fmt.Println("If you are changing to a different ETH2 client, it may resubmit an attestation you have already submitted.")
			fmt.Println("This will slash your validator!")
			fmt.Println("To prevent slashing, you must wait 15 minutes from the time you stopped the clients before starting them again.\n")
			fmt.Println("**If you did NOT change clients, you can safely ignore this warning.**\n")
			if !utils.Confirm(fmt.Sprintf("Press y when you understand the above warning, have waited, and are ready to start Rocket Pool:%s", terminal.ColorReset)) {
				fmt.Println("Cancelled.")
				return nil
			}
		}
	} else {
		fmt.Printf("%sIgnoring anti-slashing safety delay.%s\n", terminal.ColorYellow, terminal.ColorReset)
	}

	// Write a note on doppelganger protection
	doppelgangerEnabled, err := cfg.IsDoppelgangerEnabled()
	if err != nil {
		fmt.Printf("%sCouldn't check if you have Doppelganger Protection enabled: %s\nIf you do, your validator will miss up to 3 attestations when it starts.\nThis is *intentional* and does not indicate a problem with your node.%s\n\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
	} else if doppelgangerEnabled {
		fmt.Printf("%sNOTE: You currently have Doppelganger Protection enabled.\nYour validator will miss up to 3 attestations when it starts.\nThis is *intentional* and does not indicate a problem with your node.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	}

	// Start service
	err = rp.StartService(getComposeFiles(c))
	if err != nil {
		return err
	}

	// Remove the upgrade flag if it's there
	return rp.RemoveUpgradeFlagFile()
}

// Check if the VC has changed and force a wait for slashing protection if it has
func checkForValidatorChange(rp *client.Client, cfg *config.RocketPoolConfig) error {
	// Get the container prefix
	prefix, err := getContainerPrefix(rp)
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
	} else if (currentValidatorName == "nimbus-eth2" && pendingValidatorName == "nimbus-validator-client") || (pendingValidatorName == "nimbus-eth2" && currentValidatorName == "nimbus-validator-client") {
		// Handle the transition from Nimbus v22.11.x to Nimbus v22.12.x where they split the VC into its own container
		fmt.Printf("Validator client [%s] was previously used, you are changing to [%s] but the Smartnode will migrate your slashing database automatically to this new client. No slashing prevention delay is necessary.\n", currentValidatorName, pendingValidatorName)
	} else {
		consensusClient, _ := cfg.GetSelectedConsensusClient()
		// Warn about Lodestar
		if consensusClient == cfgtypes.ConsensusClient_Lodestar {
			fmt.Printf("%sNOTE:\nIf this is your first time running Lodestar and you have existing minipools, you must run `rocketpool wallet rebuild` after the Smartnode starts to generate the validator keys for it.\nIf you have run it before or you don't have any minipools, you can ignore this message.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
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
			fmt.Printf("%sValidator is currently running, stopping it...%s\n", terminal.ColorYellow, terminal.ColorReset)
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
			fmt.Printf("%s=== WARNING ===\n", terminal.ColorRed)
			fmt.Printf("You have changed your validator client from %s to %s. Only %s has elapsed since you stopped %s.\n", currentValidatorName, pendingValidatorName, time.Since(validatorFinishTime), currentValidatorName)
			fmt.Printf("If you were actively validating while using %s, starting %s without waiting will cause your validators to be slashed due to duplicate attestations!", currentValidatorName, pendingValidatorName)
			fmt.Println("To prevent slashing, Rocket Pool will delay activating the new client for 15 minutes.")
			fmt.Println("See the documentation for a more detailed explanation: https://docs.rocketpool.net/guides/node/maintenance/node-migration.html#slashing-and-the-slashing-database")
			fmt.Printf("If you have read the documentation, understand the risks, and want to bypass this cooldown, run `rocketpool service start --%s`.%s\n\n", ignoreSlashTimerFlag.Name, terminal.ColorReset)

			// Wait for 15 minutes
			for remainingTime > 0 {
				fmt.Printf("Remaining time: %s", remainingTime)
				time.Sleep(1 * time.Second)
				remainingTime = time.Until(safeStartTime)
				fmt.Printf("%s\r", terminal.ClearLine)
			}

			fmt.Println(terminal.ColorReset)
			fmt.Println("You may now safely start the validator without fear of being slashed.")
		}
	}

	return nil
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

// Get the name of the container responsible for validator duties based on the client name
func getContainerNameForValidatorDuties(CurrentValidatorClientName string, rp *client.Client) (string, error) {
	prefix, err := getContainerPrefix(rp)
	if err != nil {
		return "", err
	}

	return prefix + ValidatorContainerSuffix, nil
}
