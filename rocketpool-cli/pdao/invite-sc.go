package pdao

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func proposeSecurityCouncilInvite(id string, addressFlag string, yes bool) error {
	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the ID
	if id == "" {
		id = prompt.Prompt("Please enter an ID for the member you'd like to invite: (no spaces)", "^\\S+$", "Invalid ID")
	}
	id, err = cliutils.ValidateDAOMemberID("id", id)
	if err != nil {
		return err
	}

	// Get the address
	if addressFlag == "" {
		addressFlag = prompt.Prompt("Please enter the member's address:", "^0x[0-9a-fA-F]{40}$", "Invalid member address")
	}
	address, err := cliutils.ValidateAddress("address", addressFlag)
	if err != nil {
		return err
	}

	// Check submissions
	canResponse, err := rp.PDAOCanProposeInviteToSecurityCouncil(id, address)
	if err != nil {
		return err
	}
	if !canResponse.CanPropose {
		fmt.Println("Cannot propose inviting member:")
		if canResponse.MemberAlreadyExists {
			fmt.Printf("The node %s is already a member of the security council.\n", address.Hex())
		}
		if canResponse.IsRplLockingDisallowed {
			fmt.Println("Please enable RPL locking using the command 'rocketpool node allow-rpl-locking' to raise proposals.")
		}
		return nil
	}

	// Assign max fee
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(yes || prompt.Confirm("Are you sure you want to propose inviting %s (%s) to the security council?", id, address.Hex())) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit
	response, err := rp.PDAOProposeInviteToSecurityCouncil(id, address, canResponse.BlockNumber)
	if err != nil {
		return err
	}

	fmt.Printf("Proposing invitation to security council...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Proposal successfully created.")
	return nil

}
