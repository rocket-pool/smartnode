package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func masquerade(c *cli.Context) error {
	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	fmt.Printf("Masquerading allows you to set your node address to any address you want. All commands will act as though your node wallet is for that address. Since you don't have the private key for that address, you can't submit transactions or sign messages though; commands will be %sread-only%s until you end the masquerade with `rocketpool wallet end-masquerade`.\n", colorYellow, colorReset)
	fmt.Println()

	// Get address
	addressString := c.String("address")
	if addressString == "" {
		addressString = prompt.Prompt("Please enter an address to masquerade as:", "^0x[0-9a-fA-F]{40}$", "Invalid address")
	}

	address, err := cliutils.ValidateAddress("address", addressString)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to masquerade as %s%s%s?", colorBlue, addressString, colorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Call API
	_, err = rp.Masquerade(address)
	if err != nil {
		return fmt.Errorf("error running masquerade: %w", err)
	}

	fmt.Printf("Your node is now masquerading as address %s%s%s.\n", colorBlue, addressString, colorReset)

	return nil
}
