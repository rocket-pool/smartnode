package node

import (
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/math"
)


func nodeSend(c *cli.Context, amount float64, token string, toAddress common.Address) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
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

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to send %.6f %s to %s? This action cannot be undone!", math.RoundDown(eth.WeiToEth(amountWei), 6), token, toAddress.Hex()))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Send tokens
    if _, err := rp.NodeSend(amountWei, token, toAddress); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully sent %.6f %s to %s.\n", math.RoundDown(eth.WeiToEth(amountWei), 6), token, toAddress.Hex())
    return nil

}

