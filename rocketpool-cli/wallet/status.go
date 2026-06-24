package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
)

func getStatus() error {

	// Get RP client
	rp := rocketpool.NewClient()
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
			fmt.Printf("Wallet Address: %s\n", color.LightBlue(status.NodeAddress.Hex()))
			color.YellowPrintln("Due to this mismatch, the node cannot submit transactions. Use the command 'rocketpool wallet end-masquerade' to end masquerading and restore your wallet address.")
		} else {
			fmt.Printf("The node wallet has not been initialized, but you are currently masquerading as %s\n", color.LightBlue(status.AccountAddress.Hex()))
			color.YellowPrintln("The node cannot submit transactions. Use the command 'rocketpool wallet end-masquerade' to end masquerading.")
		}
		if status.IsObserve {
			fmt.Println()
			fmt.Printf("The node is in %s, observing address %s.\n", color.Yellow("observe mode"), color.LightBlue(status.AccountAddress.Hex()))
			if status.NodeAddress != emptyAddress {
				fmt.Printf("Wallet Address (fee recipient): %s\n", color.LightBlue(status.NodeAddress.Hex()))
			}
			fmt.Println(" - The node and watchtower loops are using the masquerade address.")
			fmt.Println(" - Transactions will not be submitted.")
			fmt.Println(" - Your fee recipient remains set to your real node wallet address")
			color.YellowPrintln("Run 'rocketpool wallet end-masquerade' and restart the node/watchtower daemons when you have finished observing.")
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

	return nil

}
