package account

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func initPassword(c *cli.Context, password string) error {

    // Get services
    pm, err := services.GetPasswordManager(c)
    if err != nil { return err }

    // Response
    response := &types.InitPasswordResponse{}

    // Check if password already exists
    if pm.PasswordExists() {
        return api.PrintResponse(&types.InitPasswordResponse{
            Error: "The node password is already set",
        })
    }

    // Set password
    if err := pm.SetPassword(password); err != nil {
        return api.PrintResponse(&types.InitPasswordResponse{
            Error: err.Error(),
        })
    }

    // Print response
    return api.PrintResponse(response)

}

