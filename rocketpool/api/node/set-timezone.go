package node

import (
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func setTimezoneLocation(c *cli.Context, timezoneLocation string) error {

    // Get services
    if err := services.RequireNodeRegistered(c); err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return err }

    // Response
    response := &types.SetNodeTimezoneResponse{}

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return api.PrintResponse(&types.SetNodeTimezoneResponse{
            Error: err.Error(),
        })
    }

    // Set timezone location
    txReceipt, err := node.SetTimezoneLocation(rp, timezoneLocation, opts)
    if err != nil {
        return api.PrintResponse(&types.SetNodeTimezoneResponse{
            Error: err.Error(),
        })
    }
    response.TxHash = txReceipt.TxHash.Hex()

    // Print response
    return api.PrintResponse(response)

}

