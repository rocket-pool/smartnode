package security

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func leave(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
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
	err = gas.AssignMaxFeeAndLimit(canLeave.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to leave the security council? This action cannot be undone!")) {
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
