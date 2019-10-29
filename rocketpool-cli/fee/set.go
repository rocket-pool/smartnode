package fee

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/fee"
    "github.com/rocket-pool/smartnode/shared/services"
)


// Set the target user fee to vote for
func setTargetUserFee(c *cli.Context, feePercent float64) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        DB: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Set target user fee
    response, err := fee.SetTargetUserFee(p, feePercent)
    if err != nil { return err }

    // Print output & return
    if response.Success {
        fmt.Fprintln(p.Output, fmt.Sprintf("Target user fee to vote for successfully set to %.2f%%", feePercent))
    }
    return nil

}

