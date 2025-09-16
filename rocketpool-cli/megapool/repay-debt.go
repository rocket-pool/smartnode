package megapool

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/urfave/cli"
)

func repayDebt(c *cli.Context) error {

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
	if megapoolDetails.Megapool.NodeDebt != nil && megapoolDetails.Megapool.NodeDebt.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("You have %.6f ETH of megapool debt.\n", math.RoundDown(eth.WeiToEth(megapoolDetails.Megapool.NodeDebt), 6))
	} else {
		fmt.Println("You have no megapool debt.")
		return nil
	}

	// Get amount to repay
	amountStr := prompt.Prompt("Enter the amount of megapool debt to repay (in ETH):", "^\\d+(\\.\\d+)?$", "Invalid amount")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("Invalid amount '%s': %w\n", amountStr, err)
	}

	amountWei := eth.EthToWei(amount)
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
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to repay %.6f ETH of megapool debt?", math.RoundDown(eth.WeiToEth(amountWei), 6)))) {
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
	fmt.Printf("Successfully repaid %.6f ETH of megapool debt.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil

}
