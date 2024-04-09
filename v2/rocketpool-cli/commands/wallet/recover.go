package wallet

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

func recoverWallet(c *cli.Context) error {
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

	// Get & check wallet status
	statusResponse, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}
	status := statusResponse.Data.WalletStatus
	if status.Wallet.IsOnDisk {
		fmt.Println("The node wallet is already initialized.")
		return nil
	}

	// Prompt a notice about test recovery
	fmt.Printf("%sNOTE:\nThis command will fully regenerate your node wallet's private key and (unless explicitly disabled) the validator keys for your minipools.\nIf you just want to test recovery to ensure it works without actually regenerating the files, please use `rocketpool wallet test-recovery` instead.%s\n\n", terminal.ColorYellow, terminal.ColorReset)

	// Set password if not set
	var password string
	var savePassword bool
	if c.String(PasswordFlag.Name) != "" {
		password = c.String(PasswordFlag.Name)
	} else {
		password = PromptNewPassword()
	}

	// Ask about saving
	savePassword = utils.Confirm("Would you like to save the password to disk? If you do, your node will be able to handle transactions automatically after a client restart; otherwise, you will have to manually enter the password after each restart with `rocketpool wallet set-password`.")

	// Handle validator key recovery skipping
	skipValidatorKeyRecovery := c.Bool(skipValidatorRecoveryFlag.Name)
	if !skipValidatorKeyRecovery && !ready {
		fmt.Printf("%sEth Clients are not available.%s Validator keys cannot be recovered until they are synced and ready.\n", terminal.ColorYellow, terminal.ColorReset)
		fmt.Println("You can recover them later with 'rocketpool wallet rebuild'")
		if !utils.Confirm("Would you like to skip recovering the validator keys, and recover the node wallet only?") {
			fmt.Println("Cancelled.")
			return nil
		}
		skipValidatorKeyRecovery = true
	}

	// Prompt for mnemonic
	var mnemonic string
	if c.String(mnemonicFlag.Name) != "" {
		mnemonic = c.String(mnemonicFlag.Name)
	} else {
		mnemonic = PromptMnemonic()
	}
	mnemonic = strings.TrimSpace(mnemonic)

	// Check for custom keys
	if !skipValidatorKeyRecovery {
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
	}

	// Check for a search-by-address operation
	addressString := c.String(addressFlag.Name)
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
		response, err := rp.Api.Wallet.SearchAndRecover(mnemonic, address, &skipValidatorKeyRecovery, password, savePassword)
		if err != nil {
			return err
		}

		// Log & return
		fmt.Println("The node wallet was successfully recovered.")
		fmt.Printf("Derivation path: %s\n", response.Data.DerivationPath)
		fmt.Printf("Wallet index:    %d\n", response.Data.Index)
		fmt.Printf("Node account:    %s\n", response.Data.AccountAddress.Hex())
		if !skipValidatorKeyRecovery {
			if len(response.Data.ValidatorKeys) > 0 {
				fmt.Println("Validator keys:")
				for _, key := range response.Data.ValidatorKeys {
					fmt.Println(key.HexWithPrefix())
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
			fmt.Println("Recovering node wallet and validator keys...")
		} else {
			fmt.Println("Recovering node wallet only (ignoring validator keys)...")
		}

		// Recover wallet
		response, err := rp.Api.Wallet.Recover(derivationPath, mnemonic, &skipValidatorKeyRecovery, walletIndex, password, savePassword)
		if err != nil {
			return err
		}

		// Log & return
		fmt.Println("The node wallet was successfully recovered.")
		fmt.Printf("Node account: %s\n", response.Data.AccountAddress.Hex())
		if !skipValidatorKeyRecovery {
			if len(response.Data.ValidatorKeys) > 0 {
				fmt.Println("Validator keys:")
				for _, key := range response.Data.ValidatorKeys {
					fmt.Println(key.HexWithPrefix())
				}
			} else {
				fmt.Println("No validator keys were found.")
			}
		}
	}

	return nil
}
