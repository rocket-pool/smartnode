package wallet

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

const (
	colorReset  string = "\033[0m"
	colorRed    string = "\033[31m"
	colorGreen  string = "\033[32m"
	colorYellow string = "\033[33m"
)

func testMnemonic(c *cli.Context) error {

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
	if !status.WalletInitialized {
		fmt.Println("The node wallet has not been initialized yet.")
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

	// Test wallet recovery
	response, err := rp.TestMnemonic(mnemonic, derivationPath, walletIndex)
	if err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Your current node address:  %s\n", response.CurrentAddress.Hex())
	fmt.Printf("The recovered test address: %s\n\n", response.RecoveredAddress.Hex())
	if response.CurrentAddress == response.RecoveredAddress {
		fmt.Printf("%sYour addresses match! You have the correct mnemonic (and derivation path if you specified a custom one).%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("%sYour addresses do not match! You either have an incorrect mnemonic, you made a mistake while entering it, or you used the wrong derivation path.%s\n", colorRed, colorReset)
	}

	return nil

}
