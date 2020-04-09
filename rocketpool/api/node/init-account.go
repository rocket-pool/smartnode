package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Can initialise the node account
func canInitNodeAccount(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        AM: true,
        PasswordOptional: true,
        NodeAccountOptional: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check
    canInit := node.CanInitNodeAccount(p)

    // Get error message
    var message string
    if canInit.NodePasswordDidNotExist {
        message = "Node password is not set"
    } else if canInit.HadExistingAccount {
        message = "Node account is already initialized"
    }

    // Print response
    api.PrintResponse(p.Output, canInit, message)
    return nil

}


// Initialise the node account
func initNodeAccount(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        PM: true,
        AM: true,
        PasswordOptional: true,
        NodeAccountOptional: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Check & init account
    accountSet := node.CanInitNodeAccount(p)
    if !accountSet.HadExistingAccount {
        accountSet, err = node.InitNodeAccount(p)
        if err != nil { return err }
    }

    // Get error message
    var message string
    if accountSet.NodePasswordDidNotExist {
        message = "Node password is not set"
    } else if accountSet.HadExistingAccount {
        message = "Node account is already initialized"
    }

    // Print response
    api.PrintResponse(p.Output, accountSet, message)
    return nil

}

