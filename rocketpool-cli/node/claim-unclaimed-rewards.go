package node

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/urfave/cli"
)

func claimUnclaimedRewards(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
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
		fmt.Printf("You have %.6f ETH in unclaimed rewards.\n", math.RoundDown(eth.WeiToEth(status.UnclaimedRewards), 6))
		fmt.Printf("Your node %s%s%s's rewards were distributed, but the withdrawal address (at the time of distribution) was unable to accept ETH. ",
			colorBlue, status.AccountAddress, colorReset)
		fmt.Println("Before continuing, please use the command `rocketpool node set-primary-withdrawal-address` to configure an address that can accept ETH")
	} else {
		fmt.Println("You have no unclaimed rewards.")
		fmt.Println("Unclaimed rewards occur when a withdrawal address cannot accept ETH during distribution.")
		fmt.Println("If you have unclaimed rewards in the future, you can use this command to claim them.")
		return nil
	}

	// Check the node can claim unclaimed rewards
	canClaim, err := rp.CanClaimUnclaimedRewards(status.AccountAddress)
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canClaim.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to claim %.6f ETH in unclaimed rewards?", math.RoundDown(eth.WeiToEth(status.UnclaimedRewards), 6)))) {
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
	fmt.Printf("Successfully claimed %.6f ETH in unclaimed rewards.\n", math.RoundDown(eth.WeiToEth(status.UnclaimedRewards), 6))
	return nil

}
