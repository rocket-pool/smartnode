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

// Config
const (
	stakeRPLWarningMessage = "NOTE: By staking RPL, your node will automatically initialize voting power to itself. If you would like to delegate your on-chain voting power, you should run the command `rocketpool pdao initialize-voting` before staking RPL."
)

func nodeStakeRpl(c *cli.Context) error {

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

	// If hotfix is live and voting isn't initialized, display a warning
	err = warnIfVotingUninitialized(rp, c, stakeRPLWarningMessage)
	if err != nil {
		return nil
	}

	// If a custom nonce is set, print the multi-transaction warning
	if c.GlobalUint64("nonce") != 0 {
		cliutils.PrintMultiTransactionNonceWarning()
	}

	// Check for fixed-supply RPL balance
	rplBalance := *(status.AccountBalances.RPL)
	if status.AccountBalances.FixedSupplyRPL.Cmp(big.NewInt(0)) > 0 {

		// Confirm swapping RPL
		if c.Bool("swap") || cliutils.Confirm(fmt.Sprintf("The node has a balance of %.6f old RPL. Would you like to swap it for new RPL before staking?", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6))) {

			// Check allowance
			allowance, err := rp.GetNodeSwapRplAllowance()
			if err != nil {
				return err
			}

			if allowance.Allowance.Cmp(status.AccountBalances.FixedSupplyRPL) < 0 {
				fmt.Println("Before swapping legacy RPL for new RPL, you must first give the new RPL contract approval to interact with your legacy RPL.")
				fmt.Println("This only needs to be done once for your node.")

				// If a custom nonce is set, print the multi-transaction warning
				if c.GlobalUint64("nonce") != 0 {
					cliutils.PrintMultiTransactionNonceWarning()
				}

				// Calculate max uint256 value
				maxApproval := big.NewInt(2)
				maxApproval = maxApproval.Exp(maxApproval, big.NewInt(256), nil)
				maxApproval = maxApproval.Sub(maxApproval, big.NewInt(1))

				// Get approval gas
				approvalGas, err := rp.NodeSwapRplApprovalGas(maxApproval)
				if err != nil {
					return err
				}
				// Assign max fees
				err = gas.AssignMaxFeeAndLimit(approvalGas.GasInfo, rp, c.Bool("yes"))
				if err != nil {
					return err
				}

				// Prompt for confirmation
				if !(c.Bool("yes") || cliutils.Confirm("Do you want to let the new RPL contract interact with your legacy RPL?")) {
					fmt.Println("Cancelled.")
					return nil
				}

				// Approve RPL for swapping
				response, err := rp.NodeSwapRplApprove(maxApproval)
				if err != nil {
					return err
				}
				hash := response.ApproveTxHash
				fmt.Printf("Approving legacy RPL for swapping...\n")
				cliutils.PrintTransactionHash(rp, hash)
				if _, err = rp.WaitForTransaction(hash); err != nil {
					return err
				}
				fmt.Println("Successfully approved access to legacy RPL.")

				// If a custom nonce is set, increment it for the next transaction
				if c.GlobalUint64("nonce") != 0 {
					rp.IncrementCustomNonce()
				}
			}

			// Check RPL can be swapped
			canSwap, err := rp.CanNodeSwapRpl(status.AccountBalances.FixedSupplyRPL)
			if err != nil {
				return err
			}
			if !canSwap.CanSwap {
				fmt.Println("Cannot swap RPL:")
				if canSwap.InsufficientBalance {
					fmt.Println("The node's old RPL balance is insufficient.")
				}
				return nil
			}
			fmt.Println("RPL Swap Gas Info:")
			// Assign max fees
			err = gas.AssignMaxFeeAndLimit(canSwap.GasInfo, rp, c.Bool("yes"))
			if err != nil {
				return err
			}

			// Prompt for confirmation
			if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to swap %.6f old RPL for new RPL?", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6)))) {
				fmt.Println("Cancelled.")
				return nil
			}

			// Swap RPL
			swapResponse, err := rp.NodeSwapRpl(status.AccountBalances.FixedSupplyRPL)
			if err != nil {
				return err
			}

			fmt.Printf("Swapping old RPL for new RPL...\n")
			cliutils.PrintTransactionHash(rp, swapResponse.SwapTxHash)
			if _, err = rp.WaitForTransaction(swapResponse.SwapTxHash); err != nil {
				return err
			}

			// Log
			fmt.Printf("Successfully swapped %.6f old RPL for new RPL.\n", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6))
			fmt.Println("")

			// If a custom nonce is set, increment it for the next transaction
			if c.GlobalUint64("nonce") != 0 {
				rp.IncrementCustomNonce()
			}

			// Get new account RPL balance
			rplBalance.Add(status.AccountBalances.RPL, status.AccountBalances.FixedSupplyRPL)

		}

	}

	// Get min/max per minipool RPL stake amounts
	rplPrice, err := rp.RplPrice()
	if err != nil {
		return err
	}

	// Get stake amount
	var amountWei *big.Int
	if c.String("amount") == "5" {
		// Set RPL stake amount to 5% of borrowed ETH for one minipool
		amountWei = rplPrice.FivePercentBorrowedRplStake

	} else if c.String("amount") == "10" {

		// Set RPL stake amount to 10% of borrowed ETH for one minipool
		amountWei = rplPrice.TenPercentBorrowedRplStake

	} else if c.String("amount") == "15" {

		// Set RPL stake amount to 15% of borrowed ETH for one minipool
		amountWei = rplPrice.FifteenPercentBorrowedRplStake

	} else if c.String("amount") == "all" {

		// Set amount to node's entire RPL balance
		amountWei = &rplBalance

	} else if c.String("amount") != "" {

		// Parse amount
		stakeAmount, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("Invalid stake amount '%s': %w", c.String("amount"), err)
		}
		amountWei = eth.EthToWei(stakeAmount)

	} else {
		// Get the RPL stake amounts for 5,10,15% borrowed ETH per LEB8
		fivePercentBorrowedRplStake := rplPrice.FivePercentBorrowedRplStake
		tenPercentBorrowedRplStake := rplPrice.TenPercentBorrowedRplStake
		fifteenPercentBorrowedRplStake := rplPrice.FifteenPercentBorrowedRplStake

		// Prompt for amount option
		amountOptions := []string{
			fmt.Sprintf("5%% of borrowed ETH (%.6f RPL) for one minipool?", math.RoundUp(eth.WeiToEth(fivePercentBorrowedRplStake), 6)),
			fmt.Sprintf("10%% of borrowed ETH (%.6f RPL) for one minipool?", math.RoundUp(eth.WeiToEth(tenPercentBorrowedRplStake), 6)),
			fmt.Sprintf("15%% of borrowed ETH (%.6f RPL) for one minipool?", math.RoundUp(eth.WeiToEth(fifteenPercentBorrowedRplStake), 6)),
			fmt.Sprintf("Your entire RPL balance (%.6f RPL)?", math.RoundDown(eth.WeiToEth(&rplBalance), 6)),
			"A custom amount",
		}
		selected, _ := cliutils.Select("Please choose an amount of RPL to stake:", amountOptions)
		switch selected {
		case 0:
			amountWei = fivePercentBorrowedRplStake
		case 1:
			amountWei = tenPercentBorrowedRplStake
		case 2:
			amountWei = fifteenPercentBorrowedRplStake
		case 3:
			amountWei = &rplBalance
		}

		// Prompt for custom amount
		if amountWei == nil {
			inputAmount := cliutils.Prompt("Please enter an amount of RPL to stake:", "^\\d+(\\.\\d+)?$", "Invalid amount")
			stakeAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("Invalid stake amount '%s': %w", inputAmount, err)
			}
			amountWei = eth.EthToWei(stakeAmount)
		}

	}

	// Check allowance
	allowance, err := rp.GetNodeStakeRplAllowance()
	if err != nil {
		return err
	}

	if allowance.Allowance.Cmp(amountWei) < 0 {
		fmt.Println("Before staking RPL, you must first give the staking contract approval to interact with your RPL.")
		fmt.Println("This only needs to be done once for your node.")

		// If a custom nonce is set, print the multi-transaction warning
		if c.GlobalUint64("nonce") != 0 {
			cliutils.PrintMultiTransactionNonceWarning()
		}

		// Calculate max uint256 value
		maxApproval := big.NewInt(2)
		maxApproval = maxApproval.Exp(maxApproval, big.NewInt(256), nil)
		maxApproval = maxApproval.Sub(maxApproval, big.NewInt(1))

		// Get approval gas
		approvalGas, err := rp.NodeStakeRplApprovalGas(maxApproval)
		if err != nil {
			return err
		}
		// Assign max fees
		err = gas.AssignMaxFeeAndLimit(approvalGas.GasInfo, rp, c.Bool("yes"))
		if err != nil {
			return err
		}

		// Prompt for confirmation
		if !(c.Bool("yes") || cliutils.Confirm("Do you want to let the staking contract interact with your RPL?")) {
			fmt.Println("Cancelled.")
			return nil
		}

		// Approve RPL for staking
		response, err := rp.NodeStakeRplApprove(maxApproval)
		if err != nil {
			return err
		}
		hash := response.ApproveTxHash
		fmt.Printf("Approving RPL for staking...\n")
		cliutils.PrintTransactionHash(rp, hash)
		if _, err = rp.WaitForTransaction(hash); err != nil {
			return err
		}
		fmt.Println("Successfully approved staking access to RPL.")

		// If a custom nonce is set, increment it for the next transaction
		if c.GlobalUint64("nonce") != 0 {
			rp.IncrementCustomNonce()
		}
	}

	// Check RPL can be staked
	canStake, err := rp.CanNodeStakeRpl(amountWei)
	if err != nil {
		return err
	}
	if !canStake.CanStake {
		fmt.Println("Cannot stake RPL:")
		if canStake.InsufficientBalance {
			fmt.Println("The node's RPL balance is insufficient.")
		}
		return nil
	}

	fmt.Println("RPL Stake Gas Info:")
	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canStake.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to stake %.6f RPL? You will not be able to unstake this RPL until you exit your validators and close your minipools, or reach %.6f staked RPL (%.0f%% of bonded eth)!",
		math.RoundDown(eth.WeiToEth(amountWei), 6),
		math.RoundDown(eth.WeiToEth(status.MaximumRplStake), 6),
		status.MaximumStakeFraction*100))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Stake RPL
	stakeResponse, err := rp.NodeStakeRpl(amountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Staking RPL...\n")
	cliutils.PrintTransactionHash(rp, stakeResponse.StakeTxHash)
	if _, err = rp.WaitForTransaction(stakeResponse.StakeTxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully staked %.6f RPL.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil

}
