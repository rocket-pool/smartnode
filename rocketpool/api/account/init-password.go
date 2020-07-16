package account

import (
    "errors"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func runInitPassword(c *cli.Context, password string) {
    response, err := initPassword(c, password)
    if err != nil {
        api.PrintResponse(&types.InitPasswordResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func initPassword(c *cli.Context, password string) (*types.InitPasswordResponse, error) {

    // Get services
    pm, err := services.GetPasswordManager(c)
    if err != nil { return nil, err }

    // Response
    response := types.InitPasswordResponse{}

    // Check if password already exists
    if pm.PasswordExists() {
        return nil, errors.New("The node password is already set")
    }

    // Set password
    if err := pm.SetPassword(password); err != nil {
        return nil, err
    }

    // Return response
    return &response, nil

}

