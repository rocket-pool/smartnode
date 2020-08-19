package rocketpool

import (
    "encoding/json"
    "fmt"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Get network node fee
func (c *Client) NodeFee() (api.NodeFeeResponse, error) {
    responseBytes, err := c.callAPI("network node-fee")
    if err != nil {
        return api.NodeFeeResponse{}, fmt.Errorf("Could not get network node fee: %w", err)
    }
    var response api.NodeFeeResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeFeeResponse{}, fmt.Errorf("Could not decode network node fee response: %w", err)
    }
    if response.Error != "" {
        return api.NodeFeeResponse{}, fmt.Errorf("Could not get network node fee: %s", response.Error)
    }
    return response, nil
}

