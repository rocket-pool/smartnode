package pdao

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func finalizeProposal(c *cli.Context, proposalID uint64) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.PDao.FinalizeProposal(proposalID)
	if err != nil {
		return fmt.Errorf("error checking if proposal can be finalized: %w", err)
	}

	// Verify
	if !response.Data.CanFinalize {
		fmt.Printf("Cannot finalize proposal %d:\n", proposalID)
		if response.Data.DoesNotExist {
			fmt.Println("There is no proposal with that ID.")
		}
		if response.Data.AlreadyFinalized {
			fmt.Println("The proposal has already been finalized.")
		}
		if response.Data.InvalidState {
			fmt.Println("The proposal has not been vetoed.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to finalize proposal %d?", proposalID),
		"finalizing proposal",
		"Finalizing proposal...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully finalized proposal %d.\n", proposalID)
	return nil
}
