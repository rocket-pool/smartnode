package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Initialise the node password
func initNodePassword(c *cli.Context, password string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        PasswordOptional: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check & init password
    passwordSet := node.CanInitNodePassword(p)
    if !passwordSet.HadExistingPassword {
        passwordSet, err = node.InitNodePassword(p, password)
        if err != nil { return err }
    }

    // Get error message
    var message string
    if passwordSet.HadExistingPassword {
        message = "Node password is already set"
    }

    // Print response
    api.PrintResponse(p.Output, passwordSet, message)
    return nil

}

