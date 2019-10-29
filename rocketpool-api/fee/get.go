package fee

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/fee"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Get the current user fee
func getUserFee(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        DB: true,
        CM: true,
        LoadContracts: []string{"rocketNodeSettings"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get user fee & print response
    if response, err := fee.GetUserFee(p); err != nil {
        return err
    } else {
        api.PrintResponse(p.Output, response)
        return nil
    }

}

