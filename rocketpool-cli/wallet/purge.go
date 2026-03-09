package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
	promptcli "github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func purge(c *cli.Context) error {

	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	color.RedPrintln("WARNING: This will delete your node wallet, all of your validator keys (including externally-generated ones in the 'custom-keys' folder), and restart your Docker containers.")
	color.RedPrintln("You will NO LONGER be able to attest with this machine anymore until you recover your wallet or initialize a new one.")
	fmt.Println()
	color.RedPrintln("You MUST have your node wallet's mnemonic recorded before running this, or you will lose access to your node wallet and your validators forever!")
	fmt.Println()
	if !promptcli.Confirm("Do you want to continue?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Purge
	composeFiles := c.Parent().StringSlice("compose-file")
	err := rp.PurgeAllKeys(composeFiles)
	if err != nil {
		return fmt.Errorf("%w\n"+color.Red("THERE WAS AN ERROR DELETING YOUR KEYS. They most likely have not been deleted. Proceed with caution."), err)
	}

	fmt.Println("Deleted the node wallet and all validator keys.")
	fmt.Println("**Please verify that the keys have been removed by looking at your validator logs before continuing.**")
	fmt.Println()
	color.YellowPrintln("WARNING: If you intend to use these keys for validating again on this or any other machine, you must wait **at least fifteen minutes** after running this command before you can safely begin validating with them again.")
	color.YellowPrintln("Failure to wait **could cause you to be slashed!**")
	fmt.Println()

	// Warn about Reverse Hybrid
	color.YellowPrintln("NOTE: If you have an externally managed validator client attached to your node (\"reverse hybrid\" mode), those keys *have not been deleted by this process.*")
	fmt.Println()
	return nil

}
