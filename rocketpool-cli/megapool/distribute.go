package megapool

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func distribute(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if Saturn is already deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	// Get the gas estimate
	canResponse, err := rp.CanDistributeMegapool()
	if err != nil {
		return fmt.Errorf("error checking if megapool can distribute rewards: %w", err)
	}

	// Get pending rewards
	rewardsSplit, err := rp.CalculatePendingRewards()
	if err != nil {
		return fmt.Errorf("error calculating pending rewards: %w", err)
	}

	// Print rewards
	nodeRewards, _ := rewardsSplit.RewardSplit.NodeRewards.Float64()
	fmt.Printf("You're about to claim pending rewards from the megapool. The rewards will be distributed to the node's withdrawal address. The node share of rewards is %.2f", nodeRewards)

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to distribute your megapool rewards?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Distribute
	response, err := rp.DistributeMegapool()
	if err != nil {
		fmt.Printf("Could not distribute megapool rewards: %s. \n", err)
		return nil
	}

	// Log and wait for the transaction
	fmt.Printf("Distributing megapool rewards...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Return
	fmt.Printf("Successfully distributed megapool rewards.\n")
	return nil
}
