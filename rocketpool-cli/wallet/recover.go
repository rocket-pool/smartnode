package wallet

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	promptcli "github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func recoverWallet(c *cli.Context) error {

	// Get RP client
	rp, ready, err := rocketpool.NewClientFromCtx(c).WithStatus()
	if err != nil {
		return err
	}
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
	if status.WalletInitialized {
		fmt.Println("The node wallet is already initialized.")
		return nil
	}

	// Prompt a notice about test recovery
	fmt.Printf("%sNOTE:\nThis command will fully regenerate your node wallet's private key and (unless explicitly disabled) the validator keys for your minipools.\nIf you just want to test recovery to ensure it works without actually regenerating the files, please use `rocketpool wallet test-recovery` instead.%s\n\n", colorYellow, colorReset)

	// Set password if not set
	if !status.PasswordSet {
		var password string
		if c.String("password") != "" {
			password = c.String("password")
		} else {
			password = promptPassword()
		}
		if _, err := rp.SetPassword(password); err != nil {
			return err
		}
	}

	// Handle validator key recovery skipping
	skipValidatorKeyRecovery := c.Bool("skip-validator-key-recovery")
	if !skipValidatorKeyRecovery && !ready {
		fmt.Printf("%sEth Clients are not available.%s Validator keys cannot be recovered until they are synced and ready.\n", colorYellow, colorReset)
		fmt.Println("You can recover them later with 'rocketpool wallet rebuild'")
		if !promptcli.Confirm("Would you like to skip recovering the validator keys, and recover the node wallet only?") {
			fmt.Println("Cancelled.")
			return nil
		}
		skipValidatorKeyRecovery = true
	}

	// Prompt for mnemonic
	var mnemonic string
	if c.String("mnemonic") != "" {
		mnemonic = c.String("mnemonic")
	} else {
		mnemonic = PromptMnemonic()
	}
	mnemonic = strings.TrimSpace(mnemonic)

	// Check for custom keys
	if !skipValidatorKeyRecovery {
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
	}

	// Check for a search-by-address operation
	addressString := c.String("address")
	if addressString != "" {

		// Get the address to search for
		address := common.HexToAddress(addressString)
		fmt.Printf("Searching for the derivation path and index for wallet %s...\nNOTE: this may take several minutes depending on how large your wallet's index is.\n", address.Hex())

		if !skipValidatorKeyRecovery {
			fmt.Println("Recovering node wallet and validator keys...")
		} else {
			fmt.Println("Ignoring validator keys, searching for wallet only...")
		}

		// Recover wallet
		response, err := rp.SearchAndRecoverWallet(mnemonic, address, skipValidatorKeyRecovery)
		if err != nil {
			return err
		}

		// Log & return
		fmt.Println("The node wallet was successfully recovered.")
		fmt.Printf("Derivation path: %s\n", response.DerivationPath)
		fmt.Printf("Wallet index:    %d\n", response.Index)
		fmt.Printf("Node account:    %s\n", response.AccountAddress.Hex())
		if !skipValidatorKeyRecovery {
			if len(response.ValidatorKeys) > 0 {
				fmt.Println("Validator keys:")
				for _, key := range response.ValidatorKeys {
					fmt.Println(key.Hex())
				}
			} else {
				fmt.Println("No validator keys were found.")
			}
		}

	} else {

		// Get the derivation path
		derivationPath := c.String("derivation-path")
		if derivationPath != "" {
			fmt.Printf("Using a custom derivation path (%s).\n", derivationPath)
		}

		// Get the wallet index
		walletIndex := c.Uint("wallet-index")
		if walletIndex != 0 {
			fmt.Printf("Using a custom wallet index (%d).\n", walletIndex)
		}

		fmt.Println()

		if !skipValidatorKeyRecovery {
			fmt.Println("Recovering node wallet and validator keys...")
		} else {
			fmt.Println("Recovering node wallet only (ignoring validator keys)...")
		}

		// Recover wallet
		response, err := rp.RecoverWallet(mnemonic, skipValidatorKeyRecovery, derivationPath, walletIndex)
		if err != nil {
			return err
		}

		// Log & return
		fmt.Println("The node wallet was successfully recovered.")
		fmt.Printf("Node account: %s\n", response.AccountAddress.Hex())
		if !skipValidatorKeyRecovery {
			if len(response.ValidatorKeys) > 0 {
				fmt.Println("Validator keys:")
				for _, key := range response.ValidatorKeys {
					fmt.Println(key.Hex())
				}
			} else {
				fmt.Println("No validator keys were found.")
			}
		}
	}

	return nil

}
