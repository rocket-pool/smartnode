package rocketpool

import (
	"encoding/json"
	"fmt"
	"math/big"

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

// Get network RPL price
func (c *Client) RplPrice() (api.RplPriceResponse, error) {
	responseBytes, err := c.callAPI("network rpl-price")
	if err != nil {
		return api.RplPriceResponse{}, fmt.Errorf("Could not get network RPL price: %w", err)
	}
	var response api.RplPriceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.RplPriceResponse{}, fmt.Errorf("Could not decode network RPL price response: %w", err)
	}
	if response.Error != "" {
		return api.RplPriceResponse{}, fmt.Errorf("Could not get network RPL price: %s", response.Error)
	}
	if response.RplPrice == nil {
		response.RplPrice = big.NewInt(0)
	}
	if response.MinPerMinipoolRplStake == nil {
		response.MinPerMinipoolRplStake = big.NewInt(0)
	}
	if response.MaxPerMinipoolRplStake == nil {
		response.MaxPerMinipoolRplStake = big.NewInt(0)
	}
	return response, nil
}

// Get network stats
func (c *Client) NetworkStats() (api.NetworkStatsResponse, error) {
	responseBytes, err := c.callAPI("network stats")
	if err != nil {
		return api.NetworkStatsResponse{}, fmt.Errorf("Could not get network stats: %w", err)
	}
	var response api.NetworkStatsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkStatsResponse{}, fmt.Errorf("Could not decode network stats response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkStatsResponse{}, fmt.Errorf("Could not get network stats: %s", response.Error)
	}
	return response, nil
}

// Get the timezone map
func (c *Client) TimezoneMap() (api.NetworkTimezonesResponse, error) {
	responseBytes, err := c.callAPI("network timezone-map")
	if err != nil {
		return api.NetworkTimezonesResponse{}, fmt.Errorf("Could not get network timezone map: %w", err)
	}
	var response api.NetworkTimezonesResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkTimezonesResponse{}, fmt.Errorf("Could not decode network timezone map response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkTimezonesResponse{}, fmt.Errorf("Could not get network timezone map: %s", response.Error)
	}
	return response, nil
}
