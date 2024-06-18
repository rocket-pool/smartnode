package pdao

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func proposeSecurityCouncilInvite(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get the ID
	id := c.String("id")
	if id == "" {
		id = cliutils.Prompt("Please enter an ID for the member you'd like to invite: (no spaces)", "^\\S+$", "Invalid ID")
	}
	id, err = cliutils.ValidateDAOMemberID("id", id)
	if err != nil {
		return err
	}

	// Get the address
	addressString := c.String("address")
	if addressString == "" {
		addressString = cliutils.Prompt("Please enter the member's address:", "^0x[0-9a-fA-F]{40}$", "Invalid member address")
	}
	address, err := cliutils.ValidateAddress("address", addressString)
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
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to propose inviting %s (%s) to the security council?", id, address.Hex()))) {
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
