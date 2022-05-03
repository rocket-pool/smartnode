package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func rebuildWallet(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckExecutionClientStatus(rp)
	if err != nil {
		return err
	}

	// Get & check wallet status
	status, err := rp.WalletStatus()
	if err != nil {
		return err
	}
	if !status.WalletInitialized {
		fmt.Println("The node wallet is not initialized.")
		return nil
	}

	// Log
	fmt.Println("Rebuilding node validator keystores...")

	// Rebuild wallet
	response, err := rp.RebuildWallet()
	if err != nil {
		return err
	}

	// Log & return
	fmt.Println("The node wallet was successfully rebuilt.")
	if len(response.ValidatorKeys) > 0 {
		fmt.Println("Validator keys:")
		for _, key := range response.ValidatorKeys {
			fmt.Println(key.Hex())
		}
	} else {
		fmt.Println("No validator keys were found.")
	}
	return nil

}
