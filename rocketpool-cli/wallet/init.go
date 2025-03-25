package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	promptcli "github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/term"
)

func initWallet(c *cli.Context) error {

	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
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

	// Prompt for user confirmation before printing sensitive information
	if !(c.GlobalBool("secure-session") ||
		promptcli.ConfirmSecureSession("Creating a wallet will print sensitive information to your screen.")) {
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

	// Get the derivation path
	derivationPath := c.String("derivation-path")
	if derivationPath != "" {
		fmt.Printf("Using a custom derivation path (%s).\n\n", derivationPath)
	}

	// Initialize wallet
	response, err := rp.InitWallet(derivationPath)
	if err != nil {
		return err
	}

	// Print mnemonic
	fmt.Println("Your mnemonic phrase to recover your wallet is printed below. It can be used to recover your node account and validator keys if they are lost.")
	fmt.Println("Record this phrase somewhere secure and private. Do not share it with anyone as it will give them control of your node account and validators.")
	fmt.Println("==============================================================================================================================================")
	fmt.Println("")
	fmt.Println(response.Mnemonic)
	fmt.Println("")
	fmt.Println("==============================================================================================================================================")
	fmt.Println("")

	// Confirm mnemonic
	if !c.Bool("confirm-mnemonic") {
		confirmMnemonic(response.Mnemonic)
	}

	// Do a recover to save the wallet
	recoverResponse, err := rp.RecoverWallet(response.Mnemonic, true, derivationPath, 0)
	if err != nil {
		return fmt.Errorf("error saving wallet: %w", err)
	}

	// Sanity check the addresses
	if recoverResponse.AccountAddress != response.AccountAddress {
		return fmt.Errorf("Expected %s, but generated %s upon saving", response.AccountAddress, recoverResponse.AccountAddress)
	}

	// Clear terminal output
	_ = term.Clear()

	// Log & return
	fmt.Println("The node wallet was successfully initialized.")
	fmt.Printf("Node account: %s\n", response.AccountAddress.Hex())
	return nil

}
