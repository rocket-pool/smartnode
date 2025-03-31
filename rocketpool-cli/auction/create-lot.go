package auction

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func createLot(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check lot can be created
	canCreate, err := rp.CanCreateLot()
	if err != nil {
		return fmt.Errorf("Error checking if creating lot is possible: %w", err)
	}
	if !canCreate.CanCreate {
		fmt.Println("Cannot create lot:")
		if canCreate.InsufficientBalance {
			fmt.Println("The auction contract does not have a sufficient RPL balance to create a lot.")
		}
		if canCreate.CreateLotDisabled {
			fmt.Println("Lot creation is currently disabled.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canCreate.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to create this lot?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Create lot
	response, err := rp.CreateLot()
	if err != nil {
		return err
	}

	fmt.Printf("Creating lot...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully created a new lot with ID %d.\n", response.LotId)
	return nil

}
