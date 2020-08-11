package wallet

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func exportWallet(c *cli.Context) (*api.ExportWalletResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    pm, err := services.GetPasswordManager(c)
    if err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }

    // Response
    response := api.ExportWalletResponse{}

    // Get password
    password, _ := pm.GetPassword()
    response.Password = password

    // Serialize wallet
    wallet, err := w.String()
    if err != nil {
        return nil, err
    }
    response.Wallet = wallet

    // Return response
    return &response, nil

}

