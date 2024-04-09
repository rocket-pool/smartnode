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

func nodeWithdrawRpl(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get node status
	status, err := rp.Api.Node.Status()
	if err != nil {
		return err
	}

	// Get withdrawal mount
	var amountWei *big.Int
	if c.String(amountFlag) == "max" {
		// Set amount to maximum withdrawable amount
		var maxAmount big.Int
		if status.Data.RplStake.Cmp(status.Data.MaximumRplStake) > 0 {
			maxAmount.Sub(status.Data.RplStake, status.Data.MaximumRplStake)
		}
		amountWei = &maxAmount
	} else if c.String(amountFlag) != "" {
		// Parse amount
		withdrawalAmount, err := strconv.ParseFloat(c.String(amountFlag), 64)
		if err != nil {
			return fmt.Errorf("invalid withdrawal amount '%s': %w", c.String(amountFlag), err)
		}
		amountWei = eth.EthToWei(withdrawalAmount)
	} else {
		// Get maximum withdrawable amount
		maxAmount := big.NewInt(0)
		maxAmount.Sub(status.Data.RplStake, status.Data.MaximumRplStake)
		maxAmount.Sub(maxAmount, status.Data.RplLocked)
		if maxAmount.Sign() == 1 {
			// Prompt for maximum amount
			if utils.Confirm(fmt.Sprintf("Would you like to withdraw the maximum amount of staked RPL (%.6f RPL)?", math.RoundDown(eth.WeiToEth(maxAmount), 6))) {
				amountWei = maxAmount
			} else {

				// Prompt for custom amount
				inputAmount := utils.Prompt("Please enter an amount of staked RPL to withdraw:", "^\\d+(\\.\\d+)?$", "Invalid amount")
				withdrawalAmount, err := strconv.ParseFloat(inputAmount, 64)
				if err != nil {
					return fmt.Errorf("invalid withdrawal amount '%s': %w", inputAmount, err)
				}
				amountWei = eth.EthToWei(withdrawalAmount)

			}
		} else {
			fmt.Printf("Cannot withdraw staked RPL - you have %.6f RPL staked, but are not allowed to withdraw below %.6f RPL (%.2f%% collateral).\n",
				math.RoundDown(eth.WeiToEth(status.Data.RplStake), 6),
				math.RoundDown(eth.WeiToEth(status.Data.MaximumRplStake), 6),
				math.RoundDown(eth.WeiToEth(status.Data.MaximumStakeFraction), 2)*100,
			)
			return nil
		}
	}

	// Build the TX
	response, err := rp.Api.Node.WithdrawRpl(amountWei)
	if err != nil {
		return err
	}

	// Verify
	if !response.Data.CanWithdraw {
		fmt.Println("Cannot withdraw staked RPL:")
		if response.Data.InsufficientBalance {
			fmt.Println("The node's staked RPL balance is insufficient.")
		}
		if response.Data.BelowMaxRplStake {
			fmt.Printf("Withdrawing this much RPL would decrease the node below %.2f%% collateral, which is not allowed.", math.RoundDown(eth.WeiToEth(status.Data.MaximumStakeFraction), 2)*100)
			fmt.Println()
		}
		if response.Data.MinipoolsUndercollateralized {
			fmt.Println("Remaining staked RPL is not enough to collateralize the node's minipools.")
		}
		if response.Data.WithdrawalDelayActive {
			fmt.Println("The withdrawal delay period has not passed.")
		}
		if response.Data.HasDifferentRplWithdrawalAddress {
			fmt.Println("The RPL withdrawal address has been set, and is not the node address. RPL can only be withdrawn from the RPL withdrawal address.")
		}
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		fmt.Sprintf("Are you sure you want to withdraw %.6f staked RPL? This may decrease your node's RPL rewards.", math.RoundDown(eth.WeiToEth(amountWei), 6)),
		"RPL withdrawal",
		"Withdrawing RPL...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully withdrew %.6f staked RPL.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil
}
