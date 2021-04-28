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
    if response.RplStake == nil { response.RplStake = big.NewInt(0) }
    if response.EffectiveRplStake == nil { response.EffectiveRplStake = big.NewInt(0) }
    if response.MinimumRplStake == nil { response.MinimumRplStake = big.NewInt(0) }
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
    responseBytes, err := c.callAPI(fmt.Sprintf("node register \"%s\"", timezoneLocation))
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


// Set the node's withdrawal address
func (c *Client) SetNodeWithdrawalAddress(withdrawalAddress common.Address) (api.SetNodeWithdrawalAddressResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node set-withdrawal-address %s", withdrawalAddress.Hex()))
    if err != nil {
        return api.SetNodeWithdrawalAddressResponse{}, fmt.Errorf("Could not set node withdrawal address: %w", err)
    }
    var response api.SetNodeWithdrawalAddressResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.SetNodeWithdrawalAddressResponse{}, fmt.Errorf("Could not decode set node withdrawal address response: %w", err)
    }
    if response.Error != "" {
        return api.SetNodeWithdrawalAddressResponse{}, fmt.Errorf("Could not set node withdrawal address: %s", response.Error)
    }
    return response, nil
}


// Set the node's timezone location
func (c *Client) SetNodeTimezone(timezoneLocation string) (api.SetNodeTimezoneResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node set-timezone \"%s\"", timezoneLocation))
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


// Check whether the node can swap RPL tokens
func (c *Client) CanNodeSwapRpl(amountWei *big.Int) (api.CanNodeSwapRplResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node can-swap-rpl %s", amountWei.String()))
    if err != nil {
        return api.CanNodeSwapRplResponse{}, fmt.Errorf("Could not get can node swap RPL status: %w", err)
    }
    var response api.CanNodeSwapRplResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanNodeSwapRplResponse{}, fmt.Errorf("Could not decode can node swap RPL response: %w", err)
    }
    if response.Error != "" {
        return api.CanNodeSwapRplResponse{}, fmt.Errorf("Could not get can node swap RPL status: %s", response.Error)
    }
    return response, nil
}


// Approves old RPL for a token swap
func (c *Client) NodeSwapRplApprove(amountWei *big.Int) (api.NodeSwapRplApproveResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node swap-rpl-approve-rpl %s", amountWei.String()))
    if err != nil {
        return api.NodeSwapRplApproveResponse{}, fmt.Errorf("Could not approve old RPL: %w", err)
    }
    var response api.NodeSwapRplApproveResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeSwapRplApproveResponse{}, fmt.Errorf("Could not decode node swap RPL approve response: %w", err)
    }
    if response.Error != "" {
        return api.NodeSwapRplApproveResponse{}, fmt.Errorf("Could not approve old RPL tokens for swapping: %s", response.Error)
    }
    return response, nil
}


// Swap node's old RPL tokens for new RPL tokens
func (c *Client) NodeSwapRpl(amountWei *big.Int, approvalTxHash common.Hash) (api.NodeSwapRplSwapResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node swap-rpl %s %s", amountWei.String(), approvalTxHash.String()))
    if err != nil {
        return api.NodeSwapRplSwapResponse{}, fmt.Errorf("Could not swap node's RPL tokens: %w", err)
    }
    var response api.NodeSwapRplSwapResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeSwapRplSwapResponse{}, fmt.Errorf("Could not decode node swap RPL tokens response: %w", err)
    }
    if response.Error != "" {
        return api.NodeSwapRplSwapResponse{}, fmt.Errorf("Could not swap node's RPL tokens: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can stake RPL
func (c *Client) CanNodeStakeRpl(amountWei *big.Int) (api.CanNodeStakeRplResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node can-stake-rpl %s", amountWei.String()))
    if err != nil {
        return api.CanNodeStakeRplResponse{}, fmt.Errorf("Could not get can node stake RPL status: %w", err)
    }
    var response api.CanNodeStakeRplResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanNodeStakeRplResponse{}, fmt.Errorf("Could not decode can node stake RPL response: %w", err)
    }
    if response.Error != "" {
        return api.CanNodeStakeRplResponse{}, fmt.Errorf("Could not get can node stake RPL status: %s", response.Error)
    }
    return response, nil
}


// Approve RPL for staking against the node
func (c *Client) NodeStakeRplApprove(amountWei *big.Int) (api.NodeStakeRplApproveResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node stake-rpl-approve-rpl %s", amountWei.String()))
    if err != nil {
        return api.NodeStakeRplApproveResponse{}, fmt.Errorf("Could not approve RPL for staking: %w", err)
    }
    var response api.NodeStakeRplApproveResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeStakeRplApproveResponse{}, fmt.Errorf("Could not decode stake node RPL approve response: %w", err)
    }
    if response.Error != "" {
        return api.NodeStakeRplApproveResponse{}, fmt.Errorf("Could not approve RPL for staking: %s", response.Error)
    }
    return response, nil
}


