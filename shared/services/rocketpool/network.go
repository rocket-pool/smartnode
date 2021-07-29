package rocketpool

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
    if response.RplPrice == nil { response.RplPrice = big.NewInt(0) }
    if response.MinPerMinipoolRplStake == nil { response.MinPerMinipoolRplStake = big.NewInt(0) }
    if response.MaxPerMinipoolRplStake == nil { response.MaxPerMinipoolRplStake = big.NewInt(0) }
    return response, nil
}


// Get network node fee
func (c *Client) Challenge(address common.Address) (api.SetNodeTimezoneResponse, error) {
    responseBytes, err := c.callAPI("network challenge", address.Hex())
    if err != nil {
        return api.SetNodeTimezoneResponse{}, fmt.Errorf("Could not challenge: %w", err)
    }
    var response api.SetNodeTimezoneResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.SetNodeTimezoneResponse{}, fmt.Errorf("Could not decode challenge response: %w", err)
    }
    if response.Error != "" {
        return api.SetNodeTimezoneResponse{}, fmt.Errorf("Could not challenge: %s", response.Error)
    }
    return response, nil
}


// Get network node fee
func (c *Client) Decide(address common.Address) (api.SetNodeTimezoneResponse, error) {
    responseBytes, err := c.callAPI("network decide", address.Hex())
    if err != nil {
        return api.SetNodeTimezoneResponse{}, fmt.Errorf("Could not decide: %w", err)
    }
    var response api.SetNodeTimezoneResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.SetNodeTimezoneResponse{}, fmt.Errorf("Could not decode decide response: %w", err)
    }
    if response.Error != "" {
        return api.SetNodeTimezoneResponse{}, fmt.Errorf("Could not decide: %s", response.Error)
    }
    return response, nil
}

