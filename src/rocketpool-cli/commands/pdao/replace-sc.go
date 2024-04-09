package pdao

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/utils"
	"github.com/urfave/cli/v2"
)

var scReplaceExistingAddressFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "existing-address",
	Aliases: []string{"e"},
	Usage:   "The address of the existing member",
}
var scReplaceNewIdFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "new-id",
	Aliases: []string{"ni"},
	Usage:   "A descriptive ID of the new entity to invite",
}
var scReplaceNewAddressFlag *cli.StringFlag = &cli.StringFlag{
	Name:    "new-address",
	Aliases: []string{"na"},
	Usage:   "The address of the new entity to invite",
}

func proposeSecurityCouncilReplace(c *cli.Context) error {
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
	oldAddressString := c.String(scReplaceExistingAddressFlag.Name)
	if oldAddressString == "" {
		options := make([]string, len(membersResponse.Data.Members))
		for i, member := range membersResponse.Data.Members {
			options[i] = fmt.Sprintf("%d: %s (%s), joined %s\n", i+1, member.ID, member.Address, member.JoinedTime)
		}
		selection, _ := cliutils.Select("Which member would you like to replace?", options)
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
	newID := c.String(scReplaceNewIdFlag.Name)
	if newID == "" {
		newID = cliutils.Prompt("Please enter an ID for the member you'd like to invite: (no spaces)", "^\\S+$", "Invalid ID")
	}
	newID, err = utils.ValidateDaoMemberID("id", newID)
	if err != nil {
		return err
	}

	// Get the new address
	newAddressString := c.String(scReplaceNewAddressFlag.Name)
	if newAddressString == "" {
		newAddressString = cliutils.Prompt("Please enter the member's address:", "^0x[0-9a-fA-F]{40}$", "Invalid member address")
	}
	newAddress, err := input.ValidateAddress("address", newAddressString)
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.PDao.ReplaceMemberOfSecurityCouncil(oldAddress, newID, newAddress)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanPropose {
		fmt.Println("You cannot currently submit this proposal:")
		if response.Data.InsufficientRpl {
			fmt.Printf("You do not have enough unlocked RPL (proposals require locking %.6f RPL, but you only have %.6f RPL staked and unlocked).", eth.WeiToEth(response.Data.ProposalBond), eth.WeiToEth(big.NewInt(0).Sub(response.Data.StakedRpl, response.Data.LockedRpl)))
		}
		if response.Data.NewMemberAlreadyExists {
			fmt.Println("The new address is already a member of the security council.")
		}
		if response.Data.OldMemberDoesNotExist {
			fmt.Println("The existing address is not a member of the security council.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to propose removing %s (%s) from the security council and inviting %s (%s)?", oldID, oldAddress.Hex(), newID, newAddress.Hex()),
		"proposing security council replace",
		"Proposing replace in security council...",
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
