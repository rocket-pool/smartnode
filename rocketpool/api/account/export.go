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
    if password, err := pm.GetPassword(); err != nil {
        return api.PrintResponse(&types.ExportAccountResponse{
            Error: "The node password is not set",
        })
    } else {
        response.Password = password
    }

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil {
        return api.PrintResponse(&types.ExportAccountResponse{
            Error: "The node account does not exist",
        })
    } else {
        response.KeystorePath = nodeAccount.URL.Path
    }

    // Read node account keystore file
    if keystoreFile, err := ioutil.ReadFile(nodeAccount.URL.Path); err != nil {
        return api.PrintResponse(&types.ExportAccountResponse{
            Error: fmt.Sprintf("Could not read the node account keystore file: %s", err),
        })
    } else {
        response.KeystoreFile = string(keystoreFile)
    }

    // Print response
    return api.PrintResponse(response)

}

