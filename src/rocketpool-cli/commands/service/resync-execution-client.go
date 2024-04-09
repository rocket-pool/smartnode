package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/urfave/cli/v2"
)

// Destroy and resync the Execution client from scratch
func resyncExecutionClient(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return err
	}
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smartnode.")
	}

	// Check the client mode
	if !cfg.IsLocalMode() {
		fmt.Println("You use an externally-managed Execution Client. The Smart Node cannot resync it for you.")
		return nil
	}

	fmt.Println("This will delete the chain data of your primary Execution client and resync it from scratch.")
	fmt.Printf("%sYou should only do this if your Execution client has failed and can no longer start or sync properly.\nThis is meant to be a last resort.%s\n", terminal.ColorYellow, terminal.ColorReset)

	// Prompt for confirmation
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(fmt.Sprintf("%sAre you SURE you want to delete and resync your main Execution client from scratch? This cannot be undone!%s", terminal.ColorRed, terminal.ColorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stop Execution
	executionContainerName := cfg.GetDockerArtifactName(config.ExecutionClientSuffix)
	fmt.Printf("Stopping %s...\n", executionContainerName)
	err = rp.StopContainer(executionContainerName)
	if err != nil {
		fmt.Printf("%sWARNING: Stopping main Execution client container failed: %s%s\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
	}

	// Get Execution volume name
	volume, err := rp.GetClientVolumeName(executionContainerName, clientDataVolumeName)
	if err != nil {
		return fmt.Errorf("Error getting Execution client volume name: %w", err)
	}

	// Remove the EC
	fmt.Printf("Deleting %s...\n", executionContainerName)
	err = rp.RemoveContainer(executionContainerName)
	if err != nil {
		return fmt.Errorf("Error deleting main Execution client container: %w", err)
	}

	// Delete the EC volume
	fmt.Printf("Deleting volume %s...\n", volume)
	err = rp.DeleteVolume(volume)
	if err != nil {
		return fmt.Errorf("Error deleting volume: %w", err)
	}

	// Restart Rocket Pool
	fmt.Printf("Rebuilding %s and restarting the Smart Node stack...\n", executionContainerName)
	err = startService(c, true)
	if err != nil {
		return fmt.Errorf("Error starting the Smart Node stack: %s", err)
	}

	fmt.Printf("\nDone! Your main Execution client is now resyncing. You can follow its progress with `rocketpool service logs ec`.\n")
	return nil
}
