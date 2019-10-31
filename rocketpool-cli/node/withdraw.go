package node

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Withdraw resources from the node
func withdrawFromNode(c *cli.Context, amount float64, unit string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        CM: true,
        NodeContract: true,
        LoadContracts: []string{"rocketNodeAPI"},
        LoadAbis: []string{"rocketNodeContract"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get args
    amountWei := eth.EthToWei(amount)

    // Check deposit can be withdrawn from node
    canWithdraw, err := node.CanWithdrawFromNode(p, amountWei, unit)
    if err != nil { return err }

    // Check response
    if canWithdraw.InsufficientNodeBalance {
        fmt.Fprintln(p.Output, "Withdrawal amount exceeds available balance on node contract")
    }
    if !canWithdraw.Success {
        return nil
    }

    // Withdraw from node
    withdrawn, err := node.WithdrawFromNode(p, amountWei, unit)
    if err != nil { return err }

    // Print output & return
    if withdrawn.Success {
        fmt.Fprintln(p.Output, fmt.Sprintf("Successfully withdrew %.2f %s from node contract to account", amount, unit))
    }
    return nil

}

