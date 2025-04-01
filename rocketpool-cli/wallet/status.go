package wallet

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = cliutils.PrintNetwork(cfg.GetNetwork(), isNew)
	if err != nil {
		return err
	}

	// Get wallet status
	status, err := rp.WalletStatus()
	if err != nil {
		return err
	}

	// Masquerading
	if status.WalletInitialized && status.NodeAddress != status.AccountAddress {
		fmt.Printf("The node wallet is initialized, but you are currently masquerading as %s%s%s\n", colorBlue, status.AccountAddress.Hex(), colorReset)
		fmt.Printf("Wallet Address: %s\n", status.NodeAddress.Hex())
		fmt.Printf("%sDue to this mismatch, the node cannot submit transactions. Use the command 'rocketpool wallet restore-address' to end masquerading and restore your wallet address.%s", colorYellow, colorReset)
	}
	// Not Masquerading
	if status.WalletInitialized && status.NodeAddress == status.AccountAddress {
		fmt.Println("The node wallet is initialized")
		fmt.Printf("Wallet Address: %s", status.AccountAddress.Hex())
	}
	if !status.WalletInitialized {
		fmt.Print("The node wallet has not been initialized.")
	}
	fmt.Println()
	return nil

}
