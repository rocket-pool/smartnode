package node

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const TimeFormat = "2006-01-02, 15:04 -0700 MST"

func nodeWithdrawRpl(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get node status
	status, err := rp.NodeStatus()
	if err != nil {
		return err
	}

	var unstakingPeriodEnd time.Time

	if status.IsSaturnDeployed {
		days := int(status.UnstakingPeriodDuration.Hours()) / 24
		hours := int(status.UnstakingPeriodDuration.Hours()) % 24
		var unstakingDurationString string
		if hours > 0 {
			unstakingDurationString = fmt.Sprintf("%d days, %d hours", days, hours)
		} else {
			unstakingDurationString = fmt.Sprintf("%d days", days)
		}

		fmt.Print("The RPL withdrawal process has changed in Saturn. It is now a 2-step process:")
		fmt.Println()
		fmt.Print("1. Request to unstake a certain RPL amount;")
		fmt.Println()
		fmt.Printf("2. Wait for the unstaking period to end (currently %s%s%s), and then withdraw the RPL.", colorYellow, unstakingDurationString, colorReset)
		fmt.Println()

		fmt.Println()
		fmt.Printf("Your node has %.6f RPL on its legacy stake (previously associated to minipools) and %.6f RPL staked on its megapool.", math.RoundDown(eth.WeiToEth(status.RplStakeLegacy), 6), math.RoundDown(eth.WeiToEth(status.RplStakeMegapool), 6))
		fmt.Println()
		fmt.Printf("Your node currently has %.6f RPL locked on pDAO proposals.", math.RoundDown(eth.WeiToEth(status.NodeRPLLocked), 6))
		fmt.Println()
		fmt.Printf("Your node's RPL withdrawal address is %s%s%s.\n", colorBlue, status.RPLWithdrawalAddress.String(), colorReset)
		fmt.Println()

		// Check if the node has unstaking RPL and if the unstaking period passed considering the last unstake time
		hasUnstakingRPL := status.UnstakingRPL.Cmp(big.NewInt(0)) > 0
		unstakingPeriodEnd = status.LastRPLUnstakeTime.Add(status.UnstakingPeriodDuration)
		var cooldownPassed bool
		if unstakingPeriodEnd.Before(status.LatestBlockTime) {
			cooldownPassed = true
		}
		timeUntilUnstakingPeriodEnd := time.Until(unstakingPeriodEnd).Round(time.Second)

		// Print unstaking RPL details
		if !cooldownPassed && hasUnstakingRPL {
			fmt.Printf("You have %.6f RPL currently unstaking until %s (%s from now).\n", math.RoundDown(eth.WeiToEth(status.UnstakingRPL), 6), unstakingPeriodEnd.Format(TimeFormat), timeUntilUnstakingPeriodEnd.String())
		} else {
			fmt.Printf("You have %.6f RPL unstaked and ready to be withdrawn to your RPL withdrawal address.\n", eth.WeiToEth(status.UnstakingRPL))
		}

		// Prompt for a selection
		options := []string{
			"withdraw unstaked RPL",
			"request to unstake RPL",
			"request to unstake legacy RPL",
		}
		selected, _ := prompt.Select("Please select one of the three options below.\n", options)

		// Selection 1
		if options[selected] == "withdraw unstaked RPL" {
			// Check if RPL can be withdrawn and get gas info
			if !cooldownPassed || !hasUnstakingRPL {
				fmt.Println("You have no RPL eligible to be withdrawn.")
				return nil
			}
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

		// Helper to print warnings
		notifyUnstakingRPLStatus := func() {
			// Inform users that their unstaked RPL will be withdrawn before staked RPL is moved to unstaking
			if cooldownPassed && hasUnstakingRPL {
				fmt.Printf("You have %.6f RPL unstaked and ready to be withdrawn to your RPL withdrawal address. Requesting to unstake more RPL will automatically withdraw %.6f RPL to the RPL withdrawal address.\n", eth.WeiToEth(status.UnstakingRPL), eth.WeiToEth(status.UnstakingRPL))
				fmt.Println()
			}
			// Inform users that the unstaking period will reset if they make another unstaking request
			if !cooldownPassed && hasUnstakingRPL {
				fmt.Printf("You have %.6f RPL currently unstaking until %s (%s from now).\n", math.RoundDown(eth.WeiToEth(status.UnstakingRPL), 6), unstakingPeriodEnd.Format(TimeFormat), timeUntilUnstakingPeriodEnd.String())
				fmt.Printf("%sRequesting to unstake additional RPL will reset the unstaking period.\n%s", colorYellow, colorReset)
				fmt.Printf("%sThe unstaking period is %s.\n%s", colorYellow, unstakingDurationString, colorReset)

				if !prompt.Confirm("Are you sure you would like to continue?") {
					os.Exit(0)
				}
				fmt.Println()
			}
		}

		// Selection 2
		if options[selected] == "request to unstake RPL" {
			if status.RplStakeMegapool.Cmp(big.NewInt(0)) == 0 {
				fmt.Println("You have no staked RPL eligible to be unstaked.")
				return nil
			}

			// Get the maximum withdrawable amount for megapool staked rpl
			var maxAmount big.Int
			var amountWei *big.Int
			withdrawableFromLocked := new(big.Int).Sub(status.TotalRplStake, status.NodeRPLLocked)
			withdrawableFromLegacy := new(big.Int).Sub(status.TotalRplStake, status.RplStakeLegacy)

			// maxAmount = min(withdrawableFromLocked, withdrawableFromLegacy, RplStakeMegapool)
			if withdrawableFromLocked.Cmp(withdrawableFromLegacy) < 0 {
				maxAmount.Set(withdrawableFromLocked)
			} else {
				maxAmount.Set(withdrawableFromLegacy)
			}
			if status.RplStakeMegapool.Cmp(&maxAmount) < 0 {
				maxAmount.Set(status.RplStakeMegapool)
			}

			// Print warning messages if applicable
			notifyUnstakingRPLStatus()

			fmt.Printf("You have %.6f RPL staked on your megapool and can request to unstake up to %.6f RPL.\n", math.RoundDown(eth.WeiToEth(status.RplStakeMegapool), 6), math.RoundDown(eth.WeiToEth(&maxAmount), 6))
			// Prompt for maximum amount
			if prompt.Confirm("Would you like to unstake the maximum amount of staked RPL?") {
				amountWei = &maxAmount
			} else {
				// Prompt for custom amount
				inputAmount := prompt.Prompt("Please enter an amount of staked RPL to unstake:", "^\\d+(\\.\\d+)?$", "Invalid amount")
				withdrawalAmount, err := strconv.ParseFloat(inputAmount, 64)
				if err != nil {
					return fmt.Errorf("Invalid unstake amount '%s': %w", inputAmount, err)
				}
				amountWei = eth.EthToWei(withdrawalAmount)
			}

			// Check if RPL can be unstaked
			canWithdraw, err := rp.CanNodeUnstakeRpl(amountWei)
			if err != nil {
				return err
			}
			if !canWithdraw.CanUnstake {
				fmt.Println("Cannot unstake RPL:")
				if canWithdraw.InsufficientBalance {
					fmt.Println("The node's staked RPL balance is insufficient.")
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
			if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to unstake %.6f RPL?", math.RoundDown(eth.WeiToEth(amountWei), 6)))) {
				fmt.Println("Cancelled.")
				return nil
			}

			// Request to unstake RPL
			response, err := rp.NodeUnstakeRpl(amountWei)
			if err != nil {
				return err
			}

			fmt.Printf("Unstaking RPL...\n")
			cliutils.PrintTransactionHash(rp, response.TxHash)
			if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
				return err
			}

			// Log & return
			fmt.Printf("Successfully unstaked %.6f RPL.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
			return nil
		}

		// Selection 3
		if options[selected] == "request to unstake legacy RPL" {
			if status.RplStakeLegacy.Cmp(big.NewInt(0)) == 0 {
				fmt.Println("You have no legacy RPL eligible to be unstaked.")
				return nil
			}
			fmt.Println("Unstaking legacy RPL follows the same 2-step process as unstaking megapool staked RPL.")
			fmt.Printf("Unstaked legacy RPL can be withdrawn after an unstaking period of %s%s%s.\n", colorYellow, unstakingDurationString, colorReset)
			fmt.Println()

			// Get the maximum withdrawable amount based on constraints
			// defined by decreaseNodeLegacyRPLStake in RocketNodeStaking.sol
			var maxAmount big.Int
			var amountWei *big.Int
			withdrawableFromLegacy := new(big.Int).Sub(status.RplStakeLegacy, status.RplStakeThreshold)
			withdrawableFromTotal := new(big.Int).Sub(status.TotalRplStake, status.NodeRPLLocked)
			if withdrawableFromLegacy.Cmp(withdrawableFromTotal) < 0 {
				maxAmount.Set(withdrawableFromLegacy)
			} else {
				maxAmount.Set(withdrawableFromTotal)
			}

			if maxAmount.Sign() == 1 {
				// Print warning messages if applicable
				notifyUnstakingRPLStatus()

				fmt.Printf("You have %.6f legacy RPL and can request to unstake up to %.6f RPL.\n", math.RoundDown(eth.WeiToEth(status.RplStakeLegacy), 6), math.RoundDown(eth.WeiToEth(&maxAmount), 6))
				// Prompt for maximum amount
				if prompt.Confirm(fmt.Sprintf("Would you like to unstake the maximum amount of legacy RPL (%.6f RPL)?", math.RoundDown(eth.WeiToEth(&maxAmount), 6))) {
					amountWei = &maxAmount
				} else {
					// Prompt for custom amount
					inputAmount := prompt.Prompt("Please enter an amount of legacy RPL to unstake:", "^\\d+(\\.\\d+)?$", "Invalid amount")
					withdrawalAmount, err := strconv.ParseFloat(inputAmount, 64)
					if err != nil {
						return fmt.Errorf("Invalid withdrawal amount '%s': %w", inputAmount, err)
					}
					amountWei = eth.EthToWei(withdrawalAmount)
				}
			} else {
				fmt.Printf("Cannot unstake legacy RPL - you have %.6f legacy RPL, but are not allowed to unstake below %.6f RPL (%d%% of borrowed ETH).\n",
					math.RoundDown(eth.WeiToEth(status.RplStakeLegacy), 6),
					math.RoundDown(eth.WeiToEth(status.RplStakeThreshold), 6),
					uint32(status.RplStakeThresholdFraction*100),
				)
				return nil
			}

			// Check legacy RPL can be unstaked
			canUnstakeLegacyRpl, err := rp.CanNodeUnstakeLegacyRpl(amountWei)
			if err != nil {
				return err
			}

			if !canUnstakeLegacyRpl.CanUnstake {
				if canUnstakeLegacyRpl.InsufficientBalance {
					fmt.Println("The node's staked RPL balance is insufficient.")
				}
				if canUnstakeLegacyRpl.BelowMaxRPLStake {
					fmt.Println("Remaining staked RPL is not enough to collateralize the node's minipools.")
				}
				if canUnstakeLegacyRpl.HasDifferentRPLWithdrawalAddress {
					fmt.Println("The RPL withdrawal address has been set, and is not the node address. RPL can only be withdrawn from the RPL withdrawal address.")
				}

			}

			// Assign max fees
			err = gas.AssignMaxFeeAndLimit(canUnstakeLegacyRpl.GasInfo, rp, c.Bool("yes"))
			if err != nil {
				return err
			}

			// Prompt for confirmation
			if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to unstake %.6f legacy RPL? This may decrease your node's RPL rewards.", math.RoundDown(eth.WeiToEth(amountWei), 6)))) {
				fmt.Println("Cancelled.")
				return nil
			}

			// Withdraw RPL
			response, err := rp.NodeUnstakeLegacyRpl(amountWei)
			if err != nil {
				return err
			}

			fmt.Printf("Unstaking legacy RPL...\n")
			cliutils.PrintTransactionHash(rp, response.TxHash)
			if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
				return err
			}

			// Log & return
			fmt.Printf("Successfully unstaked %.6f legacy RPL.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
			return nil

		}

	}

	// Saturn not deployed. Run the legacy withdraw command

	// Get withdrawal mount
	var amountWei *big.Int
	if c.String("amount") == "max" {

		// Set amount to maximum withdrawable amount
		var maxAmount big.Int
		if status.TotalRplStake.Cmp(status.RplStakeThreshold) > 0 {
			maxAmount.Sub(status.TotalRplStake, status.RplStakeThreshold)
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
		maxAmount.Sub(status.TotalRplStake, status.RplStakeThreshold)
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
				math.RoundDown(eth.WeiToEth(status.TotalRplStake), 6),
				math.RoundDown(eth.WeiToEth(status.RplStakeThreshold), 6),
				uint32(status.RplStakeThresholdFraction*100),
			)
			return nil
		}

	}

	// Check RPL can be withdrawn
	canWithdraw, err := rp.CanNodeWithdrawRplV1_3_1(amountWei)
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
	response, err := rp.NodeWithdrawRplV1_3_1(amountWei)
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
