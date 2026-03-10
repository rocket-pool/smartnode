package network

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

const (
	signallingAddressLink string = "https://docs.rocketpool.net/pdao/participate#setting-your-snapshot-signalling-address"
)

// indexFlag is -1 if not set, else the index to generate the tree for
func generateRewardsTree(indexFlag int64, yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
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
		color.YellowPrintln("NOTE: in order to generate a Merkle rewards tree for a past rewards interval, you will likely need to have access to an Execution client with archival state.")
		color.YellowPrintln("By default, your Smart Node's Execution client will not provide this.")
		fmt.Println()
		color.YellowPrintln("Please specify the URL of an archive-capable EC in the Smart Node section of the `rocketpool service config` Terminal UI.")
		fmt.Println()
		color.YellowPrintln("If you need one, Alchemy provides a free service which you can use: https://www.alchemy.com/ethereum")
		fmt.Println()
	} else {
		color.GreenPrintln("You have an archive EC specified at [%s]. This will be used for tree generation.", archiveEcUrl)
		fmt.Println()
	}

	// Get the index
	var index uint64
	if indexFlag > -1 {
		index = uint64(indexFlag)
	} else {
		indexString := prompt.Prompt("Which interval would you like to generate the Merkle rewards tree for?", "^\\d+$", "Invalid interval. Please provide a number.")
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
		if yes {
			fmt.Println("Overwriting existing rewards file.")
		} else if !prompt.Confirm("You already have a rewards file for this interval. Would you like to overwrite it?") {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Create the generation request
	_, err = rp.GenerateRewardsTree(index)
	if err != nil {
		return err
	}

	fmt.Printf("Your request to generate the rewards tree for interval %d has been applied, and your `watchtower` container will begin the process during its next duty check (typically 5 minutes).\n", index)
	fmt.Println("You can follow its progress with", color.Green("`rocketpool service logs watchtower`."))

	if yes || prompt.Confirm("Would you like to restart the watchtower container now, so it starts generating the file immediately?") {
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
