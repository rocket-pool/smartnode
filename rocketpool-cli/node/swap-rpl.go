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

func nodeSwapRpl(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get swap amount
	var amountWei *big.Int
	if c.String("amount") == "all" {

		// Set amount to node's entire fixed-supply RPL balance
		status, err := rp.NodeStatus()
		if err != nil {
			return err
		}
		amountWei = status.AccountBalances.FixedSupplyRPL

	} else if c.String("amount") != "" {

		// Parse amount
		swapAmount, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("Invalid swap amount '%s': %w", c.String("amount"), err)
		}
		amountWei = eth.EthToWei(swapAmount)

	} else {

		// Get entire fixed-supply RPL balance amount
		status, err := rp.NodeStatus()
		if err != nil {
			return err
		}
		entireAmount := status.AccountBalances.FixedSupplyRPL

		// Prompt for entire amount
		if prompt.Confirm(fmt.Sprintf("Would you like to swap your entire old RPL balance (%.6f RPL)?", math.RoundDown(eth.WeiToEth(entireAmount), 6))) {
			amountWei = entireAmount
		} else {

			// Prompt for custom amount
			inputAmount := prompt.Prompt("Please enter an amount of old RPL to swap:", "^\\d+(\\.\\d+)?$", "Invalid amount")
			swapAmount, err := strconv.ParseFloat(inputAmount, 64)
			if err != nil {
				return fmt.Errorf("Invalid swap amount '%s': %w", inputAmount, err)
			}
			amountWei = eth.EthToWei(swapAmount)

		}

	}

	// Check allowance
	allowance, err := rp.GetNodeSwapRplAllowance()
	if err != nil {
		return err
	}

	if allowance.Allowance.Cmp(amountWei) < 0 {
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
		if !(c.Bool("yes") || prompt.Confirm("Do you want to let the new RPL contract interact with your legacy RPL?")) {
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
	canSwap, err := rp.CanNodeSwapRpl(amountWei)
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
	if !(c.Bool("yes") || prompt.Confirm(fmt.Sprintf("Are you sure you want to swap %.6f old RPL for new RPL?", math.RoundDown(eth.WeiToEth(amountWei), 6)))) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Swap RPL
	swapResponse, err := rp.NodeSwapRpl(amountWei)
	if err != nil {
		return err
	}

	fmt.Printf("Swapping old RPL for new RPL...\n")
	cliutils.PrintTransactionHash(rp, swapResponse.SwapTxHash)
	if _, err = rp.WaitForTransaction(swapResponse.SwapTxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Successfully swapped %.6f old RPL for new RPL.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil

}
