package security

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/tx"
	"github.com/urfave/cli/v2"
)

var replaceExistingAddressFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "existing-address",
	Aliases: []string{"e"},
	Usage:   "The address of the existing member",
}
var replaceNewIdFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "new-id",
	Aliases: []string{"ni"},
	Usage:   "A descriptive ID of the new entity to invite",
}
var replaceNewAddressFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "new-address",
	Aliases: []string{"na"},
	Usage:   "The address of the new entity to invite",
}

func proposeReplace(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get the list of members
	membersResponse, err := rp.Api.Security.Members()
	if err != nil {
		return fmt.Errorf("error getting list of security council members: %w", err)
	}

	// Get the address of the member to replace
	var oldID string
	var oldAddress common.Address
	oldAddressString := c.String(replaceExistingAddressFlag.Name)
	if oldAddressString == "" {
		options := make([]string, len(membersResponse.Data.Members))
		for i, member := range membersResponse.Data.Members {
			options[i] = fmt.Sprintf("%d: %s (%s), joined %s\n", i+1, member.ID, member.Address, member.JoinedTime)
		}
		selection, _ := utils.Select("Which member would you like to replace?", options)
		member := membersResponse.Data.Members[selection]
		oldID = member.ID
		oldAddress = member.Address
	} else {
		oldAddress, err = input.ValidateAddress("address", oldAddressString)
		if err != nil {
			return err
		}
		found := false
		for _, member := range membersResponse.Data.Members {
			if member.Address == oldAddress {
				oldID = member.ID
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("address %s is not a member of the security council", oldAddress.Hex())
		}
	}

	// Get the new ID
	newID := c.String(replaceNewIdFlag.Name)
	if newID == "" {
		newID = utils.Prompt("Please enter an ID for the member you'd like to invite: (no spaces)", "^\\S+$", "Invalid ID")
	}
	newID, err = input.ValidateDAOMemberID("id", newID)
	if err != nil {
		return err
	}

	// Get the new address
	newAddressString := c.String("new-address")
	if newAddressString == "" {
		newAddressString = utils.Prompt("Please enter the member's address:", "^0x[0-9a-fA-F]{40}$", "Invalid member address")
	}
	newAddress, err := input.ValidateAddress("address", newAddressString)
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.Security.ProposeReplace(oldAddress, newID, newAddress)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanPropose {
		fmt.Println("Cannot invite member:")
		if response.Data.NewMemberAlreadyExists {
			fmt.Println("The new member is already on the Security Council.")
		}
		if response.Data.OldMemberDoesNotExist {
			fmt.Println("The existing member is not on the Security Council.")
		}
		return nil
	}

	// Run the TX
	err = tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to propose removing %s (%s) from the security council and inviting %s (%s)?", oldID, oldAddress.Hex(), newID, newAddress.Hex()),
		"proposing security council replace",
		"Proposing replace in security council...",
	)
	if err != nil {
		return err
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil
}
