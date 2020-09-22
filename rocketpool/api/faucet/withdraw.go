package faucet

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/types/api"
)


func faucetWithdraw(c *cli.Context, token string) (*api.FaucetWithdrawResponse, error) {

    // Get services
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }

    // Response
    response := api.FaucetWithdrawResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    // Send faucet withdrawal request
    withdrawalResponse, err := postFaucetWithdrawal(token, nodeAccount.Address)
    if err != nil {
        return nil, err
    }
    response.Error = withdrawalResponse.Error

    // Return response
    return &response, nil

}

