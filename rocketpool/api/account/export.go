package account

import (
    "errors"
    "fmt"
    "io/ioutil"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func exportAccount(c *cli.Context) (*api.ExportAccountResponse, error) {

    // Get services
    if err := services.RequireNodeAccount(c); err != nil { return nil, err }
    pm, err := services.GetPasswordManager(c)
    if err != nil { return nil, err }
    am, err := services.GetAccountManager(c)
    if err != nil { return nil, err }

    // Response
    response := api.ExportAccountResponse{}

    // Get password
    password, _ := pm.GetPassword()
    response.Password = password

    // Get node account
    nodeAccount, _ := am.GetNodeAccount()
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

