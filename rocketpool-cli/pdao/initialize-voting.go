package pdao

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func initializeVoting(c *cli.Context) error {
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

	resp, err := rp.CanInitializeVoting()
	if err != nil {
		return fmt.Errorf("error calling get-voting-initialized: %w", err)
	}

	if resp.VotingInitialized {
		fmt.Println("Node voting was already initialized")
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(resp.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to initialize voting?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Initialize voting
	response, err := rp.InitializeVoting()
	if err != nil {
		return fmt.Errorf("error calling initialize-voting: %w", err)
	}

	fmt.Printf("Initializing voting...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return fmt.Errorf("error initializing voting: %w", err)
	}

	// Log & return
	fmt.Println("Successfully initialized voting.")
	return nil
}
