package rocketpool

import (
    "encoding/json"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Get node status
func (c *Client) NodeStatus() (api.NodeStatusResponse, error) {
    responseBytes, err := c.callAPI("node status")
    if err != nil {
        return api.NodeStatusResponse{}, fmt.Errorf("Could not get node status: %w", err)
    }
    var response api.NodeStatusResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeStatusResponse{}, fmt.Errorf("Could not decode node status response: %w", err)
    }
    if response.Error != "" {
        return api.NodeStatusResponse{}, fmt.Errorf("Could not get node status: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can be registered
func (c *Client) CanRegisterNode() (api.CanRegisterNodeResponse, error) {
    responseBytes, err := c.callAPI("node can-register")
    if err != nil {
        return api.CanRegisterNodeResponse{}, fmt.Errorf("Could not get can register node status: %w", err)
    }
    var response api.CanRegisterNodeResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanRegisterNodeResponse{}, fmt.Errorf("Could not decode can register node response: %w", err)
    }
    if response.Error != "" {
        return api.CanRegisterNodeResponse{}, fmt.Errorf("Could not get can register node status: %s", response.Error)
    }
    return response, nil
}


// Register the node
func (c *Client) RegisterNode(timezoneLocation string) (api.RegisterNodeResponse, error) {
    responseBytes, err := c.callAPIWithGasOpts(fmt.Sprintf("node register \"%s\"", timezoneLocation))
    if err != nil {
        return api.RegisterNodeResponse{}, fmt.Errorf("Could not register node: %w", err)
    }
    var response api.RegisterNodeResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.RegisterNodeResponse{}, fmt.Errorf("Could not decode register node response: %w", err)
    }
    if response.Error != "" {
        return api.RegisterNodeResponse{}, fmt.Errorf("Could not register node: %s", response.Error)
    }
    return response, nil
}


// Set the node's timezone location
func (c *Client) SetNodeTimezone(timezoneLocation string) (api.SetNodeTimezoneResponse, error) {
    responseBytes, err := c.callAPIWithGasOpts(fmt.Sprintf("node set-timezone \"%s\"", timezoneLocation))
    if err != nil {
        return api.SetNodeTimezoneResponse{}, fmt.Errorf("Could not set node timezone: %w", err)
    }
    var response api.SetNodeTimezoneResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.SetNodeTimezoneResponse{}, fmt.Errorf("Could not decode set node timezone response: %w", err)
    }
    if response.Error != "" {
        return api.SetNodeTimezoneResponse{}, fmt.Errorf("Could not set node timezone: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can make a deposit
func (c *Client) CanNodeDeposit(amountWei *big.Int) (api.CanNodeDepositResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node can-deposit %s", amountWei.String()))
    if err != nil {
        return api.CanNodeDepositResponse{}, fmt.Errorf("Could not get can node deposit status: %w", err)
    }
    var response api.CanNodeDepositResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanNodeDepositResponse{}, fmt.Errorf("Could not decode can node deposit response: %w", err)
    }
    if response.Error != "" {
        return api.CanNodeDepositResponse{}, fmt.Errorf("Could not get can node deposit status: %s", response.Error)
    }
    return response, nil
}


// Make a node deposit
func (c *Client) NodeDeposit(amountWei *big.Int, minFee float64) (api.NodeDepositResponse, error) {
    responseBytes, err := c.callAPIWithGasOpts(fmt.Sprintf("node deposit %s %f", amountWei.String(), minFee))
    if err != nil {
        return api.NodeDepositResponse{}, fmt.Errorf("Could not make node deposit: %w", err)
    }
    var response api.NodeDepositResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeDepositResponse{}, fmt.Errorf("Could not decode node deposit response: %w", err)
    }
    if response.Error != "" {
        return api.NodeDepositResponse{}, fmt.Errorf("Could not make node deposit: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can send tokens
func (c *Client) CanNodeSend(amountWei *big.Int, token string) (api.CanNodeSendResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node can-send %s %s", amountWei.String(), token))
    if err != nil {
        return api.CanNodeSendResponse{}, fmt.Errorf("Could not get can node send status: %w", err)
    }
    var response api.CanNodeSendResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanNodeSendResponse{}, fmt.Errorf("Could not decode can node send response: %w", err)
    }
    if response.Error != "" {
        return api.CanNodeSendResponse{}, fmt.Errorf("Could not get can node send status: %s", response.Error)
    }
    return response, nil
}


// Send tokens from the node to an address
func (c *Client) NodeSend(amountWei *big.Int, token string, toAddress common.Address) (api.NodeSendResponse, error) {
    responseBytes, err := c.callAPIWithGasOpts(fmt.Sprintf("node send %s %s %s", amountWei.String(), token, toAddress.Hex()))
    if err != nil {
        return api.NodeSendResponse{}, fmt.Errorf("Could not send tokens from node: %w", err)
    }
    var response api.NodeSendResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeSendResponse{}, fmt.Errorf("Could not decode node send response: %w", err)
    }
    if response.Error != "" {
        return api.NodeSendResponse{}, fmt.Errorf("Could not send tokens from node: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can burn tokens
func (c *Client) CanNodeBurn(amountWei *big.Int, token string) (api.CanNodeBurnResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node can-burn %s %s", amountWei.String(), token))
    if err != nil {
        return api.CanNodeBurnResponse{}, fmt.Errorf("Could not get can node burn status: %w", err)
    }
    var response api.CanNodeBurnResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanNodeBurnResponse{}, fmt.Errorf("Could not decode can node burn response: %w", err)
    }
    if response.Error != "" {
        return api.CanNodeBurnResponse{}, fmt.Errorf("Could not get can node burn status: %s", response.Error)
    }
    return response, nil
}


// Burn tokens owned by the node for ETH
func (c *Client) NodeBurn(amountWei *big.Int, token string) (api.NodeBurnResponse, error) {
    responseBytes, err := c.callAPIWithGasOpts(fmt.Sprintf("node burn %s %s", amountWei.String(), token))
    if err != nil {
        return api.NodeBurnResponse{}, fmt.Errorf("Could not burn tokens owned by node: %w", err)
    }
    var response api.NodeBurnResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeBurnResponse{}, fmt.Errorf("Could not decode node burn response: %w", err)
    }
    if response.Error != "" {
        return api.NodeBurnResponse{}, fmt.Errorf("Could not burn tokens owned by node: %s", response.Error)
    }
    return response, nil
}

