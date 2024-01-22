package service

import (
	"fmt"

	"github.com/rivo/tview"
	cliconfig "github.com/rocket-pool/smartnode/rocketpool-cli/service/config"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/shared/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/urfave/cli/v2"
)

// Configure the service
func configureService(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Load the config, checking to see if it's new (hasn't been installed before)
	var oldCfg *config.RocketPoolConfig
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading user settings: %w", err)
	}

	// Check if this is a new install
	isUpdate, err := rp.IsFirstRun()
	if err != nil {
		return fmt.Errorf("error checking for first-run status: %w", err)
	}

	// For upgrades, move the config to the old one and create a new upgraded copy
	if isUpdate {
		oldCfg = cfg
		cfg = cfg.CreateCopy()
		err = cfg.UpdateDefaults()
		if err != nil {
			return fmt.Errorf("error upgrading configuration with the latest parameters: %w", err)
		}
	}

	// Save the config and exit in headless mode
	if c.NumFlags() > 0 {
		err := configureHeadless(c, cfg)
		if err != nil {
			return fmt.Errorf("error updating config from provided arguments: %w", err)
		}
		return rp.SaveConfig(cfg)
	}

	// Check for native mode
	isNative := rp.IsNative()

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

			fmt.Printf("%sWARNING: You have requested to change networks.\n\nAll of your existing chain data, your node wallet, and your validator keys will be removed. If you had a Checkpoint Sync URL provided for your Consensus client, it will be removed and you will need to specify a different one that supports the new network.\n\nPlease confirm you have backed up everything you want to keep, because it will be deleted if you answer `y` to the prompt below.\n\n%s", terminal.ColorYellow, terminal.ColorReset)

			if !utils.Confirm("Would you like the Smartnode to automatically switch networks for you? This will destroy and rebuild your `data` folder and all of Rocket Pool's Docker containers.") {
				fmt.Println("To change networks manually, please follow the steps laid out in the Node Operator's guide (https://docs.rocketpool.net/guides/node/mainnet.html).")
				return nil
			}

			err = changeNetworks(c, rp, fmt.Sprintf("%s%s", prefix, ApiContainerSuffix))
			if err != nil {
				fmt.Printf("%s%s%s\nThe Smartnode could not automatically change networks for you, so you will have to run the steps manually. Please follow the steps laid out in the Node Operator's guide (https://docs.rocketpool.net/guides/node/mainnet.html).\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
			}
			return nil
		}

		// Query for service start if this is a new installation
		if isNew {
			if !utils.Confirm("Would you like to start the Smartnode services automatically now?") {
				fmt.Println("Please run `rocketpool service start` when you are ready to launch.")
				return nil
			}
			return startService(c, true)
		}

		// Query for service start if this is old and there are containers to change
		if len(md.ContainersToRestart) > 0 {
			fmt.Println("The following containers must be restarted for the changes to take effect:")
			for _, container := range md.ContainersToRestart {
				fmt.Printf("\t%s_%s\n", prefix, container)
			}
			if !utils.Confirm("Would you like to restart them automatically now?") {
				fmt.Println("Please run `rocketpool service start` when you are ready to apply the changes.")
				return nil
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
			return startService(c, true)
		}
	} else {
		fmt.Println("Your changes have not been saved. Your Smartnode configuration is the same as it was before.")
		return nil
	}

	return err
}

// Updates a configuration from the provided CLI arguments headlessly
func configureHeadless(c *cli.Context, cfg *config.RocketPoolConfig) error {
	// Root params
	for _, param := range cfg.GetParameters() {
		err := updateConfigParamFromCliArg(c, "", param, cfg)
		if err != nil {
			return err
		}
	}

	// Subconfigs
	for sectionName, subconfig := range cfg.GetSubconfigs() {
		for _, param := range subconfig.GetParameters() {
			err := updateConfigParamFromCliArg(c, sectionName, param, cfg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Updates a config parameter from a CLI flag
func updateConfigParamFromCliArg(c *cli.Context, sectionName string, param *cfgtypes.Parameter, cfg *config.RocketPoolConfig) error {
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
