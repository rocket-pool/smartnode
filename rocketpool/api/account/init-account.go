package account

import (
    "errors"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func runInitAccount(c *cli.Context) {
    response, err := initAccount(c)
    if err != nil {
        api.PrintResponse(&types.InitAccountResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func initAccount(c *cli.Context) (*types.InitAccountResponse, error) {

    // Get services
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }

    // Response
    response := types.InitAccountResponse{}

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

