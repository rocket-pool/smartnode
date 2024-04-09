package service

import (
	"fmt"
	"path/filepath"

	"github.com/dustin/go-humanize"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/urfave/cli/v2"
)

// Import the EC volume from an external folder
func importEcData(c *cli.Context, sourceDir string) error {
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
	sourceDir, err = filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("Error converting to absolute path: %w", err)
	}

	// Check the source dir
	fmt.Println("Checking source directory...")
	migratorName := cfg.GetDockerArtifactName(string(config.ContainerID_EcMigrator))
	sourceBytes, err := rp.GetDirSizeViaEcMigrator(migratorName, sourceDir)
	if err != nil {
		return err
	}

	fmt.Println("This will import execution layer chain data that you previously exported into your execution client.")
	fmt.Println("If your execution client is running, it will be shut down.")
	fmt.Println("Once the import is complete, your execution client will restart automatically.\n")

	// Get the volume to import into
	executionContainerName := cfg.GetDockerArtifactName(config.ExecutionClientSuffix)
	volume, err := rp.GetClientVolumeName(executionContainerName, clientDataVolumeName)
	if err != nil {
		return fmt.Errorf("Error getting execution client volume name: %w", err)
	}

	// Make sure the target volume has enough space
	if err != nil {
		fmt.Printf("%sWARNING: Couldn't check the disk space used by the source folder: %s\nPlease verify you have enough free space to import the chain data before proceeding!%s\n\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
	} else {
		sourceBytesHuman := humanize.IBytes(sourceBytes)
		volumePath, err := rp.GetClientVolumeSource(executionContainerName, clientDataVolumeName)
		if err != nil {
			err = fmt.Errorf("error getting execution volume source path: %w", err)
			fmt.Printf("%sWARNING: Couldn't check the disk space free on the Docker volume partition: %s\nPlease verify you have enough free space to import the chain data before proceeding!%s\n\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
		} else {
			targetFree, err := getPartitionFreeSpace(rp, volumePath)
			if err != nil {
				fmt.Printf("%sWARNING: Couldn't check the disk space free on the Docker volume partition: %s\nPlease verify you have enough free space to import the chain data before proceeding!%s\n\n", terminal.ColorRed, err.Error(), terminal.ColorReset)
			} else {
				freeSpaceHuman := humanize.IBytes(targetFree)

				fmt.Printf("%sChain data size:         %s%s\n", terminal.ColorBlue, sourceBytesHuman, terminal.ColorReset)
				fmt.Printf("%sDocker drive free space: %s%s\n", terminal.ColorBlue, freeSpaceHuman, terminal.ColorReset)
				if targetFree < sourceBytes {
					return fmt.Errorf("%sYour Docker drive does not have enough space to hold the chain data. Please free up more space and try again.%s", terminal.ColorRed, terminal.ColorReset)
				}

				fmt.Printf("%sYour Docker drive has enough space to store the chain data.%s\n\n", terminal.ColorGreen, terminal.ColorReset)
			}
		}
	}

	// Prompt for confirmation
	fmt.Printf("%sNOTE: Importing will *delete* your existing chain data!%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	fmt.Printf("%sOnce started, this process *will not stop* until the import is complete - even if you exit the command with Ctrl+C.\nPlease do not exit until it finishes so you can watch its progress.%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm("Are you sure you want to delete your existing execution layer chain data and import other data from a backup?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	fmt.Printf("Stopping %s...\n", executionContainerName)
	err = rp.StopContainer(executionContainerName)
	if err != nil {
		return fmt.Errorf("Error stopping main execution container: %w", err)
	}

	// Run the migrator
	fmt.Printf("Importing data from %s to volume %s...\n", sourceDir, volume)
	err = rp.RunEcMigrator(migratorName, volume, sourceDir, "import")
	if err != nil {
		return fmt.Errorf("Error running EC migrator: %w", err)
	}

	// Restart the EC
	fmt.Printf("Restarting %s...\n", executionContainerName)
	err = rp.StartContainer(executionContainerName)
	if err != nil {
		return fmt.Errorf("Error starting main execution client: %w", err)
	}

	fmt.Println("\nDone! Your chain data has been imported.")
	return nil
}
