package account

import (
    "errors"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func initAccount(c *cli.Context) (*api.InitAccountResponse, error) {

    // Get services
    if err := services.RequireNodePassword(c); err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }

    // Response
    response := api.InitAccountResponse{}

    // Check if node account already exists
    if am.NodeAccountExists() {
        return nil, errors.New("The node account already exists")
    }

    // Create node account
    nodeAccount, err := am.CreateNodeAccount()
    if err != nil {
        return nil, err
    }
    response.AccountAddress = nodeAccount.Address.Hex()

    // Return response
    return &response, nil

}

