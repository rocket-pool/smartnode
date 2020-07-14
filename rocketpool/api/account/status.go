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

    // Get account status
    passwordExists := pm.PasswordExists()
    accountExists := am.NodeAccountExists()

    // Get account address
    var accountAddress string
    if accountExists {
        nodeAccount, _ := am.GetNodeAccount()
        accountAddress = nodeAccount.Address.Hex()
    }

    // Print response
    api.PrintResponse(&types.AccountStatusResponse{
        PasswordExists: passwordExists,
        AccountExists: accountExists,
        AccountAddress: accountAddress,
    })
    return nil

}

