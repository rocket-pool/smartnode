package wallet

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func testRecovery(mnemonic, addressFlag string, skipValidatorKeyRecovery bool, derivationPath string, walletIndex uint, yes bool) error {

	// Only check client status when testing validator key recovery
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

	// A dry run still derives keys on the daemon, so it contends with a real
	// recovery for the same wallet; check before asking for the mnemonic
	running, err := checkForRunningKeyRecovery(rp, cfg)
	if err != nil {
		return err
	}
	if running {
		return nil
	}

	// Explain what this does and confirm, before asking for the mnemonic
	effects := []string{
		"Derive your node wallet's private key from the mnemonic phrase you provide",
	}
	if !skipValidatorKeyRecovery {
		effects = append(effects, "Derive the validator keys for every validator on this node")
	}
	effects = append(effects,
		"Write nothing to disk - this is a dry run to confirm that recovery would work",
		"Hold the daemon's recovery lock while it runs, so a real recovery cannot start until it finishes",
	)
	if !confirmRecoveryOperation(yes, "You are about to test recovering your node wallet.", effects) {
		return nil
	}
	color.YellowPrintln("Use `rocketpool wallet recover` to actually recover the wallet and validator keys.")
	fmt.Println()

	// Prompt for mnemonic
	if mnemonic == "" {
		mnemonic = PromptMnemonic()
	}
	mnemonic = strings.TrimSpace(mnemonic)

	// Handle validator key recovery skipping
	// Check for custom keys
	if !skipValidatorKeyRecovery {
		customKeyPasswordFile, err := promptForCustomKeyPasswords(rp, cfg, true)
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
					fmt.Println("*** WARNING ***")
					fmt.Printf("An error occurred while removing the custom keystore password file: %s\n", err.Error())
					fmt.Println()
					fmt.Println("This file contains the passwords to your custom validator keys.")
					fmt.Println("You *must* delete it manually as soon as possible so nobody can read it.")
					fmt.Println()
					fmt.Println("The file is located here:")
					fmt.Println()
					fmt.Printf("\t%s\n", customKeyPasswordFile)
					fmt.Println()
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
			if !ready {
				return fmt.Errorf("unable to recover validator keys without synced and ready clients")
			}
			fmt.Println("Testing recovery of node wallet and validator keys...")
		} else {
			fmt.Println("Ignoring validator keys, searching for wallet only...")
		}

		// Test recover wallet
		stopProgress := startRecoveryProgressReporter(rp)
		response, err := rp.TestSearchAndRecoverWallet(mnemonic, address, skipValidatorKeyRecovery)
		stopProgress()
		if err != nil {
			return err
		}

		// Log & return
		fmt.Println("The node wallet was successfully found - recovery is possible.")
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
			if !ready {
				return fmt.Errorf("unable to recover validator keys without synced and ready clients")
			}
			fmt.Println("Testing recovery of node wallet and validator keys...")
		} else {
			fmt.Println("Testing recovery of node wallet only (ignoring validator keys)...")
		}

		// Test recover wallet
		stopProgress := startRecoveryProgressReporter(rp)
		response, err := rp.TestRecoverWallet(mnemonic, skipValidatorKeyRecovery, derivationPath, walletIndex)
		stopProgress()
		if err != nil {
			return err
		}

		// Log & return
		fmt.Println("The node wallet was successfully found - recovery is possible.")
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
