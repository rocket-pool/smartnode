package pdao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func defeatProposal(c *cli.Context, proposalID uint64, challengedIndex uint64) error {
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

	// Check the status
	canResponse, err := rp.PDAOCanDefeatProposal(proposalID, challengedIndex)
	if err != nil {
		return fmt.Errorf("error checking if proposal can be defeated: %w", err)
	}
	if !canResponse.CanDefeat {
		fmt.Printf("Cannot defeat proposal %d with index %d:\n", proposalID, challengedIndex)
		if canResponse.DoesNotExist {
			fmt.Println("There is no proposal with that ID.")
		}
		if canResponse.AlreadyDefeated {
			fmt.Println("The proposal has already been defeated.")
		}
		if canResponse.InvalidChallengeState {
			fmt.Println("The provided tree index is not in the 'Challenged' state.")
		}
		if canResponse.StillInChallengeWindow {
			fmt.Println("The proposal is still inside of its challenge window.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to defeat proposal %d?", proposalID))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Defeat proposal
	response, err := rp.PDAODefeatProposal(proposalID, challengedIndex)
	if err != nil {
		fmt.Printf("Could not defeat proposal %d: %s.\n", proposalID, err)
	}

	fmt.Printf("Defeating proposal...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		fmt.Printf("Could not defeat proposal %d: %s.\n", proposalID, err)
	} else {
		fmt.Printf("Successfully defeated proposal %d.\n", proposalID)
	}

	// Return
	return nil
}
