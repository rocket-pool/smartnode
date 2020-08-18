package rocketpool

import (
    "encoding/json"
    "fmt"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Get wallet status
func (c *Client) WalletStatus() (api.WalletStatusResponse, error) {

    // Call API
    responseBytes, err := c.callAPI("wallet status")
    if err != nil {
        return api.WalletStatusResponse{}, fmt.Errorf("Could not get wallet status: %w", err)
    }

    // Unmarshal response
    var response api.WalletStatusResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.WalletStatusResponse{}, fmt.Errorf("Could not decode wallet status: %w", err)
    }

    // Return
    return response, nil

}

