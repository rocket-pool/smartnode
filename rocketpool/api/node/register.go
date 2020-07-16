package node

import (
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func canRegisterNode(c *cli.Context) error {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Response
    response := &types.CanRegisterNodeResponse{}

    // Sync
    var wg errgroup.Group

    // Check node is not already registered
    wg.Go(func() error {
        nodeAccount, _ := am.GetNodeAccount()
        exists, err := node.GetNodeExists(rp, nodeAccount.Address)
        if err == nil {
            response.AlreadyRegistered = exists
        }
        return err
    })

    // Check node registrations are enabled
    wg.Go(func() error {
        registrationEnabled, err := settings.GetNodeRegistrationEnabled(rp)
        if err == nil {
            response.RegistrationDisabled = !registrationEnabled
        }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return api.PrintResponse(&types.CanRegisterNodeResponse{
            Error: err.Error(),
        })
    }

    // Update & print response
    response.CanRegister = !(response.AlreadyRegistered || response.RegistrationDisabled)
    return api.PrintResponse(response)

}


func registerNode(c *cli.Context, timezoneLocation string) error {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Get txor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return api.PrintResponse(&types.RegisterNodeResponse{
            Error: err.Error(),
        })
    }

    // Register node
    if _, err := node.RegisterNode(rp, timezoneLocation, opts); err != nil {
        return api.PrintResponse(&types.RegisterNodeResponse{
            Error: err.Error(),
        })
    }

    // Print response
    return api.PrintResponse(&types.RegisterNodeResponse{})

}

