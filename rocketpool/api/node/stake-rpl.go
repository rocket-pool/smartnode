package node

import (
    "math/big"

    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func canNodeStakeRpl(c *cli.Context, amountWei *big.Int) (*api.CanNodeStakeRplResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.CanNodeStakeRplResponse{}

    // Update & return response
    return &response, nil

}


func nodeStakeRpl(c *cli.Context, amountWei *big.Int) (*api.NodeStakeRplResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeStakeRplResponse{}

    // Return response
    return &response, nil

}

