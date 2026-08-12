package wallet

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	promptcli "github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func recoverWallet(password, mnemonic, addressFlag string, skipValidatorKeyRecovery bool, derivationPath string, walletIndex uint, yes bool) error {

	// Only check client status when recovering validator keys
	rp := rocketpool.NewClient()
	ready := false
	if !skipValidatorKeyRecovery {
		var err error
		rp, ready, err = rp.WithStatus()
		if err != nil {
			return err
		}
	}
	defer rp.Close()

	// Load the config
	cfg, _, err := rp.LoadConfig()
	if err != nil {
		return err
	}

	running, err := checkForRunningKeyRecovery(rp, cfg)
	if err != nil {
		return err
	}
	if running {
		return nil
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

	// Explain what this does and confirm, before asking for anything sensitive
	effects := []string{
		"Regenerate your node wallet's private key from the mnemonic phrase you provide",
	}
	if skipValidatorKeyRecovery {
		effects = append(effects, "Leave your validator keys alone (--skip-validator-key-recovery was set)")
	} else {
		effects = append(effects, "Regenerate the validator keystores for every validator on this node, for all supported Validator Clients")
	}
	effects = append(effects, "Overwrite the existing node wallet and keystore files on disk")
	if !confirmRecoveryOperation(yes, "You are about to recover your node wallet.", effects) {
		return nil
	}
	color.YellowPrintln("If you only want to check that recovery works without writing any files, cancel and use `rocketpool wallet test-recovery` instead.")
	fmt.Println()

	// Set password if not set
	if !status.PasswordSet {
		if password == "" {
			password = promptPassword()
		}
		if _, err := rp.SetPassword(password); err != nil {
			return err
		}
	}

	// Handle validator key recovery skipping
	if !skipValidatorKeyRecovery && !ready {
		fmt.Println(color.Yellow("Eth Clients are not available.") + " Validator keys cannot be recovered until they are synced and ready.")
		fmt.Println("You can recover them later with 'rocketpool wallet rebuild'")
		if !promptcli.Confirm("Would you like to skip recovering the validator keys, and recover the node wallet only?") {
			fmt.Println("Cancelled.")
			return nil
		}
		skipValidatorKeyRecovery = true
	}

	// Prompt for mnemonic
	if mnemonic == "" {
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
	if addressFlag != "" {
		// Get the address to search for
		address := common.HexToAddress(addressFlag)
		fmt.Printf("Searching for the derivation path and index for wallet %s...\nNOTE: this may take several minutes depending on how large your wallet's index is.\n", address.Hex())

		if !skipValidatorKeyRecovery {
			fmt.Println("Recovering node wallet and validator keys...")
		} else {
			fmt.Println("Ignoring validator keys, searching for wallet only...")
		}

		// Recover wallet
		stopProgress := startRecoveryProgressReporter(rp)
		response, err := rp.SearchAndRecoverWallet(mnemonic, address, skipValidatorKeyRecovery)
		stopProgress()
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
		if derivationPath != "" {
			fmt.Printf("Using a custom derivation path (%s).\n", derivationPath)
		}

		// Get the wallet index
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
		stopProgress := startRecoveryProgressReporter(rp)
		response, err := rp.RecoverWallet(mnemonic, skipValidatorKeyRecovery, derivationPath, walletIndex)
		stopProgress()
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
