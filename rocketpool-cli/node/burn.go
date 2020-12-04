package node

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/math"
)


func nodeBurn(c *cli.Context, amount float64, token string) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get amount in wei
    amountWei := eth.EthToWei(amount)

    // Check tokens can be burned
    canBurn, err := rp.CanNodeBurn(amountWei, token)
    if err != nil {
        return err
    }
    if !canBurn.CanBurn {
        fmt.Println("Cannot burn tokens:")
        if canBurn.InsufficientBalance {
            fmt.Printf("The node's %s balance is insufficient.\n", token)
        }
        if canBurn.InsufficientCollateral {
            fmt.Printf("The %s contract contains insufficient ETH for trade.\n", token)
        }
        return nil
    }

    // Burn tokens
    if _, err := rp.NodeBurn(amountWei, token); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully burned %.6f %s for ETH.\n", math.RoundDown(eth.WeiToEth(amountWei), 6), token)
    return nil

}

