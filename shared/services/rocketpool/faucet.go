package rocketpool

import (
    "encoding/json"
    "fmt"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Withdraw from faucet
func (c *Client) FaucetWithdraw(token string) (api.FaucetWithdrawResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("faucet withdraw %s", token))
    if err != nil {
        return api.FaucetWithdrawResponse{}, fmt.Errorf("Could not withdraw from faucet: %w", err)
    }
    var response api.FaucetWithdrawResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.FaucetWithdrawResponse{}, fmt.Errorf("Could not decode faucet withdrawal response: %w", err)
    }
    if response.Error != "" {
        return api.FaucetWithdrawResponse{}, fmt.Errorf("Could not withdraw from faucet: %s", response.Error)
    }
    return response, nil
}

