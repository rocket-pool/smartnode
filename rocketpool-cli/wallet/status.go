package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
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
	emptyAddress := common.Address{}
	if status.IsMasquerading {
		if status.NodeAddress != emptyAddress {
			fmt.Printf("The node wallet is initialized, but you are currently masquerading as %s\n", color.LightBlue(status.AccountAddress.Hex()))
			fmt.Printf("Wallet Address: %s\n", status.NodeAddress)
			color.YellowPrintln("Due to this mismatch, the node cannot submit transactions. Use the command 'rocketpool wallet end-masquerade' to end masquerading and restore your wallet address.")
		} else {
			fmt.Printf("The node wallet has not been initialized, but you are currently masquerading as %s\n", color.LightBlue(status.AccountAddress.Hex()))
			color.YellowPrintln("The node cannot submit transactions. Use the command 'rocketpool wallet end-masquerade' to end masquerading.")
		}
	} else {
		// Not Masquerading
		if status.WalletInitialized {
			fmt.Println("The node wallet is initialized")
			fmt.Printf("Wallet Address: %s", status.AccountAddress)
		}
		if !status.WalletInitialized {
			fmt.Print("The node wallet has not been initialized.")
		}
	}

	fmt.Println()
	return nil

}
