package node

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/node"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


// Can register the node with Rocket Pool
func canRegisterNode(c *cli.Context) error {

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

    // Get error message
    var message string
    if canRegister.HadExistingContract {
        message = "Node is already registered with Rocket Pool"
    } else if canRegister.RegistrationsDisabled {
        message = "Node registrations are currently disabled in Rocket Pool"
    } else if canRegister.InsufficientAccountBalance {
        message = "Node account has insufficient ETH balance for registration"
    }

    // Print response
    api.PrintResponse(p.Output, canRegister, message)
    return nil

}


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
        var message string
        if canRegister.HadExistingContract {
            message = "Node is already registered with Rocket Pool"
        } else if canRegister.RegistrationsDisabled {
            message = "Node registrations are currently disabled in Rocket Pool"
        } else if canRegister.InsufficientAccountBalance {
            message = "Node account has insufficient ETH balance for registration"
        }
        api.PrintResponse(p.Output, canRegister, message)
        return nil
    }

    // Register node
    registered, err := node.RegisterNode(p, timezone)
    if err != nil { return err }

    // Print response
    api.PrintResponse(p.Output, registered, "")
    return nil

}

