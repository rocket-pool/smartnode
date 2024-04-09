package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/node-manager-core/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	cliwallet "github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared"
	"github.com/rocket-pool/smartnode/v2/shared/config"
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

	if isNew {
		return fmt.Errorf("No configuration detected. Please run `rocketpool service config` to set up your Smart Node before running it.")
	}

	// Check if this is a new install
	oldVersion := strings.TrimPrefix(cfg.Version, "v")
	currentVersion := strings.TrimPrefix(shared.RocketPoolVersion, "v")
	isUpdate := oldVersion != currentVersion
	if isUpdate && !ignoreConfigSuggestion {
		if c.Bool("yes") || utils.Confirm("Smart Node upgrade detected - starting will overwrite certain settings with the latest defaults (such as container versions).\nYou may want to run `service config` first to see what's changed.\n\nWould you like to continue starting the service?") {
			cfg.UpdateDefaults()
			rp.SaveConfig(cfg)
			fmt.Printf("%sUpdated settings successfully.%s\n", terminal.ColorGreen, terminal.ColorReset)
		} else {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Update the templates for metrics and notifications
	if cfg.Metrics.EnableMetrics.Value {
		err := rp.UpdatePrometheusConfiguration(cfg)
		if err != nil {
			return err
		}
		err = rp.UpdateGrafanaDatabaseConfiguration(cfg)
		if err != nil {
			return err
		}
		err = rp.UpdateAlertmanagerConfiguration(cfg)
		if err != nil {
			return err
		}
	}

	// Validate the config
	errors := cfg.Validate()
	if len(errors) > 0 {
		fmt.Printf("%sYour configuration encountered errors. You must correct the following in order to start the Smart Node:\n\n", terminal.ColorRed)
		for _, err := range errors {
			fmt.Printf("%s\n\n", err)
		}
		fmt.Println(terminal.ColorReset)
		return nil
	}

	if !c.Bool(ignoreSlashTimerFlag.Name) { // Do the client swap check
		firstRun, err := checkForValidatorChange(rp, cfg)
		if err != nil {
			fmt.Printf("%sWARNING: couldn't verify that the Validator Client container can be safely restarted:\n\t%s\n", terminal.ColorYellow, err.Error())
			fmt.Println("If you are changing to a different client, it may resubmit an attestation you have already submitted.")
			fmt.Println("This will slash your validator!")
			fmt.Println("To prevent slashing, you must wait 15 minutes from the time you stopped the clients before starting them again.")
			fmt.Println()
			fmt.Println("**If you did NOT change clients, you can safely ignore this warning.**")
			fmt.Println()
			if !cliutils.Confirm(fmt.Sprintf("Press y when you understand the above warning, have waited, and are ready to start the Smart Node:%s", terminal.ColorReset)) {
				fmt.Println("Cancelled.")
				return nil
			}
		} else if firstRun {
			fmt.Println("It looks like this is your first time starting a Validator Client.")
			existingNode := cliutils.Confirm("Just to be sure, does your node have any existing, active validators attesting on the Beacon Chain?")
			if !existingNode {
				fmt.Println("Okay, great! You're safe to start. Have fun!")
			} else {
				fmt.Printf("%sSince didn't have any Validator Clients before, the Smart Node can't determine if you attested in the last 15 minutes.\n", terminal.ColorYellow)
				fmt.Println("If you did, it may resubmit an attestation you have already submitted.")
				fmt.Println("This will slash your validator!")
				fmt.Println("To prevent slashing, you must wait 15 minutes from the time you stopped the clients before starting them again.")
				fmt.Println()
				if !cliutils.Confirm(fmt.Sprintf("Press y when you understand the above warning, have waited, and are ready to start the Smart Node:%s", terminal.ColorReset)) {
					fmt.Println("Cancelled.")
					return nil
				}
			}
		}
	} else {
		fmt.Printf("%sIgnoring anti-slashing safety delay.%s\n", terminal.ColorYellow, terminal.ColorReset)
	}

	// Write a note on doppelganger protection
	if cfg.ValidatorClient.VcCommon.DoppelgangerDetection.Value {
		fmt.Printf("%sNOTE: You currently have Doppelganger Protection enabled.\nYour validator will miss up to 3 attestations when it starts.\nThis is *intentional* and does not indicate a problem with your node.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	}

	// Start service
	err = rp.StartService(getComposeFiles(c))
	if err != nil {
		return err
	}

	// Check wallet status
	fmt.Println()
	fmt.Println("Checking node wallet status...")
	var status *wallet.WalletStatus
	retries := 5
	for i := 0; i < retries; i++ {
		response, err := rp.Api.Wallet.Status()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		status = &response.Data.WalletStatus
		break
	}

	// Handle errors
	if status == nil {
		fmt.Println("The Smart Node couldn't check your node wallet status yet. Check on it again later with `rocketpool wallet status`. If you haven't madea wallet yet, you can do so now with `rocketpool wallet init`.")
		return nil
	}

	// All set
	if status.Wallet.IsLoaded {
		fmt.Printf("Your node wallet with address %s%s%s is loaded and ready to use.\n", terminal.ColorBlue, status.Wallet.WalletAddress.Hex(), terminal.ColorReset)
		return nil
	}

	// Prompt for password
	if status.Wallet.IsOnDisk {
		return promptForPassword(c, rp)
	}

	// Init
	fmt.Println("You don't have a node wallet yet.")
	if c.Bool(cliutils.YesFlag.Name) || !cliutils.Confirm("Would you like to create one now?") {
		fmt.Println("Please create one using `rocketpool wallet init` when you're ready.")
		return nil
	}
	err = cliwallet.InitWallet(c, rp)
	if err != nil {
		return fmt.Errorf("error initializing node wallet: %w", err)
	}

	// Get the wallet status
	return nil
}

// Prompt for the wallet password upon startup if it isn't available, but a wallet is on disk
func promptForPassword(c *cli.Context, rp *client.Client) error {
	fmt.Println("Your node wallet is saved, but the password is not stored on disk so it cannot be loaded automatically.")
	// Get the password
	passwordString := c.String(cliwallet.PasswordFlag.Name)
	if passwordString == "" {
		passwordString = cliwallet.PromptExistingPassword()
	}
	password, err := input.ValidateNodePassword("password", passwordString)
	if err != nil {
		return fmt.Errorf("error validating password: %w", err)
	}

	// Get the save flag
	savePassword := c.Bool(cliwallet.SavePasswordFlag.Name) || cliutils.Confirm("Would you like to save the password to disk? If you do, your node will be able to handle transactions automatically after a client restart; otherwise, you will have to repeat this command to manually enter the password after each restart.")

	// Run it
	_, err = rp.Api.Wallet.SetPassword(password, savePassword)
	if err != nil {
		fmt.Printf("%sError setting password: %s%s\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
		fmt.Println("Your service has started, but you'll need to provide the node wallet password later with `rocketpool wallet set-password`.")
		return nil
	}

	// Refresh the status
	response, err := rp.Api.Wallet.Status()
	if err != nil {
		fmt.Printf("Wallet password set.\n%sError checking node wallet: %s%s\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
		fmt.Println("Please check the service logs with `rocketpool service logs node` for more information.")
		return nil
	}
	status := response.Data.WalletStatus
	if !status.Wallet.IsLoaded {
		fmt.Println("Wallet password set, but the node wallet could not be loaded.")
		fmt.Println("Please check the service logs with `rocketpool service logs node` for more information.")
		return nil
	}
	fmt.Printf("Your node wallet with address %s%s%s is now loaded and ready to use.\n", terminal.ColorBlue, status.Wallet.WalletAddress.Hex(), terminal.ColorReset)
	return nil
}

// Check if the VC has changed and force a wait for slashing protection if it has
func checkForValidatorChange(rp *client.Client, cfg *config.SmartNodeConfig) (bool, error) {
	// Get the current validator client
	vcName := cfg.GetDockerArtifactName(config.ValidatorClientSuffix)

	// Check if it exists
	exists, err := rp.CheckIfContainerExists(vcName)
	if err != nil {
		return false, fmt.Errorf("error checking if validator client container exists: %w", err)
	}
	if !exists {
		return true, nil
	}

	// Get the client flavor
	currentTag, err := rp.GetDockerImage(vcName)
	if err != nil {
		return false, fmt.Errorf("error getting current validator client image: %w", err)
	}
	currentVcType, err := getDockerImageName(currentTag)
	if err != nil {
		return false, fmt.Errorf("error parsing current validator image [%s]: %w", currentTag, err)
	}

	// Get the new validator client according to the settings file
	pendingTag := cfg.GetVcContainerTag()
	pendingValidatorName, err := getDockerImageName(pendingTag)
	if err != nil {
		return false, fmt.Errorf("error parsing pending validator image [%s]]: %w", pendingTag, err)
	}

	// Compare the clients and warn if necessary
	if currentVcType == pendingValidatorName {
		fmt.Printf("Validator Client [%s] was previously used - no slashing prevention delay necessary.\n", currentVcType)
		return false, nil
	} else if currentVcType == "" {
		return true, nil
	} else {
		validatorFinishTime, err := rp.GetDockerContainerShutdownTime(vcName)
		if err != nil {
			return false, fmt.Errorf("error getting validator client shutdown time: %w", err)
		}

		// If it hasn't exited yet, shut it down
		zeroTime := time.Time{}
		status, err := rp.GetDockerStatus(vcName)
		if err != nil {
			return false, fmt.Errorf("error getting container [%s] status: %w", vcName, err)
		}
		if validatorFinishTime == zeroTime || status == "running" {
			fmt.Printf("%sValidator Client is currently running, stopping it...%s\n", terminal.ColorYellow, terminal.ColorReset)
			err := rp.StopContainer(vcName)
			validatorFinishTime = time.Now()
			if err != nil {
				return false, fmt.Errorf("error stopping container [%s]: %w", vcName, err)
			}
		}

		// Print the warning and start the time lockout
		safeStartTime := validatorFinishTime.Add(15 * time.Minute)
		remainingTime := time.Until(safeStartTime)
		if remainingTime <= 0 {
			fmt.Printf("The validator has been offline for %s, which is long enough to prevent slashing.\n", time.Since(validatorFinishTime))
			fmt.Println("The new client can be safely started.")
			return false, nil
		}

		fmt.Printf("%s=== WARNING ===\n", terminal.ColorRed)
		fmt.Printf("You have changed your Validator Client from %s to %s. Only %s has elapsed since you stopped %s.\n", currentVcType, pendingValidatorName, time.Since(validatorFinishTime), currentVcType)
		fmt.Printf("If you were actively validating while using %s, starting %s without waiting will cause your validators to be slashed due to duplicate attestations!", currentVcType, pendingValidatorName)
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

	return false, nil
}

// Extract the image name from a Docker image string
func getDockerImageName(image string) (string, error) {
	// Return the empty string if the validator didn't exist (probably because this is the first time starting it up)
	if image == "" {
		return "", nil
	}

	reg := regexp.MustCompile(dockerImageRegex)
	matches := reg.FindStringSubmatch(image)
	if matches == nil {
		return "", fmt.Errorf("error parsing the Docker image string [%s]", image)
	}
	imageIndex := reg.SubexpIndex("image")
	if imageIndex == -1 {
		return "", fmt.Errorf("image name not found in Docker image [%s]", image)
	}

	imageName := matches[imageIndex]
	return imageName, nil
}
