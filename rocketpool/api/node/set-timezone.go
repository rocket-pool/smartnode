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

    // Get transactor
    opts, err := am.GetNodeAccountTransactor()
    if err != nil {
        return api.PrintResponse(&types.SetNodeTimezoneResponse{
            Error: err.Error(),
        })
    }

    // Set timezone location
    if _, err := node.SetTimezoneLocation(rp, timezoneLocation, opts); err != nil {
        return api.PrintResponse(&types.SetNodeTimezoneResponse{
            Error: err.Error(),
        })
    }

    // Print response
    return api.PrintResponse(&types.SetNodeTimezoneResponse{})

}

