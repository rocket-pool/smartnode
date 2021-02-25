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

    _ = rp

    // Update & return response
    response.CanCreate = !(response.InsufficientBalance || response.CreateLotDisabled)
    return &response, nil

}


func createLot(c *cli.Context) (*api.CreateLotResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CreateLotResponse{}

    // Get transactor
    opts, err := w.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Create lot
    lotIndex, txReceipt, err := auction.CreateLot(rp, opts)
    if err != nil {
        return nil, err
    }
    response.LotId = lotIndex
    response.TxHash = txReceipt.TxHash

    // Return response
    return &response, nil

}

