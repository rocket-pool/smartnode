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

        // Get node status
        status, err := rp.NodeStatus()
        if err != nil {
            return err
        }

        // Set amount to node old RPL balance
        amountWei = status.AccountBalances.FixedSupplyRPL

    } else if c.String("amount") != "" {

        // Parse amount
        swapAmount, err := strconv.ParseFloat(c.String("amount"), 64)
        if err != nil {
            return fmt.Errorf("Invalid swap amount '%s': %w", c.String("amount"), err)
        }
        amountWei = eth.EthToWei(swapAmount)

    } else {

        // Prompt for amount
        // TODO: implement

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

