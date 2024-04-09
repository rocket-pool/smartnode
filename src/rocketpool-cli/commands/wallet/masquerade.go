package wallet

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

var (
	masqueradeAddressFlag *cli.StringFlag = &cli.StringFlag{
		Name:    "address",
		Aliases: []string{"a"},
		Usage:   "The address you want to masquerade as",
	}
)

func masquerade(c *cli.Context) error {
	// Get client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	fmt.Printf("Masquerading allows you to set your node address to any address you want. Your daemon will \"pretend\" to be that node, and all commands will act as though your node wallet is for that address. Since you don't have the private key for that address, you can't submit transactions or sign messages though; your node will be in %sread-only mode%s until you end the masquerade with `rocketpool wallet restore-address`.\n\n", terminal.ColorYellow, terminal.ColorReset)

	// Get the address
	addressString := c.String(masqueradeAddressFlag.Name)
	if addressString != "" {
	} else {
		addressString = utils.Prompt("Please enter the address you want to masquerade as:", "^0x[0-9a-fA-F]{40}$", "Invalid address")
	}
	address, err := input.ValidateAddress("address", addressString)
	if err != nil {
		return err
	}

	// Confirm
	if !(c.Bool(utils.YesFlag.Name) || utils.Confirm(fmt.Sprintf("Are you sure you want to masquerade as %s%s%s?", terminal.ColorBlue, addressString, terminal.ColorReset))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Run it
	_, err = rp.Api.Wallet.Masquerade(address)
	if err != nil {
		return fmt.Errorf("error running masquerade: %w", err)
	}

	fmt.Printf("Your node is now masquerading as address %s%s%s.\n\n", terminal.ColorBlue, addressString, terminal.ColorReset)
	return nil
}
