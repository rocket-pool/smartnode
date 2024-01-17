package security

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

var inviteIdFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "id",
	Aliases: []string{"i"},
	Usage:   "A descriptive ID of the entity being invited",
}

var inviteAddressFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "address",
	Aliases: []string{"a"},
	Usage:   "The address of the entity being invited",
}

func proposeInvite(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get the ID
	id := c.String(inviteIdFlag.Name)
	if id == "" {
		id = utils.Prompt("Please enter an ID for the member you'd like to invite: (no spaces)", "^\\S+$", "Invalid ID")
	}
	id, err = input.ValidateDAOMemberID("id", id)
	if err != nil {
		return err
	}

	// Get the address
	addressString := c.String(inviteAddressFlag.Name)
	if addressString == "" {
		addressString = utils.Prompt("Please enter the member's address:", "^0x[0-9a-fA-F]{40}$", "Invalid member address")
	}
	address, err := input.ValidateAddress("address", addressString)
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.Security.ProposeInvite(id, address)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanPropose {
		fmt.Println("Cannot propose inviting member:")
		if response.Data.MemberAlreadyExists {
			fmt.Printf("The node %s is already a member of the security council.\n", address.Hex())
		}
		return nil
	}

	// Run the TX
	err = tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to invite %s (%s) to the security council?", id, address),
		"inviting member to security council",
		fmt.Sprintf("Inviting %s (%s) to the security council...\n", id, address.Hex()),
	)
	if err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted an invite proposal for %s (%s).\n", id, address.Hex())
	return nil
}
