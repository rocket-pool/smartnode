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
    if c.String("amount") == "max" {

        // Set amount to node fixed-supply RPL balance
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

        // Get maximum swap amount
        status, err := rp.NodeStatus()
        if err != nil {
            return err
        }
        maxAmount := status.AccountBalances.FixedSupplyRPL

        // Prompt for maximum amount
        if cliutils.Confirm(fmt.Sprintf("Would you like to swap your entire old RPL balance (%.6f RPL)?", math.RoundDown(eth.WeiToEth(maxAmount), 6))) {
            amountWei = maxAmount
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

    // Swap RPL
    if _, err := rp.NodeSwapRpl(amountWei); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully swapped %.6f old RPL for new RPL.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
    return nil

}

