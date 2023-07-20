package wallet

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func rebuildWallet(c *cli.Context) error {

	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	// Load the config
	cfg, _, err := rp.LoadConfig()
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

	// Check for custom keys
	customKeyPasswordFile, err := promptForCustomKeyPasswords(rp, cfg, false)
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
