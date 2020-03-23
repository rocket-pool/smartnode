package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Export the node account
func exportNodeAccount(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        AM: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get node account
    account, err := node.ExportNodeAccount(p)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, account, "")
    return nil

}

