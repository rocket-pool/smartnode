package rocketpool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get megapool status
func (c *Client) MegapoolStatus() (api.MegapoolStatusResponse, error) {
	responseBytes, err := c.callAPI("megapool status")
	if err != nil {
		return api.MegapoolStatusResponse{}, fmt.Errorf("Could not get megapool status: %w", err)
	}
	var response api.MegapoolStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolStatusResponse{}, fmt.Errorf("Could not decode megapool status response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolStatusResponse{}, fmt.Errorf("Could not get megapool status: %s", response.Error)
	}

	return response, nil
}

// Check whether the node can repay megapool debt
func (c *Client) CanRepayDebt(amountWei *big.Int) (api.CanRepayDebtResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-repay-debt %s", amountWei.String()))
	if err != nil {
		return api.CanRepayDebtResponse{}, fmt.Errorf("Could not get can repay debt status: %w", err)
	}
	var response api.CanRepayDebtResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanRepayDebtResponse{}, fmt.Errorf("Could not decode can repay debt response: %w", err)
	}
	if response.Error != "" {
		return api.CanRepayDebtResponse{}, fmt.Errorf("Could not get can repay debt status: %s", response.Error)
	}
	return response, nil
}

// Repay megapool debt
func (c *Client) RepayDebt(amountWei *big.Int) (api.RepayDebtResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool repay-debt %s", amountWei.String()))
	if err != nil {
		return api.RepayDebtResponse{}, fmt.Errorf("Could not repay megapool debt: %w", err)
	}
	var response api.RepayDebtResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.RepayDebtResponse{}, fmt.Errorf("Could not decode repay debt response: %w", err)
	}
	if response.Error != "" {
		return api.RepayDebtResponse{}, fmt.Errorf("Could not repay megapool debt: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can exit the megapool queue
func (c *Client) CanExitQueue(validatorIndex uint64, expressQueue bool) (api.CanExitQueueResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-exit-queue %d %t", validatorIndex, expressQueue))
	if err != nil {
		return api.CanExitQueueResponse{}, fmt.Errorf("Could not get can exit queue status: %w", err)
	}
	var response api.CanExitQueueResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanExitQueueResponse{}, fmt.Errorf("Could not decode can exit queue response: %w", err)
	}
	if response.Error != "" {
		return api.CanExitQueueResponse{}, fmt.Errorf("Could not get can exit queue status: %s", response.Error)
	}
	return response, nil
}

// Exit the megapool queue
func (c *Client) ExitQueue(validatorIndex uint64, expressQueue bool) (api.ExitQueueResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool exit-queue %d %t", validatorIndex, expressQueue))
	if err != nil {
		return api.ExitQueueResponse{}, fmt.Errorf("Could not exit queue: %w", err)
	}
	var response api.ExitQueueResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ExitQueueResponse{}, fmt.Errorf("Could not decode exit queue response: %w", err)
	}
	if response.Error != "" {
		return api.ExitQueueResponse{}, fmt.Errorf("Could not exit queue: %s", response.Error)
	}
	return response, nil
}

// Get the megapool's auto-upgrade setting
func (c *Client) GetUseLatestDelegate(address common.Address) (api.MegapoolGetUseLatestDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool get-use-latest-delegate %s", address.Hex()))
	if err != nil {
		return api.MegapoolGetUseLatestDelegateResponse{}, fmt.Errorf("Could not get use latest delegate for megapool: %w", err)
	}
	var response api.MegapoolGetUseLatestDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolGetUseLatestDelegateResponse{}, fmt.Errorf("Could not decode get use latest delegate for megapool response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolGetUseLatestDelegateResponse{}, fmt.Errorf("Could not get use latest delegate for megapool: %s", response.Error)
	}
	return response, nil
}

// Check whether a megapool can have its auto-upgrade setting changed
func (c *Client) CanSetUseLatestDelegateMegapool(address common.Address, setting bool) (api.MegapoolCanSetUseLatestDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-set-use-latest-delegate %s %t", address.Hex(), setting))
	if err != nil {
		return api.MegapoolCanSetUseLatestDelegateResponse{}, fmt.Errorf("Could not get can set use latest delegate for megapool status: %w", err)
	}
	var response api.MegapoolCanSetUseLatestDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolCanSetUseLatestDelegateResponse{}, fmt.Errorf("Could not decode can set use latest delegate for megapool response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolCanSetUseLatestDelegateResponse{}, fmt.Errorf("Could not get can set use latest delegate for megapool status: %s", response.Error)
	}
	return response, nil
}

// Change a megapool's auto-upgrade setting
func (c *Client) SetUseLatestDelegateMegapool(address common.Address, setting bool) (api.MegapoolSetUseLatestDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool set-use-latest-delegate %s %t", address.Hex(), setting))
	if err != nil {
		return api.MegapoolSetUseLatestDelegateResponse{}, fmt.Errorf("Could not set use latest delegate for megapool: %w", err)
	}
	var response api.MegapoolSetUseLatestDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolSetUseLatestDelegateResponse{}, fmt.Errorf("Could not decode set use latest delegate for megapool response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolSetUseLatestDelegateResponse{}, fmt.Errorf("Could not set use latest delegate for megapool: %s", response.Error)
	}
	return response, nil
}

// Get the megapool's delegate address
func (c *Client) GetDelegate(address common.Address) (api.MegapoolGetDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool get-delegate %s", address.Hex()))
	if err != nil {
		return api.MegapoolGetDelegateResponse{}, fmt.Errorf("Could get delegate for megapool: %w", err)
	}
	var response api.MegapoolGetDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolGetDelegateResponse{}, fmt.Errorf("Could not decode get delegate for megapool response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolGetDelegateResponse{}, fmt.Errorf("Could not get delegate for megapool: %s", response.Error)
	}
	return response, nil
}

// Get the megapool's effective delegate address
func (c *Client) GetEffectiveDelegate(address common.Address) (api.MegapoolGetEffectiveDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool get-effective-delegate %s", address.Hex()))
	if err != nil {
		return api.MegapoolGetEffectiveDelegateResponse{}, fmt.Errorf("Could get effective delegate for megapool: %w", err)
	}
	var response api.MegapoolGetEffectiveDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolGetEffectiveDelegateResponse{}, fmt.Errorf("Could not decode get effective delegate for megapool response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolGetEffectiveDelegateResponse{}, fmt.Errorf("Could not get effective delegate for megapool: %s", response.Error)
	}
	return response, nil
}
