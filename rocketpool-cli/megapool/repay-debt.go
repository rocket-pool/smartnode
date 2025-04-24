package megapool

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/urfave/cli"
)

func repayDebt(c *cli.Context, amount float64) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Convert amount to wei
	amountWei := eth.EthToWei(amount)

	// Check if Saturn is already deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	// Check megapool debt can be repaid
	canRepay, err := rp.CanRepayDebt(amountWei)
	if err != nil {
		return err
	}

	if !canRepay.CanRepay {
		if canRepay.NotEnoughDebt {
			fmt.Println("Not enough megapool debt to repay.")
		}
		if canRepay.NotEnoughBalance {
			fmt.Println("Not enough balance to repay megapool debt.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canRepay.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to repay %.6f of megapool debt?", math.RoundDown(eth.WeiToEth(amountWei), 6)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Repay megapool debt
	response, err := rp.RepayDebt(amountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Repaying megapool debt...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully repaid %.6f of megapool debt.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil

}
