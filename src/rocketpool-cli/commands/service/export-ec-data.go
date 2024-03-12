package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

var (
	exportEcDataForceFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:  "force",
		Usage: "Bypass the free space check on the target folder",
	}
	exportEcDataDirtyFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:  "dirty",
		Usage: "Exports the execution (eth1) chain data without stopping the client. Requires a second pass (much faster) to sync the remaining files without the client running.",
	}
)

// Export the EC volume to an external folder
func exportEcData(c *cli.Context, targetDir string) error {
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

	// Make the path absolute
	targetDir, err = filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("Error converting to absolute path: %w", err)
	}

	// Make sure the target dir exists and is accessible
	targetDirInfo, err := os.Stat(targetDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("Target directory [%s] does not exist.", targetDir)
	} else if err != nil {
		return fmt.Errorf("Error reading target dir: %w", err)
	}
	if !targetDirInfo.IsDir() {
		return fmt.Errorf("Target directory [%s] is not a directory.", targetDir)
	}

	fmt.Println("This will export your execution client's chain data to an external directory, such as a portable hard drive.")
	fmt.Println("If your execution client is running, it will be shut down.")
	fmt.Println("Once the export is complete, your execution client will restart automatically.\n")

	// Get the container prefix
	prefix, err := getContainerPrefix(rp)
	if err != nil {
		return fmt.Errorf("Error getting container prefix: %w", err)
	}

	// Get the EC volume name
	executionContainerName := prefix + ExecutionContainerSuffix
	volume, err := rp.GetClientVolumeName(executionContainerName, clientDataVolumeName)
	if err != nil {
		return fmt.Errorf("Error getting execution client volume name: %w", err)
	}

	if !c.Bool(exportEcDataForceFlag.Name) {
		// Make sure the target dir has enough space
		volumeBytes, err := getVolumeSpaceUsed(rp, volume)
		if err != nil {
			fmt.Printf("%sWARNING: Couldn't check the disk space used by the Execution client volume: %s\nPlease verify you have enough free space to store the chain data in the target folder before proceeding!%s\n\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
		} else {
			volumeBytesHuman := humanize.IBytes(volumeBytes)
			targetFree, err := getPartitionFreeSpace(rp, targetDir)
			if err != nil {
				fmt.Printf("%sWARNING: Couldn't get the free space available on the target folder: %s\nPlease verify you have enough free space to store the chain data in the target folder before proceeding!%s\n\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
			} else {
				freeSpaceHuman := humanize.IBytes(targetFree)
				fmt.Printf("%sChain data size:       %s%s\n", terminal.ColorBlue, volumeBytesHuman, terminal.ColorReset)
				fmt.Printf("%sTarget dir free space: %s%s\n", terminal.ColorBlue, freeSpaceHuman, terminal.ColorReset)
				if targetFree < volumeBytes {
					return fmt.Errorf("%sYour target directory does not have enough space to hold the chain data. Please free up more space and try again or use the --%s flag to ignore this check.%s", terminal.ColorRed, exportEcDataForceFlag.Name, terminal.ColorReset)
				}

				fmt.Printf("%sYour target directory has enough space to store the chain data.%s\n\n", terminal.ColorGreen, terminal.ColorReset)
			}
		}
	}

	// Prompt for confirmation
	fmt.Printf("%sNOTE: Once started, this process *will not stop* until the export is complete - even if you exit the command with Ctrl+C.\nPlease do not exit until it finishes so you can watch its progress.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm("Are you sure you want to export your execution layer chain data?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	var result string
	// If dirty flag is used, copies chain data without stopping the eth1 client.
	// This requires a second quick pass to sync the remaining files after stopping the client.
	if !c.Bool(exportEcDataDirtyFlag.Name) {
		fmt.Printf("Stopping %s...\n", executionContainerName)
		result, err := rp.StopContainer(executionContainerName)
		if err != nil {
			return fmt.Errorf("Error stopping main execution container: %w", err)
		}
		if result != executionContainerName {
			return fmt.Errorf("Unexpected output while stopping main execution container: %s", result)
		}
	}

	// Run the migrator
	ecMigrator := cfg.Smartnode.GetEcMigratorContainerTag()
	fmt.Printf("Exporting data from volume %s to %s...\n", volume, targetDir)
	err = rp.RunEcMigrator(prefix+EcMigratorContainerSuffix, volume, targetDir, "export", ecMigrator)
	if err != nil {
		return fmt.Errorf("Error running EC migrator: %w", err)
	}

	if !c.Bool(exportEcDataDirtyFlag.Name) {
		// Restart ETH1
		fmt.Printf("Restarting %s...\n", executionContainerName)
		result, err = rp.StartContainer(executionContainerName)
		if err != nil {
			return fmt.Errorf("Error starting main execution client: %w", err)
		}
		if result != executionContainerName {
			return fmt.Errorf("Unexpected output while starting main execution client: %s", result)
		}
	}

	fmt.Println("\nDone! Your chain data has been exported.")
	return nil
}

// Get the amount of space used by a Docker volume
func getVolumeSpaceUsed(rp *client.Client, volume string) (uint64, error) {
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
