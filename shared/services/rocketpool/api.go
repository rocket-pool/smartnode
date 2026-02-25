package rocketpool

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Wait for a transaction
func (c *Client) WaitForTransaction(txHash common.Hash) (api.APIResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/wait", url.Values{"txHash": {txHash.Hex()}})
	if err != nil {
		return api.APIResponse{}, fmt.Errorf("Error waiting for tx: %w", err)
	}
	var response api.APIResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.APIResponse{}, fmt.Errorf("Error decoding wait response: %w", err)
	}
	if response.Error != "" {
		return api.APIResponse{}, fmt.Errorf("Error waiting for tx: %s", response.Error)
	}
	return response, nil
}
