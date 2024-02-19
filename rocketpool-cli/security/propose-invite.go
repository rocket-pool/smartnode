package security

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func proposeInvite(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for Houston
	houston, err := rp.IsHoustonDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}
	if !houston.IsHoustonDeployed {
		fmt.Println("This command cannot be used until Houston has been deployed.")
		return nil
	}

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

	// Check if proposal can be made
	canPropose, err := rp.SecurityCanProposeInvite(id, address)
	if err != nil {
		return err
	}
	if !canPropose.CanPropose {
		fmt.Println("Cannot propose inviting member:")
		if canPropose.MemberAlreadyExists {
			fmt.Printf("The node %s is already a member of the security council.\n", address.Hex())
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canPropose.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to submit this proposal?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Submit proposal
	response, err := rp.SecurityProposeInvite(id, address)
	if err != nil {
		return err
	}

	fmt.Printf("Inviting %s (%s) to the security council...\n", id, address.Hex())
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully submitted an invite proposal for node %s.\n", address.Hex())
	return nil

}
