package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)


func nodeSwapRpl(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
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
        if cliutils.Confirm(fmt.Sprintf("Would you like to swap your entire old RPL balance (%.6f RPL)?", math.RoundDown(eth.WeiToEth(entireAmount), 6))) {
            amountWei = entireAmount
        } else {

            // Prompt for custom amount
            inputAmount := cliutils.Prompt("Please enter an amount of old RPL to swap:", "^\\d+(\\.\\d+)?$", "Invalid amount")
            swapAmount, err := strconv.ParseFloat(inputAmount, 64)
            if err != nil {
                return fmt.Errorf("Invalid swap amount '%s': %w", inputAmount, err)
            }
            amountWei = eth.EthToWei(swapAmount)

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

    // Approve RPL for swapping
    response, err := rp.NodeSwapRplApprove(amountWei)
    if err != nil {
        return err
    }
    hash := response.ApproveTxHash
    fmt.Printf("Approving old RPL for swap...\n")
    cliutils.PrintTransactionHashNoCancel(rp, hash)
    
    // Swap RPL
    swapResponse, err := rp.NodeSwapRpl(amountWei, hash)
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

