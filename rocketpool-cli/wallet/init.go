package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/term"
)

func initWallet(c *cli.Context) error {

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

	// Prompt for user confirmation before printing sensitive information
	if !(c.GlobalBool("secure-session") ||
		cliutils.ConfirmSecureSession("Creating a wallet will print sensitive information to your screen.")) {
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

	// Initialize wallet
	response, err := rp.InitWallet()
	if err != nil {
		return err
	}

	// Print mnemonic
	printMnemonic(response.Mnemonic)

	// Clear terminal output
	_ = term.Clear()

	// Confirm mnemonic
	if !confirmMnemonic(response.Mnemonic) {
		// The user was unable to confirm the mnemonic, so remove the wallet and force them to restart the process.
		if err := rp.RemoveWallet(); err != nil {
			return err
		}

		fmt.Println("Wallet not initialized. Please try again.")
		return nil
	}

	// Clear terminal output
	_ = term.Clear()

	// Log & return
	fmt.Println("The node wallet was successfully initialized.")
	fmt.Printf("Node account: %s\n", response.AccountAddress.Hex())
	return nil

}
