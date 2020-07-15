package account

import (
    "fmt"
    "io/ioutil"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func exportAccount(c *cli.Context) error {

    // Get services
    pm, err := services.GetPasswordManager(c)
    if err != nil { return err }
    am, err := services.GetAccountManager(c)
    if err != nil { return err }

    // Response
    response := &types.ExportAccountResponse{}

    // Get password
    password, err := pm.GetPassword()
    if err != nil {
        return api.PrintResponse(&types.ExportAccountResponse{
            Error: "The node password is not set",
        })
    }
    response.Password = password

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil {
        return api.PrintResponse(&types.ExportAccountResponse{
            Error: "The node account does not exist",
        })
    }
    response.KeystorePath = nodeAccount.URL.Path

    // Read node account keystore file
    keystoreFile, err := ioutil.ReadFile(nodeAccount.URL.Path)
    if err != nil {
        return api.PrintResponse(&types.ExportAccountResponse{
            Error: fmt.Sprintf("Could not read the node account keystore file: %s", err),
        })
    }
    response.KeystoreFile = string(keystoreFile)

    // Print response
    return api.PrintResponse(response)

}

