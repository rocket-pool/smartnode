package pdao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func finalizeProposal(c *cli.Context, proposalID uint64) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check the status
	canResponse, err := rp.PDAOCanFinalizeProposal(proposalID)
	if err != nil {
		return fmt.Errorf("error checking if proposal can be finalized: %w", err)
	}
	if !canResponse.CanFinalize {
		fmt.Printf("Cannot finalize proposal %d:\n", proposalID)
		if canResponse.DoesNotExist {
			fmt.Println("There is no proposal with that ID.")
		}
		if canResponse.AlreadyFinalized {
			fmt.Println("The proposal has already been finalized.")
		}
		if canResponse.InvalidState {
			fmt.Println("The proposal has not been vetoed.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to finalize proposal %d?", proposalID))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Defeat proposal
	response, err := rp.PDAOFinalizeProposal(proposalID)
	if err != nil {
		fmt.Printf("Could not finalize proposal %d: %s.\n", proposalID, err)
	}

	fmt.Printf("Finalizing proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		fmt.Printf("Could not finalize proposal %d: %s.\n", proposalID, err)
	} else {
		fmt.Printf("Successfully finalized proposal %d.\n", proposalID)
	}

	// Return
	return nil
}
