package service

import (
	"fmt"

	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/shared/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/urfave/cli/v2"
)

// Destroy and resync the Consensus client from scratch
func resyncConsensusClient(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get the merged config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return err
	}
	if isNew {
		return fmt.Errorf("Settings file not found. Please run `rocketpool service config` to set up your Smartnode.")
	}

	fmt.Println("This will delete the chain data of your Consensus client and resync it from scratch.")
	fmt.Printf("%sYou should only do this if your Consensus client has failed and can no longer start or sync properly.\nThis is meant to be a last resort.%s\n\n", terminal.ColorYellow, terminal.ColorReset)

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
		fmt.Printf("%sYour Consensus client (%s) does not support checkpoint sync.\nIf you have active validators, they %swill be considered offline and will leak ETH%s%s while the client is syncing.%s\n\n", terminal.ColorRed, clientName, terminal.ColorBold, terminal.ColorReset, terminal.ColorRed, terminal.ColorReset)
	} else {
		// Get the current checkpoint sync URL
		checkpointSyncUrl := cfg.ConsensusCommon.CheckpointSyncProvider.Value.(string)
		if checkpointSyncUrl == "" {
			fmt.Printf("%sYou do not have a checkpoint sync provider configured.\nIf you have active validators, they %swill be considered offline and will lose ETH%s%s until your Consensus client finishes syncing.\nWe strongly recommend you configure a checkpoint sync provider with `rocketpool service config` so it syncs instantly before running this.%s\n\n", terminal.ColorRed, terminal.ColorBold, terminal.ColorReset, terminal.ColorRed, terminal.ColorReset)
		} else {
			fmt.Printf("You have a checkpoint sync provider configured (%s).\nYour Consensus client will use it to sync to the head of the Beacon Chain instantly after being rebuilt.\n\n", checkpointSyncUrl)
		}
	}

	// Get the container prefix
	prefix, err := getContainerPrefix(rp)
	if err != nil {
		return fmt.Errorf("Error getting container prefix: %w", err)
	}

	// Prompt for confirmation
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(fmt.Sprintf("%sAre you SURE you want to delete and resync your main Consensus client from scratch? This cannot be undone!%s", terminal.ColorRed, terminal.ColorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stop ETH2
	beaconContainerName := prefix + BeaconContainerSuffix
	fmt.Printf("Stopping %s...\n", beaconContainerName)
	result, err := rp.StopContainer(beaconContainerName)
	if err != nil {
		fmt.Printf("%sWARNING: Stopping Consensus client container failed: %s%s\n", terminal.ColorYellow, err.Error(), terminal.ColorReset)
	}
	if result != beaconContainerName {
		fmt.Printf("%sWARNING: Unexpected output while stopping Consensus client container: %s%s\n", terminal.ColorYellow, result, terminal.ColorReset)
	}

	// Get ETH2 volume name
	volume, err := rp.GetClientVolumeName(beaconContainerName, clientDataVolumeName)
	if err != nil {
		return fmt.Errorf("Error getting Consensus client volume name: %w", err)
	}

	// Remove ETH2
	fmt.Printf("Deleting %s...\n", beaconContainerName)
	result, err = rp.RemoveContainer(beaconContainerName)
	if err != nil {
		return fmt.Errorf("Error deleting Consensus client container: %w", err)
	}
	if result != beaconContainerName {
		return fmt.Errorf("Unexpected output while deleting Consensus client container: %s", result)
	}

	// Delete the ETH2 volume
	fmt.Printf("Deleting volume %s...\n", volume)
	result, err = rp.DeleteVolume(volume)
	if err != nil {
		return fmt.Errorf("Error deleting volume: %w", err)
	}
	if result != volume {
		return fmt.Errorf("Unexpected output while deleting volume: %s", result)
	}

	// Restart Rocket Pool
	fmt.Printf("Rebuilding %s and restarting Rocket Pool...\n", beaconContainerName)
	err = startService(c, true)
	if err != nil {
		return fmt.Errorf("Error starting Rocket Pool: %s", err)
	}

	fmt.Printf("\nDone! Your Consensus client is now resyncing. You can follow its progress with `rocketpool service logs eth2`.\n")
	return nil
}
