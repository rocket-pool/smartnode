package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
)

func getStatus(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get the config
	cfg, isNew, err := rp.LoadConfig()
	if err != nil {
		return fmt.Errorf("Error loading configuration: %w", err)
	}

	// Print what network we're on
	err = utils.PrintNetwork(cfg.GetNetwork(), isNew)
	if err != nil {
		return err
	}

	// Get wallet response
	response, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}

	// Print status & return
	emptyAddress := common.Address{}
	status := response.Data.WalletStatus
	if !status.HasAddress {
		fmt.Println("The node wallet has not been initialized with an address yet.")
		return nil
	}
	if !status.HasKeystore {
		fmt.Println("The node wallet has not been initialized yet.")
		if response.Data.AccountAddress != emptyAddress {
			fmt.Printf("Your node is currently masquerading as %s%s%s.\n", terminal.ColorBlue, response.Data.AccountAddress.Hex(), terminal.ColorReset)
			fmt.Printf("%sIt is running in 'read-only' mode and cannot transact, as does not have that node's private wallet key.%s\n", terminal.ColorYellow, terminal.ColorReset)
		}
		return nil
	}
	if !status.HasPassword {
		fmt.Println("The node wallet has been initialized, but the Smart Node doesn't have a password loaded for your node wallet so it cannot be used.")
		if response.Data.AccountAddress != emptyAddress {
			fmt.Printf("Your node is currently running as %s%s%s in %s'read-only' mode%s.\n", terminal.ColorBlue, response.Data.AccountAddress.Hex(), terminal.ColorReset, terminal.ColorYellow, terminal.ColorReset)
		}
		return nil
	}
	if status.NodeAddress != status.KeystoreAddress {
		fmt.Printf("The node wallet is initialized, but you are currently masquerading as %s%s%s.\n", terminal.ColorBlue, response.Data.AccountAddress.Hex(), terminal.ColorReset)
		fmt.Printf("Your node wallet is for %s%s%s.\n", terminal.ColorBlue, status.KeystoreAddress.Hex(), terminal.ColorReset)
		fmt.Printf("%sDue to this mismatch, your node is running in 'read-only' mode and cannot submit transactions.%s\n", terminal.ColorYellow, terminal.ColorReset)
	} else {
		fmt.Println("The node wallet is initialized and ready.")
		fmt.Printf("Node account: %s%s%s\n", terminal.ColorGreen, response.Data.AccountAddress.Hex(), terminal.ColorReset)
		fmt.Printf("%sThe node's wallet keystore matches this address; it will be able to submit transactions.%s", terminal.ColorGreen, terminal.ColorReset)
	}

	fmt.Println()
	if status.IsPasswordSaved {
		fmt.Printf("The node wallet's password %sis saved to disk%s.\n", terminal.ColorGreen, terminal.ColorReset)
		fmt.Println("The node will be able to submit transactions automatically after a restart.")
	} else {
		fmt.Printf("The node wallet's password %sis not saved to disk%s.\n", terminal.ColorYellow, terminal.ColorReset)
		fmt.Println("You will have to manually re-enter it with <placeholder> after a restart to be able to submit transactions.")
	}

	return nil
}
