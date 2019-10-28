package fee

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/fee"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Set the target user fee to vote for
func setTargetUserFee(c *cli.Context, feePercent float64) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        DB: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Set target user fee & print response
    if response, err := fee.SetTargetUserFee(p, feePercent); err != nil {
        return err
    } else {
        api.PrintResponse(p.Output, response)
        return nil
    }

}

