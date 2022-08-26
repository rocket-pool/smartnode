package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func purgeKeys(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get & check wallet status
	status, err := rp.WalletStatus()
	if err != nil {
		return err
	}
	if !status.WalletInitialized {
		fmt.Println("The node wallet is not initialized.")
		return nil
	}

	if cliutils.Confirm("WARNING:\n This command will delete all files related to validator keys. Do you want to continue?") {
		// Rebuild wallet
		_, err := rp.PurgeKeys()
		if err != nil {
			return err
		}

		fmt.Println("Deleted all validator keys.")

	}

	return nil

}
