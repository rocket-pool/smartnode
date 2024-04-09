package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/rivo/tview"
	nmc_config "github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	cliconfig "github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/service/config"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/urfave/cli/v2"
)

// Configure the service
func configureService(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Make sure the config directory exists first
	err := os.MkdirAll(rp.Context.ConfigPath, 0700)
	if err != nil {
		fmt.Printf("%sYour Smart Node user configuration directory of [%s] could not be created:%s.%s\n", terminal.ColorYellow, rp.Context.ConfigPath, err.Error(), terminal.ColorReset)
		return nil
	}

	// Load the config, checking to see if it's new (hasn't been installed before)
	var oldCfg *config.SmartNodeConfig
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading user settings: %w", err)
	}

	// Check if this is an update
	oldVersion := strings.TrimPrefix(cfg.Version, "v")
	currentVersion := strings.TrimPrefix(shared.RocketPoolVersion, "v")
	isUpdate := c.Bool(installUpdateDefaultsFlag.Name) || (oldVersion != currentVersion)

	// For upgrades, move the config to the old one and create a new upgraded copy
	if isUpdate {
		oldCfg = cfg
		cfg = cfg.CreateCopy()
		cfg.UpdateDefaults()
	}

	// Save the config and exit in headless mode
	if c.NumFlags() > 0 {
		err := updateConfigParamsFromCliArgs(c, "", cfg)
		if err != nil {
			return fmt.Errorf("error updating config from provided arguments: %w", err)
		}
		return rp.SaveConfig(cfg)
	}

	// Run the TUI
	app := tview.NewApplication()
	md := cliconfig.NewMainDisplay(app, oldCfg, cfg, isNew, isUpdate, cfg.IsNativeMode)
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
		if cfg.IsNativeMode {
			fmt.Println("Please restart your daemon service for them to take effect.")
			return nil
		}

		// Handle network changes
		if md.ChangeNetworks {
			// Remove the checkpoint sync provider
			md.Config.LocalBeaconClient.CheckpointSyncProvider.Value = ""
			err = rp.SaveConfig(md.Config)
			if err != nil {
				return fmt.Errorf("error saving config: %w", err)
			}

			fmt.Printf("%sWARNING: You have requested to change networks.\n\nAll of your existing chain data, your node wallet, and your validator keys will be removed. If you had a Checkpoint Sync URL provided for your Consensus client, it will be removed and you will need to specify a different one that supports the new network.\n\nPlease confirm you have backed up everything you want to keep, because it will be deleted if you answer `y` to the prompt below.\n\n%s", terminal.ColorYellow, terminal.ColorReset)

			if !utils.Confirm("Would you like the Smart Node to automatically switch networks for you? This will destroy and rebuild your `data` folder and all of it's Docker containers.") {
				fmt.Println("To change networks manually, please follow the steps laid out in the Node Operator's guide (https://docs.rocketpool.net/guides/node/mainnet.html).")
				return nil
			}

			nodeSuffix := config.GetContainerName(nmc_config.ContainerID_Daemon)
			nodeName := cfg.GetDockerArtifactName(nodeSuffix)
			err = changeNetworks(c, rp, nodeName)
			if err != nil {
				fmt.Printf("%s%s%s\nThe Smart Node could not automatically change networks for you, so you will have to run the steps manually. Please follow the steps laid out in the Node Operator's guide (https://docs.rocketpool.net/guides/node/mainnet.html).\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
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
				suffix := config.GetContainerName(container)
				name := cfg.GetDockerArtifactName(suffix)
				fmt.Printf("\t%s\n", name)
			}
			if !utils.Confirm("Would you like to restart them automatically now?") {
				fmt.Println("Please run `rocketpool service start` when you are ready to apply the changes.")
				return nil
			}

			fmt.Println()
			for _, container := range md.ContainersToRestart {
				suffix := config.GetContainerName(container)
				name := cfg.GetDockerArtifactName(suffix)
				fmt.Printf("Stopping %s... ", name)
				rp.StopContainer(name)
				fmt.Print("done!\n")
			}

			fmt.Println()
			fmt.Println("Applying changes and restarting containers...")
			return startService(c, true)
		}
	} else {
		fmt.Println("Your changes have not been saved. Your Smart Node configuration is the same as it was before.")
		return nil
	}

	return err
}

// Updates a config section's parameters from the CLI flags
func updateConfigParamsFromCliArgs(c *cli.Context, prefix string, section nmc_config.IConfigSection) error {
	// Handle this section's parameters
	params := section.GetParameters()
	for _, param := range params {
		var paramName string
		if prefix == "" {
			paramName = param.GetCommon().ID
		} else {
			paramName = fmt.Sprintf("%s-%s", prefix, param.GetCommon().ID)
		}

		// Ignore if it's not set
		if !c.IsSet(paramName) {
			continue
		}

		if len(param.GetOptions()) > 0 {
			selection := c.String(paramName)
			found := false
			for _, option := range param.GetOptions() {
				if option.String() == selection {
					param.SetValue(option.GetValueAsAny())
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("error setting value for %s: [%s] is not one of the valid options", paramName, selection)
			}
		} else if boolParam, ok := param.(*nmc_config.Parameter[bool]); ok {
			boolParam.Value = c.Bool(paramName)
		} else if intParam, ok := param.(*nmc_config.Parameter[int]); ok {
			intParam.Value = c.Int(paramName)
		} else if floatParam, ok := param.(*nmc_config.Parameter[float64]); ok {
			floatParam.Value = c.Float64(paramName)
		} else if stringParam, ok := param.(*nmc_config.Parameter[string]); ok {
			setting := c.String(paramName)
			if param.GetCommon().MaxLength > 0 && len(setting) > param.GetCommon().MaxLength {
				return fmt.Errorf("error setting value for %s: [%s] is too long (max length %d)", paramName, setting, param.GetCommon().MaxLength)
			}
			stringParam.Value = c.String(paramName)
		} else if uintParam, ok := param.(*nmc_config.Parameter[uint64]); ok {
			uintParam.Value = c.Uint64(paramName)
		} else if uint16Param, ok := param.(*nmc_config.Parameter[uint16]); ok {
			uint16Param.Value = uint16(c.Uint(paramName))
		} else {
			panic(fmt.Sprintf("param [%s] is not a supported type for form item binding", paramName))
		}
	}

	// Handle subconfigs
	for subconfigName, subconfig := range section.GetSubconfigs() {
		var header string
		if prefix == "" {
			header = subconfigName
		} else {
			header = prefix + "-" + subconfigName
		}
		err := updateConfigParamsFromCliArgs(c, header, subconfig)
		if err != nil {
			return fmt.Errorf("error updating params for section [%s]: %w", header, err)
		}
	}

	return nil
}
