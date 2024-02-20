package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func nodeSwapToReth(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckClientStatus(rp)
	if err != nil {
		return err
	}

	// Get swap amount
	var amountWei *big.Int
	
	if c.String("amount") == "max" {

		// get data
		nodeStatus, err := rp.NodeStatus()
		if err != nil {
			return err
		}
		queueStatus, err := rp.QueueStatus()
		if err != nil {
			return err
		}

		var availableAmountWeiWithGasBuffer big.Int
		if availableAmountWeiWithGasBuffer.Sub(nodeStatus.AccountBalances.ETH, eth.EthToWei(0.1)).Sign() == -1 {
			return fmt.Errorf("You need at least 0.1 ETH to be able to pay gas for future transactions.")
		}
		maxAmount := availableAmountWeiWithGasBuffer
		if availableAmountWeiWithGasBuffer.Cmp(queueStatus.MaxDepositPoolBalance.Sub(queueStatus.MaxDepositPoolBalance, queueStatus.DepositPoolBalance)) > 0 {
			maxAmount = *queueStatus.MaxDepositPoolBalance
		}
		amountWei = &maxAmount

	} else if c.String("amount") != "" {

		// Parse amount
		swapAmount, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("Invalid swap amount '%s': %w", c.String("amount"), err)
		}
		amountWei = eth.EthToWei(swapAmount)

	} else {

		nodeStatus, err := rp.NodeStatus()
		if err != nil {
			return err
		}
		queueStatus, err := rp.QueueStatus()
		if err != nil {
			return err
		}

		var maxAmount big.Int
		maxAmount.Sub(nodeStatus.AccountBalances.ETH, eth.EthToWei(0.1))
		if maxAmount.Sign() == 1 && maxAmount.Cmp(queueStatus.MaxDepositPoolBalance.Sub(queueStatus.MaxDepositPoolBalance, queueStatus.DepositPoolBalance)) > 0 {
			maxAmount = *queueStatus.MaxDepositPoolBalance
		}
		
		// Prompt for deposit max amount if possible
		if maxAmount.Sign() > 0 && cliutils.Confirm(fmt.Sprintf("Would you like to swap the maximum available ETH balance (%.6f ETH) (and keep some ETH to pay for future gas costs)?", math.RoundDown(eth.WeiToEth(&maxAmount), 6))){
			
			amountWei = &maxAmount

		} else {

			// Prompt for custom amount
			inputAmount := cliutils.Prompt("Please enter an amount of ETH to swap. Remember that you will need sufficient ETH to execute future transactions!", "^\\d+(\\.\\d+)?$", "Invalid amount")
			swapAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("Invalid swap amount '%s': %w", inputAmount, err)
			}
			amountWei = eth.EthToWei(swapAmount)

		}

	}

	// Check ETH can be swapped
	canStake, err := rp.CanStakeEth(amountWei)
	if err != nil {
		return err
	}
	if !canStake.CanStake {
		fmt.Println("Cannot stake ETH:")
		if canStake.InsufficientBalance {
			fmt.Println("The node's ETH balance is insufficient.")
		}
		if canStake.DepositDisabled {
			fmt.Println("ETH deposits are currently disabled.")
		}
		if canStake.BelowMinStakeAmount {
			fmt.Println("The stake amount is below the minimum accepted value.")
		}
		if canStake.DepositPoolFull {
			fmt.Println("No space left in deposit pool.")
		}
		return nil
	}
	fmt.Println("Stake ETH Gas Info:")
	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canStake.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to stake %.6f ETH for %.6f rETH?", math.RoundDown(eth.WeiToEth(amountWei), 6), math.RoundDown(eth.WeiToEth(canStake.RethAmount), 6)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stake ETH
	stakeResponse, err := rp.StakeEth(amountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Staking ETH...\n")
	cliutils.PrintTransactionHash(rp, stakeResponse.StakeTxHash)
	if _, err = rp.WaitForTransaction(stakeResponse.StakeTxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully staked %.6f ETH in return for %.6f rETH.\n", math.RoundDown(eth.WeiToEth(amountWei), 6), math.RoundDown(eth.WeiToEth(canStake.RethAmount), 6))
	return nil

}
