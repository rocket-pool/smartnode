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


func runCanRegisterNode(c *cli.Context) {
    response, err := canRegisterNode(c)
    if err != nil {
        api.PrintResponse(&types.CanRegisterNodeResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func runRegisterNode(c *cli.Context, timezoneLocation string) {
    response, err := registerNode(c, timezoneLocation)
    if err != nil {
        api.PrintResponse(&types.RegisterNodeResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func canRegisterNode(c *cli.Context) (*types.CanRegisterNodeResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := types.CanRegisterNodeResponse{}

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
        return nil, err
    }

    // Update & return response
    response.CanRegister = !(response.AlreadyRegistered || response.RegistrationDisabled)
    return &response, nil

}


func registerNode(c *cli.Context, timezoneLocation string) (*types.RegisterNodeResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := types.RegisterNodeResponse{}

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Register node
    txReceipt, err := node.RegisterNode(rp, timezoneLocation, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash.Hex()

    // Return response
    return &response, nil

}

