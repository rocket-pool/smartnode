package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func restoreAddress(c *cli.Context) error {
	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	// Get wallet status
	status, err := rp.WalletStatus()
	if err != nil {
		return err
	}

	// Check if node wallet is loaded
	if !status.WalletInitialized {
		fmt.Println("You do not currently have a node wallet loaded, so there is no address to restore. Please see `rocketpool wallet status` for more details.")
		return nil
	}

	// Compare wallet address with node address
	if status.AccountAddress == status.NodeAddress {
		fmt.Println("Your node address is set to your wallet address; you are not currently masquerading.")
		return nil
	}

	fmt.Printf("Your node wallet is %s%s%s. If you restore it, you will no longer be masquerading as %s%s%s.\n\n", colorBlue, status.AccountAddress.Hex(), colorReset, colorBlue, status.NodeAddress, colorReset)

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to end your masquerade and restore your node address to your wallet address?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Call Api
	_, err = rp.RestoreAddress()
	if err != nil {
		return fmt.Errorf("error restoring address: %w", err)
	}

	fmt.Println("Your node address has been reset to your wallet address.")

	// Return
	return nil
}
