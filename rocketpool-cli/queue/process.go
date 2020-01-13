package queue

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/queue"
    "github.com/rocket-pool/smartnode/shared/services"
)


// Process a deposit queue
func processQueue(c *cli.Context, durationId string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketDepositQueue", "rocketDepositSettings", "rocketMinipoolSettings", "rocketNode"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check deposit queue can be processed
    canProcess, err := queue.CanProcessQueue(p, durationId)
    if err != nil { return err }

    // Check response
    if canProcess.InvalidStakingDuration {
        fmt.Fprintln(p.Output, fmt.Sprintf("The staking duration '%s' is invalid", durationId))
        return nil
    }
    if canProcess.InsufficientBalance {
        fmt.Fprintln(p.Output, fmt.Sprintf("The %s deposit queue does not contain enough ETH for assignment", durationId))
    }
    if canProcess.NoAvailableNodes {
        fmt.Fprintln(p.Output, fmt.Sprintf("No minipools staking for %s are available for assignment", durationId))
    }
    if !canProcess.Success {
        return nil
    }

    // Process deposit queue
    processed, err := queue.ProcessQueue(p, durationId)
    if err != nil { return err }

    // Print output & return
    if processed.Success {
        fmt.Fprintln(p.Output, fmt.Sprintf("The %s deposit queue was processed successfully", durationId))
    }
    return nil

}

