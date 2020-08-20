package node

import (
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


func nodeSend(c *cli.Context, amount float64, token string, toAddress common.Address) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Get amount in wei
    amountWei := eth.EthToWei(amount)

    // Check tokens can be sent
    canSend, err := rp.CanNodeSend(amountWei, token)
    if err != nil {
        return err
    }
    if !canSend.CanSend {
        fmt.Println("Cannot send tokens:")
        if canSend.InsufficientBalance {
            fmt.Printf("The node's %s balance is insufficient.\n", token)
        }
        return nil
    }

    // Send tokens
    if _, err := rp.NodeSend(amountWei, token, toAddress); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully sent %.2f %s to %s.\n", eth.WeiToEth(amountWei), token, toAddress.Hex())
    return nil

}

