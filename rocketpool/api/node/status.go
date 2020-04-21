package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Get the node's status
func getNodeStatus(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        Client: true,
        CM: true,
        LoadContracts: []string{"rocketETHToken", "rocketNodeAPI", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
        ClientConn: true,
        ClientSync: true,
        RocketStorage: true,
        PasswordOptional: true,
        NodeAccountOptional: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get node status
    status, err := node.GetNodeStatus(p)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, status, "")
    return nil

}

