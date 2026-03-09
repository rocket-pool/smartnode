package node

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
)

func nodeWithdrawEth(amount string, yes bool) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get withdrawal amount
	var amountWei *big.Int
	if amount == "max" {

		// Get node status
		status, err := rp.NodeStatus()
		if err != nil {
			return err
		}

		// Set amount to maximum withdrawable amount
		amountWei = status.EthOnBehalfBalance

	} else if amount != "" {

		// Parse amount
		withdrawalAmount, err := strconv.ParseFloat(amount, 64)
		if err != nil {
			return fmt.Errorf("Invalid withdrawal amount '%s': %w", amount, err)
		}
		amountWei = eth.EthToWei(withdrawalAmount)

	} else {

		// Get node status
		status, err := rp.NodeStatus()
		if err != nil {
			return err
		}

		// Get maximum withdrawable amount
		maxAmount := status.EthOnBehalfBalance
		// Prompt for maximum amount
		if prompt.Confirm("Would you like to withdraw the maximum amount of staked ETH (%.6f ETH)?", math.RoundDown(eth.WeiToEth(maxAmount), 6)) {
			amountWei = maxAmount
		} else {

			// Prompt for custom amount
			inputAmount := prompt.Prompt("Please enter an amount of staked ETH to withdraw:", "^\\d+(\\.\\d+)?$", "Invalid amount")
			withdrawalAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("Invalid withdrawal amount '%s': %w", inputAmount, err)
			}
			amountWei = eth.EthToWei(withdrawalAmount)

		}

	}

	// Check ETH can be withdrawn
	canWithdraw, err := rp.CanNodeWithdrawEth(amountWei)
	if err != nil {
		return err
	}
	if !canWithdraw.CanWithdraw {
		fmt.Println("Cannot withdraw staked ETH:")
		if canWithdraw.InsufficientBalance {
			fmt.Println("The node's staked ETH balance is insufficient.")
		}
		if canWithdraw.HasDifferentWithdrawalAddress {
			fmt.Println("The primary withdrawal address has been set, and is not the node address. ETH can only be withdrawn from the primary withdrawal address.")
		}
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canWithdraw.GasInfo, rp, yes)
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(yes || prompt.Confirm("Are you sure you want to withdraw %.6f ETH?", math.RoundDown(eth.WeiToEth(amountWei), 6))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Withdraw ETH
	response, err := rp.NodeWithdrawEth(amountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Withdrawing ETH...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully withdrew %.6f staked ETH.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil

}
