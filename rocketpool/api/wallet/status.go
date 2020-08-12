package wallet

import (
    "github.com/rocket-pool/rocketpool-go/types"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func getStatus(c *cli.Context) (*api.WalletStatusResponse, error) {

    // Get services
    pm, err := services.GetPasswordManager(c)
    if err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }

    // Response
    response := api.WalletStatusResponse{}

    // Get wallet status
    response.PasswordSet = pm.IsPasswordSet()
    response.WalletInitialized = w.IsInitialized()

    // Get accounts if initialized
    if response.WalletInitialized {

        // Get node account
        nodeAccount, err := w.GetNodeAccount()
        if err != nil {
            return nil, err
        }
        response.AccountAddress = nodeAccount.Address

        // Get validator accounts
        validatorKeys := []types.ValidatorPubkey{}
        validatorCount, err := w.GetValidatorKeyCount()
        if err != nil {
            return nil, err
        }
        for vi := uint(0); vi < validatorCount; vi++ {
            validatorKey, err := w.GetValidatorKeyAt(vi)
            if err != nil {
                return nil, err
            }
            validatorKeys = append(validatorKeys, types.BytesToValidatorPubkey(validatorKey.PublicKey().Marshal()))
        }
        response.ValidatorKeys = validatorKeys

    }

    // Return response
    return &response, nil

}

