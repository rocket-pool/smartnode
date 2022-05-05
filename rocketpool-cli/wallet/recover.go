package wallet

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func recoverWallet(c *cli.Context) error {

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
	if status.WalletInitialized {
		fmt.Println("The node wallet is already initialized.")
		return nil
	}

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

	// Prompt for mnemonic
	var mnemonic string
	if c.String("mnemonic") != "" {
		mnemonic = c.String("mnemonic")
	} else {
		mnemonic = promptMnemonic()
	}
	mnemonic = strings.TrimSpace(mnemonic)

	// Handle validator key recovery skipping
	skipValidatorKeyRecovery := c.Bool("skip-validator-key-recovery")

	// Get the derivation path
	derivationPath := c.String("derivation-path")
	if derivationPath != "" {
		fmt.Printf("Using a custom derivation path (%s).\n\n", derivationPath)
	}

	// Log
	if skipValidatorKeyRecovery {
		fmt.Println("Recovering node wallet only (ignoring validator keys)...")
	} else {
		// Check and assign the EC status
		err = cliutils.CheckExecutionClientStatus(rp)
		if err != nil {
			return err
		}
		fmt.Println("Recovering node wallet and validator keys...")
	}

	// Recover wallet
	response, err := rp.RecoverWallet(mnemonic, skipValidatorKeyRecovery, derivationPath)
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
	return nil

}
