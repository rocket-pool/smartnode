package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func purge(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	if !cliutils.Confirm(fmt.Sprintf("%sWARNING: This will delete your node wallet, all of your validator keys (including externally-generated ones in the 'custom-keys' folder), and restart your Validator Client.\nYou will NO LONGER be able to attest with this machine anymore until you recover your wallet or initialize a new one.\n\nYou MUST have your node wallet's mnemonic recorded before running this, or you will lose access to your node wallet and your validators forever!\n\n%sDo you want to continue?", colorRed, colorReset)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Purge
	_, err = rp.Purge()
	if err != nil {
		return err
	}

	fmt.Println("Deleted the node wallet and all validator keys.")
	return nil

}
