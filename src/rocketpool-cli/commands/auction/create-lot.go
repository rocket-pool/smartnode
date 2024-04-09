package auction

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func createLot(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Check lot can be created
	response, err := rp.Api.Auction.CreateLot()
	if err != nil {
		return fmt.Errorf("Error checking if creating lot is possible: %w", err)
	}
	if !response.Data.CanCreate {
		fmt.Println("Cannot create lot:")
		if response.Data.InsufficientBalance {
			fmt.Println("The auction contract does not have a sufficient RPL balance to create a lot.")
		}
		if response.Data.CreateLotDisabled {
			fmt.Println("Lot creation is currently disabled.")
		}
		return nil
	}
	if response.Data.TxInfo.SimulationResult.SimulationError != "" {
		return fmt.Errorf("error simulating create lot: %s", response.Data.TxInfo.SimulationResult.SimulationError)
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to create this lot?",
		"create lot",
		"Creating lot...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully created a new lot.")
	return nil
}
