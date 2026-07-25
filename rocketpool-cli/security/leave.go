package security

import (
	"fmt"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func leave(yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if node can leave the security council
	canLeave, err := rp.SecurityCanLeave()
	if err != nil {
		return err
	}
	if !canLeave.CanLeave {
		fmt.Println("Cannot leave the security council:")
		if canLeave.ProposalExpired {
			fmt.Println("The proposal for you to leave the security council does not exist or has expired.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canLeave.GasLimits, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to leave the security council? This action cannot be undone!") {
		fmt.Println("Cancelled.")
		return nil
	}

	// Leave the security council
	response, err := rp.SecurityLeave()
	if err != nil {
		return err
	}

	fmt.Printf("Leaving security council...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Successfully left the security council.")
	return nil

}
