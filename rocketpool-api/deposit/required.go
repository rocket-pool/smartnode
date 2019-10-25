package deposit

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Get the current deposit RPL requirement
func getRplRequired(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        LoadContracts: []string{"rocketMinipoolSettings", "rocketNodeAPI", "rocketPool"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get RPL requirement & print response
    if response, err := deposit.GetRplRequired(p); err != nil {
        return err
    } else {
        api.PrintResponse(p.Output, response)
        return nil
    }

}

