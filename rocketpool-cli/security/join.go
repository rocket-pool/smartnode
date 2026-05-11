package security

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func join(yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if node can join the security council
	canJoin, err := rp.SecurityCanJoin()
	if err != nil {
		return err
	}
	if !canJoin.CanJoin {
		fmt.Println("Cannot join the security council:")
		if canJoin.ProposalExpired {
			fmt.Println("The proposal for you to join the security council does not exist or has expired.")
		}
		if canJoin.AlreadyMember {
			fmt.Println("The node is already a member of the security council.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canJoin.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to join the security council?") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Join the security council
	response, err := rp.SecurityJoin()
	if err != nil {
		return err
	}
	fmt.Printf("Joining the security council...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Successfully joined the security council.")
	return nil

}
