package node

import (
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func runSetTimezoneLocation(c *cli.Context, timezoneLocation string) {
    response, err := setTimezoneLocation(c, timezoneLocation)
    if err != nil {
        api.PrintResponse(&types.SetNodeTimezoneResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func setTimezoneLocation(c *cli.Context, timezoneLocation string) (*types.SetNodeTimezoneResponse, error) {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := types.SetNodeTimezoneResponse{}

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

