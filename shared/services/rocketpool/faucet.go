package rocketpool

import (
	"encoding/json"
	"fmt"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get faucet status
func (c *Client) FaucetStatus() (api.FaucetStatusResponse, error) {
	responseBytes, err := c.callAPI("faucet status")
	if err != nil {
		return api.FaucetStatusResponse{}, fmt.Errorf("Could not get faucet status: %w", err)
	}
	var response api.FaucetStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.FaucetStatusResponse{}, fmt.Errorf("Could not decode faucet status response: %w", err)
	}
	if response.Error != "" {
		return api.FaucetStatusResponse{}, fmt.Errorf("Could not get faucet status: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can withdraw RPL from the faucet
func (c *Client) CanFaucetWithdrawRpl() (api.CanFaucetWithdrawRplResponse, error) {
	responseBytes, err := c.callAPI("faucet can-withdraw-rpl")
	if err != nil {
		return api.CanFaucetWithdrawRplResponse{}, fmt.Errorf("Could not get can withdraw RPL from faucet status: %w", err)
	}
	var response api.CanFaucetWithdrawRplResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanFaucetWithdrawRplResponse{}, fmt.Errorf("Could not decode can withdraw RPL from faucet response: %w", err)
	}
	if response.Error != "" {
		return api.CanFaucetWithdrawRplResponse{}, fmt.Errorf("Could not get can withdraw RPL from faucet status: %s", response.Error)
	}
	return response, nil
}

// Withdraw RPL from the faucet
func (c *Client) FaucetWithdrawRpl() (api.FaucetWithdrawRplResponse, error) {
	responseBytes, err := c.callAPI("faucet withdraw-rpl")
	if err != nil {
		return api.FaucetWithdrawRplResponse{}, fmt.Errorf("Could not withdraw RPL from faucet: %w", err)
	}
	var response api.FaucetWithdrawRplResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.FaucetWithdrawRplResponse{}, fmt.Errorf("Could not decode withdraw RPL from faucet response: %w", err)
	}
	if response.Error != "" {
		return api.FaucetWithdrawRplResponse{}, fmt.Errorf("Could not withdraw RPL from faucet: %s", response.Error)
	}
	return response, nil
}
