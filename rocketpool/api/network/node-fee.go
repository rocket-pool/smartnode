package network

import (
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func getNodeFee(c *cli.Context) (*api.NodeFeeResponse, error) {

    // Get services
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeFeeResponse{}

    // Get node fee
    nodeFee, err := network.GetNodeFee(rp, nil)
    if err != nil {
        return nil, err
    }
    response.NodeFee = nodeFee

    // Return response
    return &response, nil

}

