package minipool

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canCloseMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanCloseMinipoolResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanCloseMinipoolResponse{}

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return nil, err
    }

    // Validate minipool owner
    nodeAccount, _ := am.GetNodeAccount()
    if err := validateMinipoolOwner(mp, nodeAccount.Address); err != nil {
        return nil, err
    }

    // Check minipool status
    status, err := mp.GetStatus()
    if err != nil {
        return nil, err
    }
    response.InvalidStatus = (status != types.Dissolved)

    // Update & return response
    response.CanClose = !response.InvalidStatus
    return &response, nil

}


func closeMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CloseMinipoolResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CloseMinipoolResponse{}

    // Create minipool
    mp, err := minipool.NewMinipool(rp, minipoolAddress)
    if err != nil {
        return nil, err
    }

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Close
    txReceipt, err := mp.Close(opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

