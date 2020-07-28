package node

import (
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/settings"
    "github.com/urfave/cli"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canRegisterNode(c *cli.Context) (*api.CanRegisterNodeResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanRegisterNodeResponse{}

    // Sync
    var wg errgroup.Group

    // Check node is not already registered
    wg.Go(func() error {
        nodeAccount, _ := am.GetNodeAccount()
        exists, err := node.GetNodeExists(rp, nodeAccount.Address, nil)
        if err == nil {
            response.AlreadyRegistered = exists
        }
        return err
    })

    // Check node registrations are enabled
    wg.Go(func() error {
        registrationEnabled, err := settings.GetNodeRegistrationEnabled(rp, nil)
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


func registerNode(c *cli.Context, timezoneLocation string) (*api.RegisterNodeResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.RegisterNodeResponse{}

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
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

