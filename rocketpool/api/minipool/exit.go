package minipool

import (
    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canExitMinipool(c *cli.Context, minipoolAddress common.Address) (*api.CanExitMinipoolResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanExitMinipoolResponse{}

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
    status, err := mp.GetStatus(nil)
    if err != nil {
        return nil, err
    }
    response.InvalidStatus = (status != types.Staking)

    // Update & return response
    response.CanExit = !response.InvalidStatus
    return &response, nil

}


func exitMinipool(c *cli.Context, minipoolAddress common.Address) (*api.ExitMinipoolResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }

    // Exit
    // TODO: implement

    // Return response
    return &api.ExitMinipoolResponse{}, nil

}

