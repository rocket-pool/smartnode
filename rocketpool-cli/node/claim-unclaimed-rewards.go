package node

import (
	"fmt"
	"math/big"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func claimUnclaimedRewards(yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get node status
	status, err := rp.NodeStatus()
	if err != nil {
		return err
	}

	// Show unclaimed rewards status
	fmt.Printf("The node's withdrawal address is %s\n", status.PrimaryWithdrawalAddress)
	if status.UnclaimedRewards != nil && status.UnclaimedRewards.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("You have %.6f ETH in unclaimed rewards.\n", math.RoundDown(math.WeiToEth(status.UnclaimedRewards), 6))
		fmt.Printf("Your node %s has rewards that were distributed to its unclaimed rewards balance instead of its withdrawal address. ",
			color.LightBlue(status.AccountAddress.String()))
		fmt.Println("This happens whenever your fee distributor is distributed by anyone other than your node or withdrawal address.")
		fmt.Println("Claiming will send them to your current withdrawal address.")
	} else {
		fmt.Println("You have no unclaimed rewards.")
		fmt.Println("Unclaimed rewards occur when your fee distributor is distributed by anyone other than your node or withdrawal address.")
		fmt.Println("If you have unclaimed rewards in the future, you can use this command to claim them.")
		return nil
	}

	// Check the node can claim unclaimed rewards
	canClaim, err := rp.CanClaimUnclaimedRewards(status.AccountAddress)
	if err != nil {
		fmt.Println("Could not claim unclaimed rewards. If your current withdrawal address cannot accept ETH, use `rocketpool node set-primary-withdrawal-address` to reconfigure your withdrawal address.")
		return err
	}
	if !canClaim.CanClaim {
		fmt.Println("You have no unclaimed rewards.")
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canClaim.GasLimits, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to claim %.6f ETH in unclaimed rewards?", math.RoundDown(math.WeiToEth(status.UnclaimedRewards), 6)) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Claim unclaimed rewards
	response, err := rp.ClaimUnclaimedRewards(status.AccountAddress)
	if err != nil {
		return err
	}

	fmt.Printf("Claiming unclaimed rewards...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully claimed %.6f ETH in unclaimed rewards.\n", math.RoundDown(math.WeiToEth(status.UnclaimedRewards), 6))
	return nil

}
