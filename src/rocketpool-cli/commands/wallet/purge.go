package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

func purge(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	if !utils.Confirm(fmt.Sprintf("%sWARNING: This will delete your node wallet, all of your validator keys (including externally-generated ones in the 'custom-keys' folder), and restart your Docker containers.\nYou will NO LONGER be able to attest with this machine anymore until you recover your wallet or initialize a new one.\n\nYou MUST have your node wallet's mnemonic recorded before running this, or you will lose access to your node wallet and your validators forever!\n\n%sDo you want to continue?", terminal.ColorRed, terminal.ColorReset)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Purge
	composeFiles := c.StringSlice(utils.ComposeFileFlag.Name)
	err := rp.PurgeAllKeys(composeFiles)
	if err != nil {
		return fmt.Errorf("%w\n%sTHERE WAS AN ERROR DELETING YOUR KEYS. They most likely have not been deleted. Proceed with caution.%s", err, terminal.ColorRed, terminal.ColorReset)
	}

	fmt.Printf("Deleted the node wallet and all validator keys.\n**Please verify that the keys have been removed by looking at your validator logs before continuing.**\n\n")
	fmt.Printf("%sWARNING: If you intend to use these keys for validating again on this or any other machine, you must wait **at least fifteen minutes** after running this command before you can safely begin validating with them again.\nFailure to wait **could cause you to be slashed!**%s\n\n", terminal.ColorYellow, terminal.ColorReset)

	// Warn about Reverse Hybrid
	fmt.Printf("%sNOTE: If you have an externally managed validator client attached to your node (\"reverse hybrid\" mode), those keys *have not been deleted by this process.*%s\n\n", terminal.ColorYellow, terminal.ColorReset)
	return nil
}
