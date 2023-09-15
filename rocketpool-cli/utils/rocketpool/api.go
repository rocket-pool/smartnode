package rocketpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Wait for a transaction
func (c *Client) WaitForTransaction(txHash common.Hash) (api.ApiResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("wait %s", txHash.String()))
	if err != nil {
		return api.ApiResponse{}, fmt.Errorf("Error waiting for tx: %w", err)
	}
	var response api.ApiResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ApiResponse{}, fmt.Errorf("Error decoding wait response: %w", err)
	}
	if response.Error != "" {
		return api.ApiResponse{}, fmt.Errorf("Error waiting for tx: %s", response.Error)
	}
	return response, nil
}
