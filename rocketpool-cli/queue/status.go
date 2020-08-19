package queue

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


func getStatus(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Get queue status
    status, err := rp.QueueStatus()
    if err != nil {
        return err
    }

    // Print & return
    fmt.Printf("The deposit pool has a balance of %.2f ETH.\n", eth.WeiToEth(status.DepositPoolBalance))
    fmt.Printf("There are %d available minipools with a total capacity of %.2f ETH.\n", status.MinipoolQueueLength, eth.WeiToEth(status.MinipoolQueueCapacity))
    return nil

}