// Stake RPL against the node
func (c *Client) NodeStakeRpl(amountWei *big.Int, approvalTxHash common.Hash) (api.NodeStakeRplStakeResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node stake-rpl %s %s", amountWei.String(), approvalTxHash.String()))
    if err != nil {
        return api.NodeStakeRplStakeResponse{}, fmt.Errorf("Could not stake node RPL: %w", err)
    }
    var response api.NodeStakeRplStakeResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeStakeRplStakeResponse{}, fmt.Errorf("Could not decode stake node RPL response: %w", err)
    }
    if response.Error != "" {
        return api.NodeStakeRplStakeResponse{}, fmt.Errorf("Could not stake node RPL: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can withdraw RPL
func (c *Client) CanNodeWithdrawRpl(amountWei *big.Int) (api.CanNodeWithdrawRplResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node can-withdraw-rpl %s", amountWei.String()))
    if err != nil {
        return api.CanNodeWithdrawRplResponse{}, fmt.Errorf("Could not get can node withdraw RPL status: %w", err)
    }
    var response api.CanNodeWithdrawRplResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanNodeWithdrawRplResponse{}, fmt.Errorf("Could not decode can node withdraw RPL response: %w", err)
    }
    if response.Error != "" {
        return api.CanNodeWithdrawRplResponse{}, fmt.Errorf("Could not get can node withdraw RPL status: %s", response.Error)
    }
    return response, nil
}


// Withdraw RPL staked against the node
func (c *Client) NodeWithdrawRpl(amountWei *big.Int) (api.NodeWithdrawRplResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node withdraw-rpl %s", amountWei.String()))
    if err != nil {
        return api.NodeWithdrawRplResponse{}, fmt.Errorf("Could not withdraw node RPL: %w", err)
    }
    var response api.NodeWithdrawRplResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeWithdrawRplResponse{}, fmt.Errorf("Could not decode withdraw node RPL response: %w", err)
    }
    if response.Error != "" {
        return api.NodeWithdrawRplResponse{}, fmt.Errorf("Could not withdraw node RPL: %s", response.Error)
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
    responseBytes, err := c.callAPI(fmt.Sprintf("node deposit %s %f", amountWei.String(), minFee))
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


// Get the minipool address for a new deposit
func (c *Client) GetMinipoolAddress(txHash common.Hash) (api.NodeDepositMinipoolResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node get-minipool-address %s", txHash.String()))
    if err != nil {
        return api.NodeDepositMinipoolResponse{}, fmt.Errorf("Could not get minipool address: %w", err)
    }
    var response api.NodeDepositMinipoolResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.NodeDepositMinipoolResponse{}, fmt.Errorf("Could not decode minipool address response: %w", err)
    }
    if response.Error != "" {
        return api.NodeDepositMinipoolResponse{}, fmt.Errorf("Could not get minipool address: %s", response.Error)
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
    responseBytes, err := c.callAPI(fmt.Sprintf("node send %s %s %s", amountWei.String(), token, toAddress.Hex()))
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
    responseBytes, err := c.callAPI(fmt.Sprintf("node burn %s %s", amountWei.String(), token))
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


// Wait for a transaction
func (c *Client) WaitForTransaction(txHash common.Hash) (api.APIResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("node wait %s", txHash.String()))
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

