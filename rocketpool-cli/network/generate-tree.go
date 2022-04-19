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
	colorYellow string = "\033[33m"
)

func generateRewardsTree(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Print some info
	fmt.Printf("%sNOTE: in order to generate a Merkle rewards tree for a rewards interval, you will need to have access to an execution client with archival state. By default, Geth, Infura, or Pocket will not provide this.\n\nIf your primary execution client is not an archive node, please re-run this command with the `--execution-client-url` flag set to the URL of an archive node.\n\nIf you need one, Alchemy provides a free service which you can use: https://www.alchemy.com/ethereum%s\n\n", colorYellow, colorReset)

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
	if !canResponse.IsUpgraded {
		return fmt.Errorf("The Rocket Pool contracts have not been upgraded to the new rewards system yet.")
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

	// Generate the tree
	_, err = rp.GenerateRewardsTree(index, c.String("execution-client-url"))
	if err != nil {
		return err
	}

	fmt.Printf("Your node is now generating the rewards tree for interval %d. This process could take a very long time if many users have opted into the Smoothing Pool, so it will run in the background.\n\nYou can follow its progress with `rocketpool service logs api`.\n", index)

	return nil

}
