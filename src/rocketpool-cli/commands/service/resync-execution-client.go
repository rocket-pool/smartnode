package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

// Destroy and resync the Execution client from scratch
func resyncExecutionClient(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get the config
	_, isNew, err := rp.LoadConfig()
	if err != nil {
		return err
	}
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smartnode.")
	}

	fmt.Println("This will delete the chain data of your primary Execution client and resync it from scratch.")
	fmt.Printf("%sYou should only do this if your Execution client has failed and can no longer start or sync properly.\nThis is meant to be a last resort.%s\n", terminal.ColorYellow, terminal.ColorReset)

	// Get the container prefix
	prefix, err := getContainerPrefix(rp)
	if err != nil {
		return fmt.Errorf("Error getting container prefix: %w", err)
	}

	// Prompt for confirmation
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(fmt.Sprintf("%sAre you SURE you want to delete and resync your main Execution client from scratch? This cannot be undone!%s", terminal.ColorRed, terminal.ColorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stop Execution
	executionContainerName := prefix + ExecutionContainerSuffix
	fmt.Printf("Stopping %s...\n", executionContainerName)
	result, err := rp.StopContainer(executionContainerName)
	if err != nil {
		fmt.Printf("%sWARNING: Stopping main Execution container failed: %s%s\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
	}
	if result != executionContainerName {
		fmt.Printf("%sWARNING: Unexpected output while stopping main Execution container: %s%s\n", terminal.ColorYellow, result, terminal.ColorReset)
	}

	// Get Execution volume name
	volume, err := rp.GetClientVolumeName(executionContainerName, clientDataVolumeName)
	if err != nil {
		return fmt.Errorf("Error getting Execution client volume name: %w", err)
	}

	// Remove ETH1
	fmt.Printf("Deleting %s...\n", executionContainerName)
	result, err = rp.RemoveContainer(executionContainerName)
	if err != nil {
		return fmt.Errorf("Error deleting main Execution client container: %w", err)
	}
	if result != executionContainerName {
		return fmt.Errorf("Unexpected output while deleting main Execution client container: %s", result)
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
	err = startService(c, true)
	if err != nil {
		return fmt.Errorf("Error starting Rocket Pool: %s", err)
	}

	fmt.Printf("\nDone! Your main Execution client is now resyncing. You can follow its progress with `rocketpool service logs ec`.\n")
	return nil
}
