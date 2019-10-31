package node

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Get the node's status
func getNodeStatus(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketETHToken", "rocketNodeAPI", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get node status
    status, err := node.GetNodeStatus(p)
    if err != nil { return err }

    // Print output & return
    fmt.Fprintln(p.Output, fmt.Sprintf(
        "Node account %s has a balance of %.2f ETH, %.2f rETH and %.2f RPL",
        status.AccountAddress.Hex(),
        eth.WeiToEth(status.AccountBalanceEtherWei),
        eth.WeiToEth(status.AccountBalanceRethWei),
        eth.WeiToEth(status.AccountBalanceRplWei)))
    if status.Registered {
        fmt.Fprintln(p.Output, fmt.Sprintf(
            "Node registered with Rocket Pool with contract at %s, timezone '%s' and a balance of %.2f ETH and %.2f RPL",
            status.ContractAddress.Hex(),
            status.Timezone,
            eth.WeiToEth(status.ContractBalanceEtherWei),
            eth.WeiToEth(status.ContractBalanceRplWei)))
        if status.Trusted {
            fmt.Fprintln(p.Output, "Node is a trusted Rocket Pool node and will perform watchtower duties")
        }
        if !status.Active {
            fmt.Fprintln(p.Output, "Node has been marked inactive after failing to check in, and will not receive user deposits!")
            fmt.Fprintln(p.Output, "Please check smart node daemon status with `rocketpool service stats`")
        }
    } else {
        fmt.Fprintln(p.Output, "Node is not registered with Rocket Pool")
    }
    return nil

}

