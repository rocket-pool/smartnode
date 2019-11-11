package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Register the node with Rocket Pool
func registerNode(c *cli.Context, timezone string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketNodeAPI", "rocketNodeSettings"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check node can be registered
    canRegister, err := node.CanRegisterNode(p)
    if err != nil { return err }

    // Check response
    if !canRegister.Success {
        api.PrintResponse(p.Output, canRegister)
        return nil
    }

    // Register node
    registered, err := node.RegisterNode(p, timezone)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, registered)
    return nil

}

