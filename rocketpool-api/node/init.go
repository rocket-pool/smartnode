package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Initialise the node with a password and an account
func initNode(c *cli.Context, password string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        AM: true,
        PasswordOptional: true,
        NodeAccountOptional: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Init node & print response
    if response, err := node.InitNode(p, password); err != nil {
        return err
    } else {
        api.PrintResponse(p.Output, response)
        return nil
    }

}

