package megapool

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

func claim(c *cli.Context) error {

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

	megapoolDetails, err := rp.MegapoolStatus()
	if err != nil {
		return err
	}

	if megapoolDetails.Megapool.RefundValue != nil && megapoolDetails.Megapool.RefundValue.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("You have %.6f ETH of megapool refund to claim.\n", math.RoundDown(eth.WeiToEth(megapoolDetails.Megapool.RefundValue), 6))
		if megapoolDetails.Megapool.NodeDebt != nil && megapoolDetails.Megapool.NodeDebt.Cmp(big.NewInt(0)) > 0 {
			fmt.Printf("You have %.6f ETH of node debt to repay. This will be deducted from your refund.\n", math.RoundDown(eth.WeiToEth(megapoolDetails.Megapool.NodeDebt), 6))
		}
	} else {
		fmt.Println("You have no megapool refund to claim.")
		return nil
	}

	canRepay, err := rp.CanClaimMegapoolRefund()
	if err != nil {
		return err
	}

	if !canRepay.CanClaim {
		fmt.Println("You cannot claim a megapool refund at this time.")
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canRepay.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to claim your megapool refund?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Repay megapool debt
	response, err := rp.ClaimMegapoolRefund()
	if err != nil {
		return err
	}

	fmt.Printf("Claiming megapool refund...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully claimed megapool refund.\n")
	return nil

}
