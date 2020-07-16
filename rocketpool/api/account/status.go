package account

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func runGetStatus(c *cli.Context) {
    response, err := getStatus(c)
    if err != nil {
        api.PrintResponse(&types.AccountStatusResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func getStatus(c *cli.Context) (*types.AccountStatusResponse, error) {

    // Get services
    pm, err := services.GetPasswordManager(c)
    if err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }

    // Response
    response := types.AccountStatusResponse{}

    // Get account status
    response.PasswordExists = pm.PasswordExists()
    response.AccountExists = am.NodeAccountExists()

    // Get account address
    if response.AccountExists {
        nodeAccount, _ := am.GetNodeAccount()
        response.AccountAddress = nodeAccount.Address.Hex()
    }

    // Return response
    return &response, nil

}

