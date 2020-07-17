package node

import (
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func setTimezoneLocation(c *cli.Context, timezoneLocation string) (*api.SetNodeTimezoneResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.SetNodeTimezoneResponse{}

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return nil, err
    }

    // Set timezone location
    txReceipt, err := node.SetTimezoneLocation(rp, timezoneLocation, opts)
    if err != nil {
        return nil, err
    }
    response.TxHash = txReceipt.TxHash.Hex()

    // Return response
    return &response, nil

}

