package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	promptcli "github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func purge(c *cli.Context) error {

	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	if !promptcli.Confirm(fmt.Sprintf("%sWARNING: This will delete your node wallet, all of your validator keys (including externally-generated ones in the 'custom-keys' folder), and restart your Docker containers.\nYou will NO LONGER be able to attest with this machine anymore until you recover your wallet or initialize a new one.\n\nYou MUST have your node wallet's mnemonic recorded before running this, or you will lose access to your node wallet and your validators forever!\n\n%sDo you want to continue?", colorRed, colorReset)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Purge
	composeFiles := c.Parent().StringSlice("compose-file")
	err := rp.PurgeAllKeys(composeFiles)
	if err != nil {
		return fmt.Errorf("%w\n%sTHERE WAS AN ERROR DELETING YOUR KEYS. They most likely have not been deleted. Proceed with caution.%s", err, colorRed, colorReset)
	}

	fmt.Printf("Deleted the node wallet and all validator keys.\n**Please verify that the keys have been removed by looking at your validator logs before continuing.**\n\n")
	fmt.Printf("%sWARNING: If you intend to use these keys for validating again on this or any other machine, you must wait **at least fifteen minutes** after running this command before you can safely begin validating with them again.\nFailure to wait **could cause you to be slashed!**%s\n\n", colorYellow, colorReset)

	// Warn about Reverse Hybrid
	fmt.Printf("%sNOTE: If you have an externally managed validator client attached to your node (\"reverse hybrid\" mode), those keys *have not been deleted by this process.*%s\n\n", colorYellow, colorReset)
	return nil

}
