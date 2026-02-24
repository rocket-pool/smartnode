package megapool

import (
	"fmt"
	"strconv"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
	"github.com/urfave/cli"
)

func reduceBond(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	megapoolDetails, err := rp.MegapoolStatus(false)
	if err != nil {
		return err
	}

	fmt.Printf("Current active validators:                       %d\n", megapoolDetails.Megapool.ActiveValidatorCount)
	fmt.Printf("Current megapool bond:                           %.6f ETH\n", math.RoundDown(eth.WeiToEth(megapoolDetails.Megapool.NodeBond), 6))
	fmt.Printf("Current bond requirements for active validators: %.6f ETH\n", math.RoundDown(eth.WeiToEth(megapoolDetails.Megapool.BondRequirement), 6))
	fmt.Println()

	var amount float64
	// If current node bond is higher than the bond requirement, ask if the user wants to reduce the bond
	if megapoolDetails.Megapool.NodeBond.Cmp(megapoolDetails.Megapool.BondRequirement) > 0 {
		maxAmountInEth := eth.WeiToEth(megapoolDetails.Megapool.NodeBond.Sub(megapoolDetails.Megapool.NodeBond, megapoolDetails.Megapool.BondRequirement))
		fmt.Printf("You have %.6f of excess bond.\n", maxAmountInEth)
		if prompt.Confirm(fmt.Sprintf("Do you want to reduce %.6f ETH of your node bond?", maxAmountInEth)) {
			// Convert maxAmountInEth to string
			amount = maxAmountInEth
		} else {
			// Get amount to repay
			amountStr := prompt.Prompt("Enter the amount you want to reduce your bond (in ETH):", "^\\d+(\\.\\d+)?$", "Invalid amount")
			amount, err = strconv.ParseFloat(amountStr, 64)
			if err != nil {
				return fmt.Errorf("Invalid amount '%s': %w\n", amountStr, err)
			}
		}
	} else {
		fmt.Println("Your megapool bond does not exceed the bond requirement, so a bond reduction is not available.")
		return nil
	}

	amountWei := eth.EthToWei(amount)
	// Check megapool debt can be repaid
	canReduceBond, err := rp.CanReduceBond(amountWei)
	if err != nil {
		return err
	}

	if !canReduceBond.CanReduceBond {
		if canReduceBond.NotEnoughBond {
			fmt.Println("Not enough bond for a bond reduction.")
		}
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canReduceBond.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to reduce %.6f of the megapool bond?", math.RoundDown(eth.WeiToEth(amountWei), 6)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Reduce megapool bond
	response, err := rp.ReduceBond(amountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Reducing the megapool bond...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully reduced %.6f of megapool bond.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil

}
