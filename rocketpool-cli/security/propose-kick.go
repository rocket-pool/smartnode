package security

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/client"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli/v2"
)

var kickAddressesFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "addresses",
	Aliases: []string{"a"},
	Usage:   "The address(es) of the member(s) to propose kicking. Use commas to separate multiple addresses (no spaces).",
}

func proposeKick(c *cli.Context) error {
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
	members := membersResponse.Data.Members

	// Check for members
	if len(members) == 0 {
		fmt.Println("There are no members on the Security Council.")
		return nil
	}

	// Get selected members
	options := make([]utils.SelectionOption[api.SecurityMemberDetails], len(members))
	for i, member := range members {
		option := &options[i]
		option.Element = &members[i]
		option.ID = fmt.Sprint(member.Address)
		option.Display = fmt.Sprintf("%d: %s (%s), joined %s\n", i+1, member.ID, member.Address, member.JoinedTime)
	}
	selectedMembers, err := utils.GetMultiselectIndices[api.SecurityMemberDetails](c, kickAddressesFlag.Name, options, "Please select a member to kick:")
	if err != nil {
		return fmt.Errorf("error determining selected members: %w", err)
	}

	// Handle a single kick
	var txInfo *core.TransactionInfo
	var confirmMsg string
	if len(selectedMembers) == 1 {
		// Build the TX
		response, err := rp.Api.Security.ProposeKick(selectedMembers[0].Address)
		if err != nil {
			return err
		}

		// Verify
		if !response.Data.CanPropose {
			fmt.Println("Cannot propose kick from security council:")
			if response.Data.MemberDoesNotExist {
				fmt.Println("The selected member does not exist.")
			}
			return nil
		}
		txInfo = response.Data.TxInfo
		confirmMsg = fmt.Sprintf("Are you sure you want to propose kicking %s (%s) from the security council?", selectedMembers[0].ID, selectedMembers[0].Address.Hex())
	} else {
		// Handle multiple kicks
		addresses := make([]common.Address, len(selectedMembers))
		for i, member := range selectedMembers {
			addresses[i] = member.Address
		}

		// Build the TX
		response, err := rp.Api.Security.ProposeKickMulti(addresses)
		if err != nil {
			return err
		}

		// Verify
		if !response.Data.CanPropose {
			fmt.Println("Cannot propose kick from security council:")
			if len(response.Data.MembersDoNotExist) > 0 {
				fmt.Println("The following selected members do not exist on the security council:")
				for _, member := range response.Data.MembersDoNotExist {
					fmt.Printf("\t%s\n", member.Hex())
				}
			}
			return nil
		}

		// Create the kick string
		txInfo = response.Data.TxInfo
		var kickString string
		for _, member := range selectedMembers {
			kickString += fmt.Sprintf("\t- %s (%s)\n", member.ID, member.Address.Hex())
		}
		confirmMsg = fmt.Sprintf("Are you sure you want to propose kicking these members from the security council?\n%s", kickString)
	}

	// Run the TX
	err = tx.HandleTx(c, rp, txInfo,
		confirmMsg,
		"proposing kick from security council",
		"Proposing kick from security council...",
	)
	if err != nil {
		return err
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil
}
