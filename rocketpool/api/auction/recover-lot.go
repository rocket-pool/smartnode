package auction

import (
    "github.com/rocket-pool/rocketpool-go/auction"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canRecoverRplFromLot(c *cli.Context, lotIndex uint64) (*api.CanRecoverRPLFromLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanRecoverRPLFromLotResponse{}



    // Update & return response
    response.CanRecover = !(response.DoesNotExist || response.NotCleared || response.NoUnclaimedRPL || response.RPLAlreadyRecovered)
    return &response, nil

}


func recoverRplFromLot(c *cli.Context, lotIndex uint64) (*api.RecoverRPLFromLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.RecoverRPLFromLotResponse{}



    // Return response
    return &response, nil

}

