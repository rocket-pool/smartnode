package minipool

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/minipool"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Get the node's minipool statuses
func getMinipoolStatus(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        CM: true,
        LoadContracts: []string{"rocketPoolToken", "utilAddressSetStorage"},
        LoadAbis: []string{"rocketMinipool"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get minipool statuses & print response
    if response, err := minipool.GetMinipoolStatus(p); err != nil {
        return err
    } else {
        api.PrintResponse(p.Output, response)
        return nil
    }

}

