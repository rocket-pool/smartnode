package wallet

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func exportWallet(c *cli.Context) error {

	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	// Get & check wallet status
	status, err := rp.WalletStatus()
	if err != nil {
		return err
	}
	if !status.WalletInitialized {
		fmt.Println("The node wallet is not initialized.")
		return nil
	}

	if !c.GlobalBool("secure-session") {
		// Check if stdout is interactive
		stat, err := os.Stdout.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error occurred while determining whether or not the output is a tty: %v\n"+
				"Use \"rocketpool --secure-session wallet export\" to bypass.\n", err)
			os.Exit(1)
		}

		if (stat.Mode()&os.ModeCharDevice) == os.ModeCharDevice &&
			!cliutils.ConfirmSecureSession("Exporting a wallet will print sensitive information to your screen.") {
			return nil
		}
	}

	// Export wallet
	export, err := rp.ExportWallet()
	if err != nil {
		return err
	}

	// Print wallet & return
	fmt.Println("Node account private key:")
	fmt.Println("")
	fmt.Println(export.AccountPrivateKey)
	fmt.Println("")
	fmt.Println("Wallet password:")
	fmt.Println("")
	fmt.Println(export.Password)
	fmt.Println("")
	fmt.Println("Wallet file:")
	fmt.Println("============")
	fmt.Println("")
	fmt.Println(export.Wallet)
	fmt.Println("")
	fmt.Println("============")
	return nil

}
