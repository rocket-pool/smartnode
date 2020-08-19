package rocketpool

import (
    "encoding/json"
    "fmt"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Get minipool status
func (c *Client) MinipoolStatus() (api.MinipoolStatusResponse, error) {
    responseBytes, err := c.callAPI("minipool status")
    if err != nil {
        return api.MinipoolStatusResponse{}, fmt.Errorf("Could not get minipool status: %w", err)
    }
    var response api.MinipoolStatusResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.MinipoolStatusResponse{}, fmt.Errorf("Could not decode minipool status response: %w", err)
    }
    if response.Error != "" {
        return api.MinipoolStatusResponse{}, fmt.Errorf("Could not get minipool status: %s", response.Error)
    }
    return response, nil
}


// Check whether a minipool is eligible for a refund
func (c *Client) CanRefundMinipool(address common.Address) (api.CanRefundMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-refund %s", address.Hex()))
    if err != nil {
        return api.CanRefundMinipoolResponse{}, fmt.Errorf("Could not get can refund minipool status: %w", err)
    }
    var response api.CanRefundMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanRefundMinipoolResponse{}, fmt.Errorf("Could not decode can refund minipool response: %w", err)
    }
    if response.Error != "" {
        return api.CanRefundMinipoolResponse{}, fmt.Errorf("Could not get can refund minipool status: %s", response.Error)
    }
    return response, nil
}


// Refund ETH from a minipool
func (c *Client) RefundMinipool(address common.Address) (api.RefundMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("minipool refund %s", address.Hex()))
    if err != nil {
        return api.RefundMinipoolResponse{}, fmt.Errorf("Could not refund minipool: %w", err)
    }
    var response api.RefundMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.RefundMinipoolResponse{}, fmt.Errorf("Could not decode refund minipool response: %w", err)
    }
    if response.Error != "" {
        return api.RefundMinipoolResponse{}, fmt.Errorf("Could not refund minipool: %s", response.Error)
    }
    return response, nil
}


// Check whether a minipool can be dissolved
func (c *Client) CanDissolveMinipool(address common.Address) (api.CanDissolveMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-dissolve %s", address.Hex()))
    if err != nil {
        return api.CanDissolveMinipoolResponse{}, fmt.Errorf("Could not get can dissolve minipool status: %w", err)
    }
    var response api.CanDissolveMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanDissolveMinipoolResponse{}, fmt.Errorf("Could not decode can dissolve minipool response: %w", err)
    }
    if response.Error != "" {
        return api.CanDissolveMinipoolResponse{}, fmt.Errorf("Could not get can dissolve minipool status: %s", response.Error)
    }
    return response, nil
}


// Dissolve a minipool
func (c *Client) DissolveMinipool(address common.Address) (api.DissolveMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("minipool dissolve %s", address.Hex()))
    if err != nil {
        return api.DissolveMinipoolResponse{}, fmt.Errorf("Could not dissolve minipool: %w", err)
    }
    var response api.DissolveMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.DissolveMinipoolResponse{}, fmt.Errorf("Could not decode dissolve minipool response: %w", err)
    }
    if response.Error != "" {
        return api.DissolveMinipoolResponse{}, fmt.Errorf("Could not dissolve minipool: %s", response.Error)
    }
    return response, nil
}


// Check whether a minipool can be exited
func (c *Client) CanExitMinipool(address common.Address) (api.CanExitMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-exit %s", address.Hex()))
    if err != nil {
        return api.CanExitMinipoolResponse{}, fmt.Errorf("Could not get can exit minipool status: %w", err)
    }
    var response api.CanExitMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanExitMinipoolResponse{}, fmt.Errorf("Could not decode can exit minipool response: %w", err)
    }
    if response.Error != "" {
        return api.CanExitMinipoolResponse{}, fmt.Errorf("Could not get can exit minipool status: %s", response.Error)
    }
    return response, nil
}


// Exit a minipool
func (c *Client) ExitMinipool(address common.Address) (api.ExitMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("minipool exit %s", address.Hex()))
    if err != nil {
        return api.ExitMinipoolResponse{}, fmt.Errorf("Could not exit minipool: %w", err)
    }
    var response api.ExitMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.ExitMinipoolResponse{}, fmt.Errorf("Could not decode exit minipool response: %w", err)
    }
    if response.Error != "" {
        return api.ExitMinipoolResponse{}, fmt.Errorf("Could not exit minipool: %s", response.Error)
    }
    return response, nil
}


// Check whether a minipool can be withdrawn
func (c *Client) CanWithdrawMinipool(address common.Address) (api.CanWithdrawMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-withdraw %s", address.Hex()))
    if err != nil {
        return api.CanWithdrawMinipoolResponse{}, fmt.Errorf("Could not get can withdraw minipool status: %w", err)
    }
    var response api.CanWithdrawMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanWithdrawMinipoolResponse{}, fmt.Errorf("Could not decode can withdraw minipool response: %w", err)
    }
    if response.Error != "" {
        return api.CanWithdrawMinipoolResponse{}, fmt.Errorf("Could not get can withdraw minipool status: %s", response.Error)
    }
    return response, nil
}


// Withdraw a minipool
func (c *Client) WithdrawMinipool(address common.Address) (api.WithdrawMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("minipool withdraw %s", address.Hex()))
    if err != nil {
        return api.WithdrawMinipoolResponse{}, fmt.Errorf("Could not withdraw minipool: %w", err)
    }
    var response api.WithdrawMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.WithdrawMinipoolResponse{}, fmt.Errorf("Could not decode withdraw minipool response: %w", err)
    }
    if response.Error != "" {
        return api.WithdrawMinipoolResponse{}, fmt.Errorf("Could not withdraw minipool: %s", response.Error)
    }
    return response, nil
}


// Check whether a minipool can be closed
func (c *Client) CanCloseMinipool(address common.Address) (api.CanCloseMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-close %s", address.Hex()))
    if err != nil {
        return api.CanCloseMinipoolResponse{}, fmt.Errorf("Could not get can close minipool status: %w", err)
    }
    var response api.CanCloseMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanCloseMinipoolResponse{}, fmt.Errorf("Could not decode can close minipool response: %w", err)
    }
    if response.Error != "" {
        return api.CanCloseMinipoolResponse{}, fmt.Errorf("Could not get can close minipool status: %s", response.Error)
    }
    return response, nil
}


// Close a minipool
func (c *Client) CloseMinipool(address common.Address) (api.CloseMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("minipool close %s", address.Hex()))
    if err != nil {
        return api.CloseMinipoolResponse{}, fmt.Errorf("Could not close minipool: %w", err)
    }
    var response api.CloseMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CloseMinipoolResponse{}, fmt.Errorf("Could not decode close minipool response: %w", err)
    }
    if response.Error != "" {
        return api.CloseMinipoolResponse{}, fmt.Errorf("Could not close minipool: %s", response.Error)
    }
    return response, nil
}

