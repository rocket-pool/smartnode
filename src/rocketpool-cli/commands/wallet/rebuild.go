package wallet

import (
	"fmt"
	"os"

	"github.com/rocket-pool/node-manager-core/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

func rebuildWallet(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Load the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return err
	}

	// Get & check wallet status
	statusResponse, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}
	status := statusResponse.Data.WalletStatus
	if !wallet.IsWalletReady(status) {
		fmt.Println("The node wallet is not loaded or your node is in read-only mode. Please run `rocketpool wallet status` for more details.")
		return nil
	}

	// Check for custom keys
	customKeyPasswordFile, err := promptForCustomKeyPasswords(cfg, false)
	if err != nil {
		return err
	}
	if customKeyPasswordFile != "" {
		// Defer deleting the custom keystore password file
		defer func(customKeyPasswordFile string) {
			_, err := os.Stat(customKeyPasswordFile)
			if os.IsNotExist(err) {
				return
			}

			err = os.Remove(customKeyPasswordFile)
			if err != nil {
				fmt.Printf("*** WARNING ***\nAn error occurred while removing the custom keystore password file: %s\n\nThis file contains the passwords to your custom validator keys.\nYou *must* delete it manually as soon as possible so nobody can read it.\n\nThe file is located here:\n\n\t%s\n\n", err.Error(), customKeyPasswordFile)
			}
		}(customKeyPasswordFile)
	}

	// Log
	fmt.Println("Rebuilding node validator keystores...")

	// Rebuild wallet
	response, err := rp.Api.Wallet.Rebuild()
	if err != nil {
		return err
	}

	// Log & return
	fmt.Println("The node wallet was successfully rebuilt.")
	if len(response.Data.ValidatorKeys) > 0 {
		fmt.Println("Validator keys:")
		for _, key := range response.Data.ValidatorKeys {
			fmt.Println(key.Hex())
		}
	} else {
		fmt.Println("No validator keys were found.")
	}
	return nil
}
