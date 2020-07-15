package account

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func getStatus(c *cli.Context) error {

    // Get services
    pm, err := services.GetPasswordManager(c)
    if err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }

    // Response
    response := &types.AccountStatusResponse{}

    // Get account status
    response.PasswordExists = pm.PasswordExists()
    response.AccountExists = am.NodeAccountExists()

    // Get account address
    if response.AccountExists {
        nodeAccount, _ := am.GetNodeAccount()
        response.AccountAddress = nodeAccount.Address.Hex()
    }

    // Print response
    return api.PrintResponse(response)

}

