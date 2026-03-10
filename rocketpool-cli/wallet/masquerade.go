package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func masquerade(addressFlag string, yes bool) error {
	// Get RP client
	rp := rocketpool.NewClient()
	defer rp.Close()

	fmt.Println("Masquerading allows you to set your node address to any address you want. All commands will act as though your node wallet is for that address. Since you don't have the private key for that address, you can't submit transactions or sign messages though; commands will be", color.Yellow("read-only"), "until you end the masquerade with `rocketpool wallet end-masquerade`.")
	fmt.Println()

	// Get address
	if addressFlag == "" {
		addressFlag = prompt.Prompt("Please enter an address to masquerade as:", "^0x[0-9a-fA-F]{40}$", "Invalid address")
	}

	address, err := cliutils.ValidateAddress("address", addressFlag)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(yes || prompt.Confirm("Are you sure you want to masquerade as %s?", color.LightBlue(address.Hex()))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Call API
	_, err = rp.Masquerade(address)
	if err != nil {
		return fmt.Errorf("error running masquerade: %w", err)
	}

	fmt.Printf("Your node is now masquerading as address %s.\n", color.LightBlue(address.Hex()))

	return nil
}
