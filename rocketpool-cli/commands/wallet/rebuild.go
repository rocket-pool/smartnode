package wallet

import (
	"fmt"
	"os"

	"github.com/rocket-pool/node-manager-core/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/urfave/cli/v2"
)

func rebuildWallet(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c)
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

	var enablePartialRebuildValue = false
	if enablePartialRebuild.Name != "" {
		enablePartialRebuildValue = c.Bool(enablePartialRebuild.Name)
	}

	// Log
	fmt.Println("Rebuilding node validator keystores...")
	fmt.Printf("Partial rebuild enabled: %s.\n", enablePartialRebuild.Value)

	// Rebuild wallet
	response, err := rp.Api.Wallet.Rebuild(enablePartialRebuildValue)
	if err != nil {
		return err
	}

	// Handle and print failure reasons with associated public keys
	if len(response.Data.FailureReasons) > 0 {
		fmt.Println("Some keys could not be recovered. You may need to import them manually, as they are not " +
			"associated with your node wallet mnemonic. See the documentation for more details.")
		fmt.Println("Failure reasons:")
		for pubkey, reason := range response.Data.FailureReasons {
			fmt.Printf("Public Key: %s - Failure Reason: %s\n", pubkey.Hex(), reason)
		}
	} else {
		fmt.Println("No failures reported.")
	}

	if len(response.Data.RebuiltValidatorKeys) > 0 {
		fmt.Println("Validator keys:")
		for _, key := range response.Data.RebuiltValidatorKeys {
			fmt.Println(key.Hex())
		}

		if !utils.Confirm("Would you like to restart your Validator Client now so it can attest with the recovered keys?") {
			fmt.Println("Please restart the Validator Client manually at your earliest convenience to load the keys.")
			return nil
		}

		// Restart the VC
		fmt.Println("Restarting Validator Client...")
		_, err = rp.Api.Service.RestartVc()
		if err != nil {
			fmt.Printf("Error restarting Validator Client: %s\n", err.Error())
			fmt.Println("Please restart the Validator Client manually at your earliest convenience to load the keys.")
			return nil
		}
		fmt.Println("Validator Client restarted successfully.")
	} else {
		fmt.Println("No validator keys were found.")
	}

	return nil
}
