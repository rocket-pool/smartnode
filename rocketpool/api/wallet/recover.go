package wallet

import (
    "errors"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func recoverWallet(c *cli.Context, mnemonic string) (*api.RecoverWalletResponse, error) {

    // Get services
    if err := services.RequireNodePassword(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }

    // Response
    response := api.RecoverWalletResponse{}

    // Check if wallet is already initialized
    if w.IsInitialized() {
        return nil, errors.New("The wallet is already initialized")
    }

    // Recover wallet
    if err := w.Recover(mnemonic); err != nil {
        return nil, err
    }

    // Return response
    return &response, nil

}

