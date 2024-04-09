package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/urfave/cli/v2"
)

// Settings
const (
	clientDataVolumeName string = "/ethclient"
	dataFolderVolumeName string = "/.rocketpool/data"

	PruneFreeSpaceRequired uint64 = 50 * 1024 * 1024 * 1024
	dockerImageRegex       string = ".*/(?P<image>.*):.*"
)

// Flags
var (
	dockerRemoveAllFlag *cli.BoolFlag = &cli.BoolFlag{
		Name:    "all",
		Aliases: []string{"a"},
		Usage:   "Removes all Docker images, including those currently used by the Smart Node stack. This will force a full re-download of all images when the Smart Node is restarted",
	}
)

// Get the compose file paths for a CLI context
func getComposeFiles(c *cli.Context) []string {
	return c.StringSlice(utils.ComposeFileFlag.Name)
}

// Handle a network change by terminating the service, deleting everything, and starting over
func changeNetworks(c *cli.Context, rp *client.Client, apiContainerName string) error {
	// Stop all of the containers
	fmt.Print("Stopping containers... ")
	err := rp.PauseService(getComposeFiles(c))
	if err != nil {
		return fmt.Errorf("error stopping service: %w", err)
	}
	fmt.Println("done")

	// Restart the API container
	fmt.Print("Starting API container... ")
	err = rp.StartContainer(apiContainerName)
	if err != nil {
		return fmt.Errorf("error starting API container: %w", err)
	}
	fmt.Println("done")

	// Get the path of the user's data folder
	fmt.Print("Retrieving data folder path... ")
	volumePath, err := rp.GetClientVolumeSource(apiContainerName, dataFolderVolumeName)
	if err != nil {
		return fmt.Errorf("error getting data folder path: %w", err)
	}
	fmt.Printf("done, data folder = %s\n", volumePath)

	// Delete the data folder
	fmt.Print("Removing data folder... ")
	_, err = rp.Api.Service.TerminateDataFolder()
	if err != nil {
		return err
	}
	fmt.Println("done")

	// Terminate the current setup
	fmt.Print("Removing old installation... ")
	err = rp.StopService(getComposeFiles(c))
	if err != nil {
		return fmt.Errorf("error terminating old installation: %w", err)
	}
	fmt.Println("done")

	// Create new validator folder
	fmt.Print("Recreating data folder... ")
	err = os.MkdirAll(filepath.Join(volumePath, "validators"), 0775)
	if err != nil {
		return fmt.Errorf("error recreating data folder: %w", err)
	}

	// Start the service
	fmt.Print("Starting Rocket Pool... ")
	err = rp.StartService(getComposeFiles(c))
	if err != nil {
		return fmt.Errorf("error starting service: %w", err)
	}
	fmt.Println("done")

	return nil
}

// Get the amount of free space available in the target dir
func getPartitionFreeSpace(rp *client.Client, targetDir string) (uint64, error) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return 0, fmt.Errorf("error getting partition list: %w", err)
	}
	longestPath := 0
	bestPartition := disk.PartitionStat{}
	for _, partition := range partitions {
		if strings.HasPrefix(targetDir, partition.Mountpoint) && len(partition.Mountpoint) > longestPath {
			bestPartition = partition
			longestPath = len(partition.Mountpoint)
		}
	}
	diskUsage, err := disk.Usage(bestPartition.Mountpoint)
	if err != nil {
		return 0, fmt.Errorf("error getting free disk space available: %w", err)
	}
	return diskUsage.Free, nil
}
