package account

import (
    "errors"
    "fmt"
    "io/ioutil"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    types "github.com/rocket-pool/smartnode/shared/types/api"
    "github.com/rocket-pool/smartnode/shared/utils/api"
)


func runExportAccount(c *cli.Context) {
    response, err := exportAccount(c)
    if err != nil {
        api.PrintResponse(&types.ExportAccountResponse{Error: err.Error()})
    } else {
        api.PrintResponse(response)
    }
}


func exportAccount(c *cli.Context) (*types.ExportAccountResponse, error) {

    // Get services
    pm, err := services.GetPasswordManager(c)
    if err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }

    // Response
    response := types.ExportAccountResponse{}

    // Get password
    password, err := pm.GetPassword()
    if err != nil {
        return nil, errors.New("The node password is not set")
    }
    response.Password = password

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil {
        return nil, errors.New("The node account does not exist")
    }
    response.KeystorePath = nodeAccount.URL.Path

    // Read node account keystore file
    keystoreFile, err := ioutil.ReadFile(nodeAccount.URL.Path)
    if err != nil {
        return nil, fmt.Errorf("Could not read the node account keystore file: %s", err)
    }
    response.KeystoreFile = string(keystoreFile)

    // Return response
    return &response, nil

}

