package service

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/urfave/cli/v2"
)

// Prepares the execution client for pruning
func pruneExecutionClient(c *cli.Context) error {
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

	// Sanity checks
	if cfg.ExecutionClientMode.Value.(cfgtypes.Mode) == cfgtypes.Mode_External {
		fmt.Println("You are using an externally managed Execution client.\nThe Smartnode cannot prune it for you.")
		return nil
	}
	if cfg.IsNativeMode {
		fmt.Println("You are using Native Mode.\nThe Smartnode cannot prune your Execution client for you, you'll have to do it manually.")
	}
	selectedEc := cfg.ExecutionClient.Value.(cfgtypes.ExecutionClient)
	switch selectedEc {
	case cfgtypes.ExecutionClient_Besu:
		fmt.Println("You are using Besu as your Execution client.\nBesu does not need pruning.")
		return nil
	case cfgtypes.ExecutionClient_Geth:
		if cfg.Geth.EnablePbss.Value == true {
			fmt.Println("You have PBSS enabled for Geth. Pruning is no longer required when using PBSS.")
			return nil
		}
	}

	fmt.Println("This will shut down your main execution client and prune its database, freeing up disk space.")
	fmt.Println("Once pruning is complete, your execution client will restart automatically.\n")

	if selectedEc == cfgtypes.ExecutionClient_Geth {
		if cfg.UseFallbackClients.Value == false {
			fmt.Printf("%sYou do not have a fallback execution client configured.\nYour node will no longer be able to perform any validation duties (attesting or proposing blocks) until Geth is done pruning and has synced again.\nPlease configure a fallback client with `rocketpool service config` before running this.%s\n", terminal.ColorRed, terminal.ColorReset)
		} else {
			fmt.Println("You have fallback clients enabled. Rocket Pool (and your consensus client) will use that while the main client is pruning.")
		}
	}

	// Get the container prefix
	prefix, err := getContainerPrefix(rp)
	if err != nil {
		return fmt.Errorf("Error getting container prefix: %w", err)
	}

	// Prompt for confirmation
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm("Are you sure you want to prune your main execution client?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get the prune provisioner image
	pruneProvisioner := cfg.Smartnode.GetPruneProvisionerContainerTag()

	// Check for enough free space
	executionContainerName := prefix + ExecutionContainerSuffix
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
	if diskUsage.Free < PruneFreeSpaceRequired {
		return fmt.Errorf("%sYour disk must have 50 GiB free to prune, but it only has %s free. Please free some space before pruning.%s", terminal.ColorRed, freeSpaceHuman, terminal.ColorReset)
	}

	fmt.Printf("Your disk has %s free, which is enough to prune.\n", freeSpaceHuman)

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
	err = rp.RunPruneProvisioner(prefix+PruneProvisionerContainerSuffix, volume, pruneProvisioner)
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

	if selectedEc == cfgtypes.ExecutionClient_Nethermind {
		err = rp.RunNethermindPruneStarter(executionContainerName)
		if err != nil {
			return fmt.Errorf("Error starting Nethermind prune starter: %w", err)
		}
	}

	fmt.Printf("\nDone! Your main execution client is now pruning. You can follow its progress with `rocketpool service logs eth1`.\n")
	fmt.Println("Once it's done, it will restart automatically and resume normal operation.")

	fmt.Printf("%sNOTE: While pruning, you **cannot** interrupt the client (e.g. by restarting) or you risk corrupting the database!\nYou must let it run to completion!%s\n", terminal.ColorYellow, terminal.ColorReset)

	return nil
}
