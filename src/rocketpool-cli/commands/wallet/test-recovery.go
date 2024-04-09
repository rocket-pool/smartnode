package wallet

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

func testRecovery(c *cli.Context) error {
	// Get RP client
	rp, ready, err := client.NewClientFromCtx(c).WithStatus()
	if err != nil {
		return err
	}

	// Load the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return err
	}

	// Prompt a notice about test recovery
	fmt.Printf("%sNOTE:\nThis command will test the recovery of your node wallet's private key and (unless explicitly disabled) the validator keys for your minipools, but will not actually write any files; it's simply a \"dry run\" of recovery.\nUse `rocketpool wallet recover` to actually recover the wallet and validator keys.%s\n\n", terminal.ColorYellow, terminal.ColorReset)

	// Prompt for mnemonic
	var mnemonic string
	if c.String(mnemonicFlag.Name) != "" {
		mnemonic = c.String(mnemonicFlag.Name)
	} else {
		mnemonic = PromptMnemonic()
	}
	mnemonic = strings.TrimSpace(mnemonic)

	// Handle validator key recovery skipping
	skipValidatorKeyRecovery := c.Bool(skipValidatorRecoveryFlag.Name)
	if !skipValidatorKeyRecovery {
		// Check for custom keys
		customKeyPasswordFile, err := promptForCustomKeyPasswords(cfg, true)
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
	}

	// Check for a search-by-address operation
	addressString := c.String(addressFlag.Name)
	if addressString != "" {
		// Get the address to search for
		address := common.HexToAddress(addressString)
		fmt.Printf("Searching for the derivation path and index for wallet %s...\nNOTE: this may take several minutes depending on how large your wallet's index is.\n", address.Hex())

		if !skipValidatorKeyRecovery {
			if !ready {
				return fmt.Errorf("unable to recover validator keys without synced and ready clients")
			}
			fmt.Println("Testing recovery of node wallet and validator keys...")
		} else {
			fmt.Println("Ignoring validator keys, searching for wallet only...")
		}

		// Test recover wallet
		response, err := rp.Api.Wallet.TestSearchAndRecover(mnemonic, address, &skipValidatorKeyRecovery)
		if err != nil {
			return err
		}

		// Log & return
		fmt.Println("The node wallet was successfully found - recovery is possible.")
		fmt.Printf("Derivation path: %s\n", response.Data.DerivationPath)
		fmt.Printf("Wallet index:    %d\n", response.Data.Index)
		fmt.Printf("Node account:    %s\n", response.Data.AccountAddress.Hex())
		if !skipValidatorKeyRecovery {
			if len(response.Data.ValidatorKeys) > 0 {
				fmt.Println("Validator keys:")
				for _, key := range response.Data.ValidatorKeys {
					fmt.Println(key.Hex())
				}
			} else {
				fmt.Println("No validator keys were found.")
			}
		}
	} else {
		// Get the derivation path
		derivationPathString := c.String(derivationPathFlag.Name)
		var derivationPath *string
		if derivationPathString != "" {
			fmt.Printf("Using a custom derivation path (%s).\n", derivationPathString)
			derivationPath = &derivationPathString
		}

		// Get the wallet index
		walletIndexVal := c.Uint64(walletIndexFlag.Name)
		var walletIndex *uint64
		if walletIndexVal != 0 {
			fmt.Printf("Using a custom wallet index (%d).\n", walletIndex)
			walletIndex = &walletIndexVal
		}

		fmt.Println()
		if !skipValidatorKeyRecovery {
			if !ready {
				return fmt.Errorf("unable to recover validator keys without synced and ready clients")
			}
			fmt.Println("Testing recovery of node wallet and validator keys...")
		} else {
			fmt.Println("Testing recovery of node wallet only (ignoring validator keys)...")
		}

		// Test recover wallet
		response, err := rp.Api.Wallet.TestRecover(derivationPath, mnemonic, &skipValidatorKeyRecovery, walletIndex)
		if err != nil {
			return err
		}

		// Log & return
		fmt.Println("The node wallet was successfully found - recovery is possible.")
		fmt.Printf("Node account: %s\n", response.Data.AccountAddress.Hex())
		if !skipValidatorKeyRecovery {
			if len(response.Data.ValidatorKeys) > 0 {
				fmt.Println("Validator keys:")
				for _, key := range response.Data.ValidatorKeys {
					fmt.Println(key.Hex())
				}
			} else {
				fmt.Println("No validator keys were found.")
			}
		}
	}

	return nil
}
