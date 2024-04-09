package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func nodeWithdrawEth(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get withdrawal amount
	var amountWei *big.Int
	if c.String(amountFlag) == "max" {
		// Get node status
		status, err := rp.Api.Node.Status()
		if err != nil {
			return err
		}

		// Set amount to maximum withdrawable amount
		amountWei = status.Data.EthOnBehalfBalance
	} else if c.String(amountFlag) != "" {
		// Parse amount
		withdrawalAmount, err := strconv.ParseFloat(c.String(amountFlag), 64)
		if err != nil {
			return fmt.Errorf("invalid withdrawal amount '%s': %w", c.String(amountFlag), err)
		}
		amountWei = eth.EthToWei(withdrawalAmount)
	} else {
		// Get node status
		status, err := rp.Api.Node.Status()
		if err != nil {
			return err
		}

		// Get maximum withdrawable amount
		maxAmount := status.Data.EthOnBehalfBalance
		// Prompt for maximum amount
		if utils.Confirm(fmt.Sprintf("Would you like to withdraw the maximum amount of staked ETH (%.6f ETH)?", math.RoundDown(eth.WeiToEth(maxAmount), 6))) {
			amountWei = maxAmount
		} else {
			// Prompt for custom amount
			inputAmount := utils.Prompt("Please enter an amount of staked ETH to withdraw:", "^\\d+(\\.\\d+)?$", "Invalid amount")
			withdrawalAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("invalid withdrawal amount '%s': %w", inputAmount, err)
			}
			amountWei = eth.EthToWei(withdrawalAmount)
		}
	}

	// Build the TX
	response, err := rp.Api.Node.WithdrawEth(amountWei)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanWithdraw {
		fmt.Println("Cannot withdraw staked ETH:")
		if response.Data.InsufficientBalance {
			fmt.Println("The node's staked ETH balance is insufficient.")
		}
		if response.Data.HasDifferentPrimaryWithdrawalAddress {
			fmt.Println("The primary withdrawal address has been set, and is not the node address. ETH can only be withdrawn from the primary withdrawal address.")
		}
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to withdraw %.6f ETH?", math.RoundDown(eth.WeiToEth(amountWei), 6)),
		"ETH withdrawal",
		"Withdrawing ETH...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully withdrew %.6f staked ETH.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil
}
