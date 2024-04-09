package pdao

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func defeatProposal(c *cli.Context, proposalID uint64, challengedIndex uint64) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.PDao.DefeatProposal(proposalID, challengedIndex)
	if err != nil {
		return fmt.Errorf("error checking if proposal can be defeated: %w", err)
	}

	// Verify
	if !response.Data.CanDefeat {
		fmt.Printf("Cannot defeat proposal %d with index %d:\n", proposalID, challengedIndex)
		if response.Data.DoesNotExist {
			fmt.Println("There is no proposal with that ID.")
		}
		if response.Data.AlreadyDefeated {
			fmt.Println("The proposal has already been defeated.")
		}
		if response.Data.InvalidChallengeState {
			fmt.Println("The provided tree index is not in the 'Challenged' state.")
		}
		if response.Data.StillInChallengeWindow {
			fmt.Println("The proposal is still inside of its challenge window.")
		}
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to defeat proposal %d?", proposalID),
		"defeating proposal",
		"Defeating proposal...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully defeated proposal %d.\n", proposalID)
	return nil
}
