package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/shared/types"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
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
	err = cliutils.PrintNetwork(cfg.GetNetwork(), isNew)
	if err != nil {
		return err
	}

	// Get wallet status
	status, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}

	// Print status & return
	emptyAddress := common.Address{}
	switch status.Data.WalletStatus {
	case types.WalletStatus_NoAddress:
		fmt.Println("The node wallet has not been initialized with an address yet.")
	case types.WalletStatus_NoKeystore:
		fmt.Println("The node wallet has not been initialized yet.")
		if status.Data.AccountAddress != emptyAddress {
			fmt.Printf("Your node is currently masquerading as node %s.\n", status.Data.AccountAddress.Hex())
			fmt.Println("It is running in 'read-only' mode and cannot transact, as does not have that node's private wallet key.")
		}
	case types.WalletStatus_NoPassword:
		fmt.Println("The node wallet has been initialized, but the Smart Node doesn't have a password loaded for your node wallet so it cannot be used.")
		if status.Data.AccountAddress != emptyAddress {
			fmt.Printf("Your node is currently running with address %s in 'read-only' mode.\n", status.Data.AccountAddress.Hex())
		}
	case types.WalletStatus_KeystoreMismatch:
		fmt.Printf("The node wallet is initialized, but you are currently masquerading as node %s.\n", status.Data.AccountAddress.Hex())
		fmt.Println("This is NOT your node's address; you do not have the private wallet key for it. Your private key is for a different address.")
		fmt.Println("Your node is running in 'read-only' mode and cannot submit transactions for that address.")
	case types.WalletStatus_Ready:
		fmt.Println("The node wallet is initialized and ready.")
		fmt.Printf("Node account: %s\n", status.Data.AccountAddress.Hex())
		fmt.Println("%sThe node's wallet keystore matches this address; it will be able to submit transactions.%s", terminal.ColorGreen, terminal.ColorReset)
	}
	return nil
}
