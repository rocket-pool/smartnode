package wallet

import (
	"fmt"
	"os"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/urfave/cli/v2"
)

func exportWallet(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get & check wallet status
	status, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}
	if !status.Data.WalletStatus.Wallet.IsLoaded {
		fmt.Println("The node wallet is not loaded and ready for usage. Please run `rocketpool wallet status` for more details.")
		return nil
	}

	if !rp.Context.SecureSession {
		// Check if stdout is interactive
		stat, err := os.Stdout.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error occured while determining whether or not the output is a tty: %s\n"+
				"Use \"rocketpool --secure-session wallet export\" to bypass.\n", err.Error())
			os.Exit(1)
		}

		if (stat.Mode()&os.ModeCharDevice) == os.ModeCharDevice &&
			!utils.ConfirmSecureSession("Exporting a wallet will print sensitive information to your screen.") {
			return nil
		}
	}

	// Export wallet
	export, err := rp.Api.Wallet.Export()
	if err != nil {
		return err
	}

	// Print wallet & return
	fmt.Println("Node account private key:")
	fmt.Println("")
	fmt.Println(export.Data.AccountPrivateKey)
	fmt.Println("")
	fmt.Println("Wallet password:")
	fmt.Println("")
	fmt.Println(export.Data.Password)
	fmt.Println("")
	fmt.Println("Wallet file:")
	fmt.Println("============")
	fmt.Println("")
	fmt.Println(export.Data.Wallet)
	fmt.Println("")
	fmt.Println("============")
	return nil
}
