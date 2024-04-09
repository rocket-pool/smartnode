package network

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/shared/config"
	"github.com/urfave/cli/v2"
)

var (
	generateTreeEcFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "execution-client-url",
		Aliases: []string{"e"},
		Usage:   "The URL of a separate execution client you want to use for generation (ignore this flag to use your primary exeuction client). Use this if your primary client is not an archive node, and you need to provide a separate archive node URL.",
	}

	generateTreeIndexFlag *cli.Uint64Flag = &cli.Uint64Flag{
		Name:    "index",
		Aliases: []string{"i"},
		Usage:   "The index of the rewards interval you want to generate the tree for",
	}
)

func generateRewardsTree(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Print archive node info
	archiveEcUrl := cfg.ArchiveEcUrl.Value
	if archiveEcUrl == "" {
		fmt.Printf("%sNOTE: in order to generate a Merkle rewards tree for a past rewards interval, you will likely need to have access to an Execution client with archival state.\nBy default, your Smartnode's Execution client will not provide this.\n\nPlease specify the URL of an archive-capable EC in the Smartnode section of the `rocketpool service config` Terminal UI.\nIf you need one, Alchemy provides a free service which you can use: https://www.alchemy.com/ethereum%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	} else {
		fmt.Printf("%sYou have an archive EC specified at [%s]. This will be used for tree generation.%s\n\n", terminal.ColorGreen, archiveEcUrl, terminal.ColorReset)
	}

	// Get the index
	var index uint64
	if c.IsSet("index") {
		index = c.Uint64("index")
	} else {
		indexString := utils.Prompt("Which interval would you like to generate the Merkle rewards tree for?", "^\\d+$", "Invalid interval. Please provide a number.")
		index, err = strconv.ParseUint(indexString, 0, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid interval: %w", indexString, err)
		}
	}

	// Check if generation will work
	canResponse, err := rp.Api.Network.RewardsFileInfo(index)
	if err != nil {
		return err
	}
	if canResponse.Data.CurrentIndex <= index {
		fmt.Printf("The current active rewards period is interval %d. You cannot generate the tree for interval %d until the active interval is past it.\n", canResponse.Data.CurrentIndex, index)
		return nil
	}

	// Confirm file overwrite
	if canResponse.Data.TreeFileExists {
		if c.Bool("yes") {
			fmt.Println("Overwriting existing rewards file.")
		} else if !utils.Confirm("You already have a rewards file for this interval. Would you like to overwrite it?") {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Create the generation request
	_, err = rp.Api.Network.GenerateRewardsTree(index)
	if err != nil {
		return err
	}

	fmt.Printf("Your request to generate the rewards tree for interval %d has been applied, and your `node` container will begin the process during its next duty check (typically 5 minutes).\nYou can follow its progress with %s`rocketpool service logs node`%s.\n\n", index, terminal.ColorGreen, terminal.ColorReset)

	if c.Bool("yes") || utils.Confirm("Would you like to restart the node container now, so it starts generating the file immediately?") {
		container := cfg.GetDockerArtifactName(config.NodeSuffix)
		err := rp.RestartContainer(container)
		if err != nil {
			return fmt.Errorf("error restarting node: %w", err)
		}

		fmt.Println("Done!")
	}

	return nil
}
