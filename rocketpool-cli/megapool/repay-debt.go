package megapool

import (
	"fmt"
	"math/big"
	"strconv"

	cliutils "github.com/rocket-pool/smartnode/rocketpool-cli/cli"
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func repayDebt(yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	megapoolDetails, err := rp.MegapoolStatus(false)
	if err != nil {
		return err
	}
	if megapoolDetails.Megapool.NodeDebt != nil && megapoolDetails.Megapool.NodeDebt.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("You have %.6f ETH of megapool debt.\n", math.RoundDown(math.WeiToEth(megapoolDetails.Megapool.NodeDebt), 6))
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

	amountWei := math.EthToWei(amount)
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
	err = gas.AssignMaxFeeAndLimit(canRepay.GasLimits, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if prompt.Declined(yes, "Are you sure you want to repay %.6f ETH of megapool debt?", math.RoundDown(math.WeiToEth(amountWei), 6)) {
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
	fmt.Printf("Successfully repaid %.6f ETH of megapool debt.\n", math.RoundDown(math.WeiToEth(amountWei), 6))
	return nil

}
