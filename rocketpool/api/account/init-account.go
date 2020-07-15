package account

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func initAccount(c *cli.Context) error {

    // Get services
    am, err := services.GetAccountManager(c)
    if err != nil { return err }

    // Response
    response := &types.InitAccountResponse{}

    // Check if node account already exists
    if am.NodeAccountExists() {
        nodeAccount, _ := am.GetNodeAccount()
        return api.PrintResponse(&types.InitAccountResponse{
            Error: "The node account already exists",
            AccountAddress: nodeAccount.Address.Hex(),
        })
    }

    // Create node account
    nodeAccount, err := am.CreateNodeAccount()
    if err != nil {
        return api.PrintResponse(&types.InitAccountResponse{
            Error: err.Error(),
        })
    }
    response.AccountAddress = nodeAccount.Address.Hex()

    // Print response
    return api.PrintResponse(response)

}

