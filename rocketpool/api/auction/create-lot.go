package auction

import (
    "github.com/rocket-pool/rocketpool-go/auction"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canCreateLot(c *cli.Context) (*api.CanCreateLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanCreateLotResponse{}



    // Update & return response
    response.CanCreate = !(response.InsufficientBalance || response.CreateLotDisabled)
    return &response, nil

}


func createLot(c *cli.Context) (*api.CreateLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CreateLotResponse{}



    // Return response
    return &response, nil

}

