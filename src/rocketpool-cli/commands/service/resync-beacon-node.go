package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/urfave/cli/v2"
)

// Destroy and resync the Beacon Node from scratch
func resyncConsensusClient(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get the merged config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return err
	}
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smart Node.")
	}

	// Check the client mode
	if !cfg.IsLocalMode() {
		fmt.Println("You use an externally-managed Beacon Node. The Smart Node cannot resync it for you.")
		return nil
	}

	fmt.Println("This will delete the chain data of your Beacon Node and resync it from scratch.")
	fmt.Printf("%sYou should only do this if your Beacon Node has failed and can no longer start or sync properly.\nThis is meant to be a last resort.%s\n\n", terminal.ColorYellow, terminal.ColorReset)

	// Get the current checkpoint sync URL
	checkpointSyncUrl := cfg.LocalBeaconClient.CheckpointSyncProvider.Value
	if checkpointSyncUrl == "" {
		fmt.Printf("%sYou do not have a checkpoint sync provider configured.\nIf you have active validators, they %swill be considered offline and will lose ETH%s%s until your Beacon Node finishes syncing.\nWe strongly recommend you configure a checkpoint sync provider with `rocketpool service config` so it syncs instantly before running this.%s\n\n", terminal.ColorRed, terminal.ColorBold, terminal.ColorReset, terminal.ColorRed, terminal.ColorReset)
	} else {
		fmt.Printf("You have a checkpoint sync provider configured (%s).\nYour Beacon Node will use it to sync to the head of the Beacon Chain instantly after being rebuilt.\n\n", checkpointSyncUrl)
	}

	// Prompt for confirmation
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(fmt.Sprintf("%sAre you SURE you want to delete and resync your main Beacon Node from scratch? This cannot be undone!%s", terminal.ColorRed, terminal.ColorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stop the BN
	beaconContainerName := cfg.GetDockerArtifactName(config.BeaconNodeSuffix)
	fmt.Printf("Stopping %s...\n", beaconContainerName)
	err = rp.StopContainer(beaconContainerName)
	if err != nil {
		fmt.Printf("%sWARNING: Stopping Beacon Node container failed: %s%s\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
	}

	// Get the BN volume name
	volume, err := rp.GetClientVolumeName(beaconContainerName, clientDataVolumeName)
	if err != nil {
		return fmt.Errorf("Error getting Beacon Node volume name: %w", err)
	}

	// Remove the BN
	fmt.Printf("Deleting %s...\n", beaconContainerName)
	err = rp.RemoveContainer(beaconContainerName)
	if err != nil {
		return fmt.Errorf("Error deleting Beacon Node container: %w", err)
	}

	// Delete the the BN volume
	fmt.Printf("Deleting volume %s...\n", volume)
	err = rp.DeleteVolume(volume)
	if err != nil {
		return fmt.Errorf("Error deleting volume: %w", err)
	}

	// Restart the Smart Node
	fmt.Printf("Rebuilding %s and restarting the Smart Node stack...\n", beaconContainerName)
	err = startService(c, true)
	if err != nil {
		return fmt.Errorf("Error starting the Smart Node stack: %s", err)
	}

	fmt.Printf("\nDone! Your Beacon Node is now resyncing. You can follow its progress with `rocketpool service logs bn`.\n")
	return nil
}
