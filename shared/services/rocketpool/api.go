package rocketpool

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Wait for a transaction
func (c *Client) WaitForTransaction(txHash common.Hash) (api.APIResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("wait %s", txHash.String()))
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
