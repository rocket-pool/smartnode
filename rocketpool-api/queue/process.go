package queue

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/queue"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
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
    if !canProcess.Success {
        var message string
        if canProcess.InvalidStakingDuration {
            message = "The specified staking duration is invalid or disabled"
        } else if canProcess.InsufficientBalance {
            message = "The queue has an insufficient ETH balance for processing"
        } else if canProcess.NoAvailableNodes {
            message = "No minipools are currently available for assignment"
        }
        api.PrintResponse(p.Output, canProcess, message)
        return nil
    }

    // Process deposit queue
    processed, err := queue.ProcessQueue(p, durationId)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, processed, "")
    return nil

}

