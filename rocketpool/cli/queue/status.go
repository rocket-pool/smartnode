package queue

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/queue"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Get the deposit queue status
func getQueueStatus(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        LoadContracts: []string{"rocketDepositQueue", "rocketDepositSettings", "rocketMinipoolSettings"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get queue status
    status, err := queue.GetQueueStatus(p)
    if err != nil { return err }

    // Print output & return
    for _, queue := range status.Queues {
        fmt.Fprintln(p.Output, fmt.Sprintf("%s deposit queue has a balance of %.2f ETH (%d chunks)", queue.DurationId, eth.WeiToEth(queue.Balance), queue.Chunks))
    }
    return nil

}

