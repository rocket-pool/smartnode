package node

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


func nodeDeposit(c *cli.Context, amount float64) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Get amount in wei
    amountWei := eth.EthToWei(amount)

    // Check deposit can be made
    canDeposit, err := rp.CanNodeDeposit(amountWei)
    if err != nil {
        return err
    }
    if !canDeposit.CanDeposit {
        fmt.Println("Cannot make node deposit:")
        if canDeposit.InsufficientBalance {
            fmt.Println("The node's ETH balance is insufficient.")
        }
        if canDeposit.InvalidAmount {
            fmt.Println("The deposit amount is invalid.")
        }
        if canDeposit.DepositDisabled {
            fmt.Println("Node deposits are currently disabled.")
        }
        return nil
    }

    // Check current network node fee and prompt for minimum
    // TODO: implement
    minNodeFee := 0.0

    // Make deposit
    response, err := rp.NodeDeposit(amountWei, minNodeFee)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("The node deposit of %.2f ETH was made successfully.\n", eth.WeiToEth(amountWei))
    fmt.Printf("A new minipool was created at %s.\n", response.MinipoolAddress.Hex())
    return nil

}

