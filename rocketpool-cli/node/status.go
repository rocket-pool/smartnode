package node

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/math"
)


func getStatus(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get node status
    status, err := rp.NodeStatus()
    if err != nil {
        return err
    }

    // Print & return
    fmt.Printf("The node %s has a balance of %.6f ETH and %.6f nETH.\n", status.AccountAddress.Hex(), math.RoundDown(eth.WeiToEth(status.Balances.ETH), 6), math.RoundDown(eth.WeiToEth(status.Balances.NETH), 6))
    if status.Registered {
        fmt.Printf("The node is registered with Rocket Pool with a timezone location of %s.\n", status.TimezoneLocation)
        if status.Trusted {
            fmt.Println("The node is trusted - it can create empty minipools and will perform watchtower duties.")
        }
        if status.MinipoolCounts.Total > 0 {
            fmt.Printf("The node has a total of %d minipool(s):\n", status.MinipoolCounts.Total)
        } else {
            fmt.Println("The node does not have any minipools yet.")
        }
        if status.MinipoolCounts.Initialized > 0 {
            fmt.Printf("- %d initialized\n", status.MinipoolCounts.Initialized)
        }
        if status.MinipoolCounts.Prelaunch > 0 {
            fmt.Printf("- %d at prelaunch\n", status.MinipoolCounts.Prelaunch)
        }
        if status.MinipoolCounts.Staking > 0 {
            fmt.Printf("- %d staking\n", status.MinipoolCounts.Staking)
        }
        if status.MinipoolCounts.Withdrawable > 0 {
            fmt.Printf("- %d withdrawable\n", status.MinipoolCounts.Withdrawable)
        }
        if status.MinipoolCounts.Dissolved > 0 {
            fmt.Printf("- %d dissolved\n", status.MinipoolCounts.Dissolved)
        }
        if status.MinipoolCounts.RefundAvailable > 0 {
            fmt.Printf("* %d minipools have refunds available!\n", status.MinipoolCounts.RefundAvailable)
        }
        if status.MinipoolCounts.WithdrawalAvailable > 0 {
            fmt.Printf("* %d minipools are ready for withdrawal!\n", status.MinipoolCounts.WithdrawalAvailable)
        }
        if status.MinipoolCounts.CloseAvailable > 0 {
            fmt.Printf("* %d dissolved minipools can be closed!\n", status.MinipoolCounts.CloseAvailable)
        }
    } else {
        fmt.Println("The node is not registered with Rocket Pool.")
    }
    return nil

}

