package pdao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	"github.com/urfave/cli/v2"
)

var scKickAddressesFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "addresses",
	Aliases: []string{"a"},
	Usage:   "One or more addresses of the entity(s) to kick, separated by commas",
}

func proposeSecurityCouncilKick(c *cli.Context) error {
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
	selectedMembers, err := utils.GetMultiselectIndices[api.SecurityMemberDetails](c, scKickAddressesFlag.Name, options, "Please select a member to kick:")
	if err != nil {
		return fmt.Errorf("error determining selected members: %w", err)
	}

	// Handle a single kick
	var txInfo *eth.TransactionInfo
	var confirmMsg string
	if len(selectedMembers) == 1 {
		// Build the TX
		response, err := rp.Api.PDao.KickFromSecurityCouncil(selectedMembers[0].Address)
		if err != nil {
			return err
		}

		// Verify
		if !response.Data.CanPropose {
			fmt.Println("Cannot propose kick from security council:")
			if response.Data.InsufficientRpl {
				fmt.Printf("You do not have enough unlocked RPL (proposals require locking %.6f RPL, but you only have %.6f RPL staked and unlocked).", eth.WeiToEth(response.Data.ProposalBond), eth.WeiToEth(big.NewInt(0).Sub(response.Data.StakedRpl, response.Data.LockedRpl)))
			}
			if response.Data.MemberDoesNotExist {
				fmt.Println("The selected member does not exist.")
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
	} else {
		// Handle multiple kicks
		addresses := make([]common.Address, len(selectedMembers))
		for i, member := range selectedMembers {
			addresses[i] = member.Address
		}

		// Build the TX
		response, err := rp.Api.PDao.KickMultiFromSecurityCouncil(addresses)
		if err != nil {
			return err
		}

		// Verify
		if !response.Data.CanPropose {
			fmt.Println("Cannot propose kick from security council:")
			if response.Data.InsufficientRpl {
				fmt.Printf("You do not have enough unlocked RPL (proposals require locking %.6f RPL, but you only have %.6f RPL staked and unlocked).", eth.WeiToEth(response.Data.ProposalBond), eth.WeiToEth(big.NewInt(0).Sub(response.Data.StakedRpl, response.Data.LockedRpl)))
			}
			if len(response.Data.NonexistingMembers) > 0 {
				fmt.Println("The following selected members do not exist on the security council:")
				for _, member := range response.Data.NonexistingMembers {
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
	validated, err := tx.HandleTx(c, rp, txInfo,
		confirmMsg,
		"proposing kick from security council",
		"Proposing kick from security council...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil
}
