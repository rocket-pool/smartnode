package megapool

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/urfave/cli"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if Saturn is deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	// Get Megapool status
	status, err := rp.MegapoolStatus()
	if err != nil {
		return err
	}

	// Return if megapool isn't deployed
	if !status.Megapool.Deployed {
		fmt.Println("The node does not have a megapool.")
		return nil
	}

	fmt.Printf("Megapool Address: %s\n", status.Megapool.Address)
	fmt.Printf("Megapool Delegate Address: %s\n", status.Megapool.DelegateAddress)
	fmt.Printf("Megapool Effective Delegate Address: %s\n", status.Megapool.EffectiveDelegateAddress)
	fmt.Printf("Megapool Delegate Expiry Block: %d\n", status.Megapool.DelegateExpiry)
	fmt.Printf("Megapool Deployed: %t\n", status.Megapool.Deployed)
	fmt.Printf("Megapool UseLatestDelegate: %t\n", status.Megapool.UseLatestDelegate)
	fmt.Printf("Megapool Refund Value: %.6f ETH. \n", math.RoundDown(eth.WeiToEth(status.Megapool.RefundValue), 6))
	fmt.Printf("Megapool Pending Rewards: %.6f ETH. \n", math.RoundDown(eth.WeiToEth(status.Megapool.PendingRewards), 6))
	fmt.Printf("Megapool Validator Count: %d \n", status.Megapool.ValidatorCount)
	fmt.Printf("Node Express Ticket Count: %d\n", status.Megapool.NodeExpressTicketCount)

	return nil
}
