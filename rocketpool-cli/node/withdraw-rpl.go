package node

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func nodeWithdrawRpl(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()
	ec, err := services.GetEthClient(c)
	if err != nil {
		return err
	}

	// Get node status
	status, err := rp.NodeStatus()
	if err != nil {
		return err
	}

	if status.IsSaturnDeployed {
		// Get the latest block time
		latestBlockTimeUnix, err := services.GetEthClientLatestBlockTimestamp(ec)
		if err != nil {
			return err
		}
		latestBlockTime := time.Unix(int64(latestBlockTimeUnix), 0)
		fmt.Print("The RPL withdrawal process has changed in Saturn. It is now a 2-step process:")
		fmt.Println()
		fmt.Print("1. Request to unstake a certain RPL amount;")
		fmt.Println()
		fmt.Printf("2. Wait for the unstaking period to end (currently %s), and then withdraw the RPL.", status.UnstakingPeriodDuration)

		if status.UnstakingRPL.Cmp(big.NewInt(0)) > 0 {

			// Check if unstaking period passed considering the last unstake time
			unstakingPeriodEnd := status.LastRPLUnstakeTime.Add(status.UnstakingPeriodDuration)
			if unstakingPeriodEnd.After(latestBlockTime) {
				fmt.Printf("You have %.2f RPL currently unstaking until %s.\n", unstakingPeriodEnd.Format("2006-01-02 15:04:05"))
			} else {
				if !c.Bool("yes") || prompt.Confirm(fmt.Sprintf("You have %.2f RPL already unstaked. Would you like to withdraw it now?", eth.WeiToEth(status.UnstakingRPL))) {

					// Check RPL can be withdrawn
					canWithdraw, err := rp.CanNodeWithdrawRpl()
					if err != nil {
						return err
					}
					if !canWithdraw.CanWithdraw {
						if canWithdraw.HasDifferentRPLWithdrawalAddress {
							fmt.Println("The RPL withdrawal address has been set, and is not the node address. RPL can only be withdrawn from the RPL withdrawal address.")
						}
					}

					// Assign max fees
					err = gas.AssignMaxFeeAndLimit(canWithdraw.GasInfo, rp, c.Bool("yes"))
					if err != nil {
						return err
					}

					// Prompt for confirmation
					if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to withdraw %.6f staked RPL? This may decrease your node's RPL rewards.", math.RoundDown(eth.WeiToEth(status.UnstakingRPL), 6)))) {
						fmt.Println("Cancelled.")
						return nil
					}

					// Withdraw RPL
					response, err := rp.NodeWithdrawRpl()
					if err != nil {
						return err
					}

					fmt.Printf("Withdrawing RPL...\n")
					cliutils.PrintTransactionHash(rp, response.TxHash)
					if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
						return err
					}

					fmt.Printf("Successfully withdrew %.6f staked RPL.\n", math.RoundDown(eth.WeiToEth(status.UnstakingRPL), 6))
					return nil
				}

			}

		} else {
			fmt.Println("You have %.6f RPL ")

		}

	
	} else { // Saturn not deployed. Run the legacy withdraw command

		// Get withdrawal mount
		var amountWei *big.Int
		if c.String("amount") == "max" {

			// Set amount to maximum withdrawable amount
			var maxAmount big.Int
			if status.RplStake.Cmp(status.MaximumRplStake) > 0 {
				maxAmount.Sub(status.RplStake, status.MaximumRplStake)
			}
			amountWei = &maxAmount

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
			var maxAmount big.Int
			maxAmount.Sub(status.RplStake, status.MaximumRplStake)
			maxAmount.Sub(&maxAmount, status.NodeRPLLocked)
			if maxAmount.Sign() == 1 {
				// Prompt for maximum amount
				if prompt.Confirm(fmt.Sprintf("Would you like to withdraw the maximum amount of staked RPL (%.6f RPL)?", math.RoundDown(eth.WeiToEth(&maxAmount), 6))) {
					amountWei = &maxAmount
				} else {

					// Prompt for custom amount
					inputAmount := prompt.Prompt("Please enter an amount of staked RPL to withdraw:", "^\\d+(\\.\\d+)?$", "Invalid amount")
					withdrawalAmount, err := strconv.ParseFloat(inputAmount, 64)
					if err != nil {
						return fmt.Errorf("Invalid withdrawal amount '%s': %w", inputAmount, err)
					}
					amountWei = eth.EthToWei(withdrawalAmount)

				}
			} else {
				fmt.Printf("Cannot withdraw staked RPL - you have %.6f RPL staked, but are not allowed to withdraw below %.6f RPL (%d%% collateral).\n",
					math.RoundDown(eth.WeiToEth(status.RplStake), 6),
					math.RoundDown(eth.WeiToEth(status.MaximumRplStake), 6),
					uint32(status.MaximumStakeFraction*100),
				)
				return nil
			}

		}

		// Check RPL can be withdrawn
		canWithdraw, err := rp.CanNodeWithdrawLegacyRpl(amountWei)
		if err != nil {
			return err
		}
		if !canWithdraw.CanWithdraw {
			fmt.Println("Cannot withdraw staked RPL:")
			if canWithdraw.InsufficientBalance {
				fmt.Println("The node's staked RPL balance is insufficient.")
			}
			if canWithdraw.BelowMaxRPLStake {
				fmt.Println("Remaining staked RPL is not enough to collateralize the node's minipools.")
			}
			if canWithdraw.WithdrawalDelayActive {
				fmt.Println("The withdrawal delay period has not passed.")
			}
			if canWithdraw.HasDifferentRPLWithdrawalAddress {
				fmt.Println("The RPL withdrawal address has been set, and is not the node address. RPL can only be withdrawn from the RPL withdrawal address.")
			}
		}

		// Assign max fees
		err = gas.AssignMaxFeeAndLimit(canWithdraw.GasInfo, rp, c.Bool("yes"))
		if err != nil {
			return err
		}

		// Prompt for confirmation
		if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to withdraw %.6f staked RPL? This may decrease your node's RPL rewards.", math.RoundDown(eth.WeiToEth(amountWei), 6)))) {
			fmt.Println("Cancelled.")
			return nil
		}

		// Withdraw RPL
		response, err := rp.NodeWithdrawLegacyRpl(amountWei)
		if err != nil {
			return err
		}

		fmt.Printf("Withdrawing RPL...\n")
		cliutils.PrintTransactionHash(rp, response.TxHash)
		if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
			return err
		}

		// Log & return
		fmt.Printf("Successfully withdrew %.6f staked RPL.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
		return nil
	}
	return nil

}
