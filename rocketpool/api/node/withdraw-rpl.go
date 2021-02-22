package node

import (
    "math/big"

    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canNodeWithdrawRpl(c *cli.Context, amountWei *big.Int) (*api.CanNodeWithdrawRplResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanNodeWithdrawRplResponse{}

    // Update & return response
    return &response, nil

}


func nodeWithdrawRpl(c *cli.Context, amountWei *big.Int) (*api.NodeWithdrawRplResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeWithdrawRplResponse{}

    // Return response
    return &response, nil

}

