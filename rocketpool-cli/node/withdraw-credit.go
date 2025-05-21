package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func nodeWithdrawCredit(c *cli.Context) error {

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

	// Get withdrawal amount
	var amountWei *big.Int
	if c.String("amount") == "max" {

		// Get node status
		status, err := rp.NodeStatus()
		if err != nil {
			return err
		}

		// Set amount to maximum withdrawable amount
		amountWei = status.EthOnBehalfBalance

	} else if c.String("amount") != "" {

		// Parse amount
		withdrawalAmount, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("Invalid withdrawal amount '%s': %w", c.String("amount"), err)
		}
		amountWei = eth.EthToWei(withdrawalAmount)

	} else {

		// Get node status
		status, err := rp.NodeStatus()
		if err != nil {
			return err
		}

		// Get maximum withdrawable amount
		maxAmount := status.CreditBalance
		// Prompt for maximum amount
		if prompt.Confirm(fmt.Sprintf("Would you like to withdraw the maximum amount of credit ETH (%.6f ETH)?", math.RoundDown(eth.WeiToEth(maxAmount), 6))) {
			amountWei = maxAmount
		} else {

			// Prompt for custom amount
			inputAmount := prompt.Prompt("Please enter an amount of ETH credit to withdraw:", "^\\d+(\\.\\d+)?$", "Invalid amount")
			withdrawalAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("Invalid withdrawal amount '%s': %w", inputAmount, err)
			}
			amountWei = eth.EthToWei(withdrawalAmount)

		}

	}

	// Check credit can be withdrawn
	canWithdraw, err := rp.CanNodeWithdrawCredit(amountWei)
	if err != nil {
		return err
	}
	if !canWithdraw.CanWithdraw {
		fmt.Println("Cannot withdraw credit:")
		if canWithdraw.InsufficientBalance {
			fmt.Println("The node's credit balance is insufficient.")
		}
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canWithdraw.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to withdraw %.6f of credit?", math.RoundDown(eth.WeiToEth(amountWei), 6)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Withdraw ETH
	response, err := rp.NodeWithdrawCredit(amountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Withdrawing credit...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully withdrew %.6f credit.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil

}
