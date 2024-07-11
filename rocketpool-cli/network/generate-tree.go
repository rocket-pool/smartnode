package network

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

const (
	colorReset  string = "\033[0m"
	colorGreen  string = "\033[32m"
	colorYellow string = "\033[33m"

	signallingAddressLink string = "https://docs.rocketpool.net/guides/houston/participate#setting-your-snapshot-signalling-address"
)

func generateRewardsTree(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print archive node info
	archiveEcUrl := cfg.Smartnode.ArchiveECUrl.Value.(string)
	if archiveEcUrl == "" {
		fmt.Printf("%sNOTE: in order to generate a Merkle rewards tree for a past rewards interval, you will likely need to have access to an Execution client with archival state.\nBy default, your Smartnode's Execution client will not provide this.\n\nPlease specify the URL of an archive-capable EC in the Smartnode section of the `rocketpool service config` Terminal UI.\nIf you need one, Alchemy provides a free service which you can use: https://www.alchemy.com/ethereum%s\n\n", colorYellow, colorReset)
	} else {
		fmt.Printf("%sYou have an archive EC specified at [%s]. This will be used for tree generation.%s\n\n", colorGreen, archiveEcUrl, colorReset)
	}

	// Get the index
	var index uint64
	if c.IsSet("index") {
		index = c.Uint64("index")
	} else {
		indexString := cliutils.Prompt("Which interval would you like to generate the Merkle rewards tree for?", "^\\d+$", "Invalid interval. Please provide a number.")
		index, err = strconv.ParseUint(indexString, 0, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid interval: %w.\n", indexString, err)
		}
	}

	// Check if generation will work
	canResponse, err := rp.CanGenerateRewardsTree(index)
	if err != nil {
		return err
	}
	if canResponse.CurrentIndex <= index {
		return fmt.Errorf("The current active rewards period is interval %d. You cannot generate the tree for interval %d until the active interval is past it.", canResponse.CurrentIndex, index)
	}

	// Confirm file overwrite
	if canResponse.TreeFileExists {
		if c.Bool("yes") {
			fmt.Println("Overwriting existing rewards file.")
		} else if !cliutils.Confirm("You already have a rewards file for this interval. Would you like to overwrite it?") {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Create the generation request
	_, err = rp.GenerateRewardsTree(index)
	if err != nil {
		return err
	}

	fmt.Printf("Your request to generate the rewards tree for interval %d has been applied, and your `watchtower` container will begin the process during its next duty check (typically 5 minutes).\nYou can follow its progress with %s`rocketpool service logs watchtower`%s.\n\n", index, colorGreen, colorReset)

	if c.Bool("yes") || cliutils.Confirm("Would you like to restart the watchtower container now, so it starts generating the file immediately?") {
		container := fmt.Sprintf("%s_watchtower", cfg.Smartnode.ProjectName.Value.(string))
		response, err := rp.RestartContainer(container)
		if err != nil {
			return fmt.Errorf("Error restarting watchtower: %w", err)
		}
		if response != container {
			return fmt.Errorf("Unexpected output while restarting watchtower: %s", response)
		}

		fmt.Println("Done!")
	}

	return nil

}
