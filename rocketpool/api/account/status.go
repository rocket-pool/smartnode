package account

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func getStatus(c *cli.Context) (*api.AccountStatusResponse, error) {

    // Get services
    pm, err := services.GetPasswordManager(c)
    if err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }

    // Response
    response := api.AccountStatusResponse{}

    // Get account status
    response.PasswordSet = pm.IsPasswordSet()
    response.AccountExists = am.NodeAccountExists()

    // Get account address
    if response.AccountExists {
        nodeAccount, _ := am.GetNodeAccount()
        response.AccountAddress = nodeAccount.Address
    }

    // Return response
    return &response, nil

}

