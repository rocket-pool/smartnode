package rocketpool

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Check whether the supernode can make a deposit
func (c *Client) CanSupernodeDeposit(amountWei *big.Int, supernodeAddress common.Address, salt *big.Int) (api.CanNodeDepositResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("supernode can-deposit %s %s %s", amountWei.String(), supernodeAddress.Hex(), salt.String()))
	if err != nil {
		return api.CanNodeDepositResponse{}, fmt.Errorf("Could not get can supernode deposit status: %w", err)
	}
	var response api.CanNodeDepositResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeDepositResponse{}, fmt.Errorf("Could not decode can supernode deposit response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeDepositResponse{}, fmt.Errorf("Could not get can supernode deposit status: %s", response.Error)
	}
	return response, nil
}

// Make a supernode deposit
func (c *Client) SupernodeDeposit(amountWei *big.Int, supernodeAddress common.Address, salt *big.Int) (api.NodeDepositResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("supernode deposit %s %s %s", amountWei.String(), supernodeAddress.Hex(), salt.String()))
	if err != nil {
		return api.NodeDepositResponse{}, fmt.Errorf("Could not make supernode deposit: %w", err)
	}
	var response api.NodeDepositResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeDepositResponse{}, fmt.Errorf("Could not decode supernode deposit response: %w", err)
	}
	if response.Error != "" {
		return api.NodeDepositResponse{}, fmt.Errorf("Could not make supernode deposit: %s", response.Error)
	}
	return response, nil
}
