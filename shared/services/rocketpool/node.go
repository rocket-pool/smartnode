package rocketpool

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
	utils "github.com/rocket-pool/smartnode/shared/utils/api"
)

// Get node status
func (c *Client) NodeStatus() (api.NodeStatusResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/status", nil)
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
	utils.ZeroIfNil(&response.TotalRplStake)
	utils.ZeroIfNil(&response.RplStakeMegapool)
	utils.ZeroIfNil(&response.RplStakeLegacy)
	utils.ZeroIfNil(&response.RplStakeThreshold)
	utils.ZeroIfNil(&response.AccountBalances.ETH)
	utils.ZeroIfNil(&response.AccountBalances.RPL)
	utils.ZeroIfNil(&response.AccountBalances.RETH)
	utils.ZeroIfNil(&response.AccountBalances.FixedSupplyRPL)
	utils.ZeroIfNil(&response.PrimaryWithdrawalBalances.ETH)
	utils.ZeroIfNil(&response.PrimaryWithdrawalBalances.RPL)
	utils.ZeroIfNil(&response.PrimaryWithdrawalBalances.RETH)
	utils.ZeroIfNil(&response.PrimaryWithdrawalBalances.FixedSupplyRPL)
	utils.ZeroIfNil(&response.NodeRPLLocked)
	utils.ZeroIfNil(&response.RPLWithdrawalBalances.ETH)
	utils.ZeroIfNil(&response.RPLWithdrawalBalances.RPL)
	utils.ZeroIfNil(&response.RPLWithdrawalBalances.RETH)
	utils.ZeroIfNil(&response.RPLWithdrawalBalances.FixedSupplyRPL)
	utils.ZeroIfNil(&response.PendingMinimumRplStake)
	utils.ZeroIfNil(&response.PendingMaximumRplStake)
	utils.ZeroIfNil(&response.EthBorrowed)
	utils.ZeroIfNil(&response.EthBorrowedLimit)
	utils.ZeroIfNil(&response.PendingBorrowAmount)
	utils.ZeroIfNil(&response.CreditBalance)
	utils.ZeroIfNil(&response.FeeDistributorBalance)
	return response, nil
}

// Get active alerts from Alertmanager
func (c *Client) NodeAlerts() (api.NodeAlertsResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/alerts", nil)
	if err != nil {
		return api.NodeAlertsResponse{}, fmt.Errorf("could not get node alerts: %w", err)
	}
	var response api.NodeAlertsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeAlertsResponse{}, fmt.Errorf("could not decode node alerts response: %w", err)
	}
	if response.Error != "" {
		return api.NodeAlertsResponse{}, fmt.Errorf("could not get node alerts: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can be registered
func (c *Client) CanRegisterNode(timezoneLocation string) (api.CanRegisterNodeResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-register", url.Values{"timezoneLocation": {timezoneLocation}})
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
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/register", url.Values{"timezoneLocation": {timezoneLocation}})
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

// Checks if the node's primary withdrawal address can be set
func (c *Client) CanSetNodePrimaryWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (api.CanSetNodePrimaryWithdrawalAddressResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-set-primary-withdrawal-address", url.Values{
		"address": {withdrawalAddress.Hex()},
		"confirm": {strconv.FormatBool(confirm)},
	})
	if err != nil {
		return api.CanSetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not get can set node primary withdrawal address: %w", err)
	}
	var response api.CanSetNodePrimaryWithdrawalAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanSetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not decode can set node primary withdrawal address response: %w", err)
	}
	if response.Error != "" {
		return api.CanSetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not get can set node primary withdrawal address: %s", response.Error)
	}
	return response, nil
}

// Set the node's primary withdrawal address
func (c *Client) SetNodePrimaryWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (api.SetNodePrimaryWithdrawalAddressResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/set-primary-withdrawal-address", url.Values{
		"address": {withdrawalAddress.Hex()},
		"confirm": {strconv.FormatBool(confirm)},
	})
	if err != nil {
		return api.SetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not set node primary withdrawal address: %w", err)
	}
	var response api.SetNodePrimaryWithdrawalAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not decode set node primary withdrawal address response: %w", err)
	}
	if response.Error != "" {
		return api.SetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not set node primary withdrawal address: %s", response.Error)
	}
	return response, nil
}

// Checks if the node's primary withdrawal address can be confirmed
func (c *Client) CanConfirmNodePrimaryWithdrawalAddress() (api.CanSetNodePrimaryWithdrawalAddressResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-confirm-primary-withdrawal-address", nil)
	if err != nil {
		return api.CanSetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not get can confirm node primary withdrawal address: %w", err)
	}
	var response api.CanSetNodePrimaryWithdrawalAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanSetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not decode can confirm node primary withdrawal address response: %w", err)
	}
	if response.Error != "" {
		return api.CanSetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not get can confirm node primary withdrawal address: %s", response.Error)
	}
	return response, nil
}

// Confirm the node's primary withdrawal address
func (c *Client) ConfirmNodePrimaryWithdrawalAddress() (api.SetNodePrimaryWithdrawalAddressResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/confirm-primary-withdrawal-address", nil)
	if err != nil {
		return api.SetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not confirm node primary withdrawal address: %w", err)
	}
	var response api.SetNodePrimaryWithdrawalAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not decode confirm node primary withdrawal address response: %w", err)
	}
	if response.Error != "" {
		return api.SetNodePrimaryWithdrawalAddressResponse{}, fmt.Errorf("Could not confirm node primary withdrawal address: %s", response.Error)
	}
	return response, nil
}

// Checks if the node's RPL withdrawal address can be set
func (c *Client) CanSetNodeRPLWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (api.CanSetNodeRPLWithdrawalAddressResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-set-rpl-withdrawal-address", url.Values{
		"address": {withdrawalAddress.Hex()},
		"confirm": {strconv.FormatBool(confirm)},
	})
	if err != nil {
		return api.CanSetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not get can set node RPL withdrawal address: %w", err)
	}
	var response api.CanSetNodeRPLWithdrawalAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanSetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not decode can set node RPL withdrawal address response: %w", err)
	}
	if response.Error != "" {
		return api.CanSetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not get can set node RPL withdrawal address: %s", response.Error)
	}
	return response, nil
}

// Set the node's RPL withdrawal address
func (c *Client) SetNodeRPLWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (api.SetNodeRPLWithdrawalAddressResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/set-rpl-withdrawal-address", url.Values{
		"address": {withdrawalAddress.Hex()},
		"confirm": {strconv.FormatBool(confirm)},
	})
	if err != nil {
		return api.SetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not set node RPL withdrawal address: %w", err)
	}
	var response api.SetNodeRPLWithdrawalAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not decode set node RPL withdrawal address response: %w", err)
	}
	if response.Error != "" {
		return api.SetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not set node RPL withdrawal address: %s", response.Error)
	}
	return response, nil
}

// Checks if the node's RPL withdrawal address can be confirmed
func (c *Client) CanConfirmNodeRPLWithdrawalAddress() (api.CanSetNodeRPLWithdrawalAddressResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-confirm-rpl-withdrawal-address", nil)
	if err != nil {
		return api.CanSetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not get can confirm node RPL withdrawal address: %w", err)
	}
	var response api.CanSetNodeRPLWithdrawalAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanSetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not decode can confirm node RPL withdrawal address response: %w", err)
	}
	if response.Error != "" {
		return api.CanSetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not get can confirm node RPL withdrawal address: %s", response.Error)
	}
	return response, nil
}

// Confirm the node's RPL withdrawal address
func (c *Client) ConfirmNodeRPLWithdrawalAddress() (api.SetNodeRPLWithdrawalAddressResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/confirm-rpl-withdrawal-address", nil)
	if err != nil {
		return api.SetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not confirm node RPL withdrawal address: %w", err)
	}
	var response api.SetNodeRPLWithdrawalAddressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not decode confirm node RPL withdrawal address response: %w", err)
	}
	if response.Error != "" {
		return api.SetNodeRPLWithdrawalAddressResponse{}, fmt.Errorf("Could not confirm node RPL withdrawal address: %s", response.Error)
	}
	return response, nil
}

// Checks if the node's timezone location can be set
func (c *Client) CanSetNodeTimezone(timezoneLocation string) (api.CanSetNodeTimezoneResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-set-timezone", url.Values{"timezoneLocation": {timezoneLocation}})
	if err != nil {
		return api.CanSetNodeTimezoneResponse{}, fmt.Errorf("Could not get can set node timezone: %w", err)
	}
	var response api.CanSetNodeTimezoneResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanSetNodeTimezoneResponse{}, fmt.Errorf("Could not decode can set node timezone response: %w", err)
	}
	if response.Error != "" {
		return api.CanSetNodeTimezoneResponse{}, fmt.Errorf("Could not get can set node timezone: %s", response.Error)
	}
	return response, nil
}

// Set the node's timezone location
func (c *Client) SetNodeTimezone(timezoneLocation string) (api.SetNodeTimezoneResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/set-timezone", url.Values{"timezoneLocation": {timezoneLocation}})
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
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-swap-rpl", url.Values{"amountWei": {amountWei.String()}})
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

// Get the gas estimate for approving legacy RPL interaction
func (c *Client) NodeSwapRplApprovalGas(amountWei *big.Int) (api.NodeSwapRplApproveGasResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/get-swap-rpl-approval-gas", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.NodeSwapRplApproveGasResponse{}, fmt.Errorf("Could not get old RPL approval gas: %w", err)
	}
	var response api.NodeSwapRplApproveGasResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeSwapRplApproveGasResponse{}, fmt.Errorf("Could not decode node swap RPL approve gas response: %w", err)
	}
	if response.Error != "" {
		return api.NodeSwapRplApproveGasResponse{}, fmt.Errorf("Could not get old RPL approval gas: %s", response.Error)
	}
	return response, nil
}

// Approves old RPL for a token swap
func (c *Client) NodeSwapRplApprove(amountWei *big.Int) (api.NodeSwapRplApproveResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/swap-rpl-approve-rpl", url.Values{"amountWei": {amountWei.String()}})
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

// Swap node's old RPL tokens for new RPL tokens, waiting for the approval to be included in a block first
func (c *Client) NodeWaitAndSwapRpl(amountWei *big.Int, approvalTxHash common.Hash) (api.NodeSwapRplSwapResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/wait-and-swap-rpl", url.Values{
		"amountWei":      {amountWei.String()},
		"approvalTxHash": {approvalTxHash.Hex()},
	})
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

// Swap node's old RPL tokens for new RPL tokens
func (c *Client) NodeSwapRpl(amountWei *big.Int) (api.NodeSwapRplSwapResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/swap-rpl", url.Values{"amountWei": {amountWei.String()}})
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

// Get a node's legacy RPL allowance for swapping on the new RPL contract
func (c *Client) GetNodeSwapRplAllowance() (api.NodeSwapRplAllowanceResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/swap-rpl-allowance", nil)
	if err != nil {
		return api.NodeSwapRplAllowanceResponse{}, fmt.Errorf("Could not get node swap RPL allowance: %w", err)
	}
	var response api.NodeSwapRplAllowanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeSwapRplAllowanceResponse{}, fmt.Errorf("Could not decode node swap RPL allowance response: %w", err)
	}
	if response.Error != "" {
		return api.NodeSwapRplAllowanceResponse{}, fmt.Errorf("Could not get node swap RPL allowance: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can stake RPL
func (c *Client) CanNodeStakeRpl(amountWei *big.Int) (api.CanNodeStakeRplResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-stake-rpl", url.Values{"amountWei": {amountWei.String()}})
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

// Get the gas estimate for approving new RPL interaction
func (c *Client) NodeStakeRplApprovalGas(amountWei *big.Int) (api.NodeStakeRplApproveGasResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/get-stake-rpl-approval-gas", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.NodeStakeRplApproveGasResponse{}, fmt.Errorf("Could not get new RPL approval gas: %w", err)
	}
	var response api.NodeStakeRplApproveGasResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeStakeRplApproveGasResponse{}, fmt.Errorf("Could not decode node stake RPL approve gas response: %w", err)
	}
	if response.Error != "" {
		return api.NodeStakeRplApproveGasResponse{}, fmt.Errorf("Could not get new RPL approval gas: %s", response.Error)
	}
	return response, nil
}

// Approve RPL for staking against the node
func (c *Client) NodeStakeRplApprove(amountWei *big.Int) (api.NodeStakeRplApproveResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/stake-rpl-approve-rpl", url.Values{"amountWei": {amountWei.String()}})
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

// Stake RPL against the node waiting for approvalTxHash to be included in a block first
func (c *Client) NodeWaitAndStakeRpl(amountWei *big.Int, approvalTxHash common.Hash) (api.NodeStakeRplStakeResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/wait-and-stake-rpl", url.Values{
		"amountWei":      {amountWei.String()},
		"approvalTxHash": {approvalTxHash.Hex()},
	})
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

// Stake RPL against the node
func (c *Client) NodeStakeRpl(amountWei *big.Int) (api.NodeStakeRplStakeResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/stake-rpl", url.Values{"amountWei": {amountWei.String()}})
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

// Get a node's RPL allowance for the staking contract
func (c *Client) GetNodeStakeRplAllowance() (api.NodeStakeRplAllowanceResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/stake-rpl-allowance", nil)
	if err != nil {
		return api.NodeStakeRplAllowanceResponse{}, fmt.Errorf("Could not get node stake RPL allowance: %w", err)
	}
	var response api.NodeStakeRplAllowanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeStakeRplAllowanceResponse{}, fmt.Errorf("Could not decode node stake RPL allowance response: %w", err)
	}
	if response.Error != "" {
		return api.NodeStakeRplAllowanceResponse{}, fmt.Errorf("Could not get node stake RPL allowance: %s", response.Error)
	}
	return response, nil
}

// Checks if the node operator can set RPL locking allowed
func (c *Client) CanSetRPLLockingAllowed(allowed bool) (api.CanSetRplLockingAllowedResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-set-rpl-locking-allowed", url.Values{"allowed": {strconv.FormatBool(allowed)}})
	if err != nil {
		return api.CanSetRplLockingAllowedResponse{}, fmt.Errorf("Could not get can set RPL locking allowed: %w", err)
	}
	var response api.CanSetRplLockingAllowedResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanSetRplLockingAllowedResponse{}, fmt.Errorf("Could not decode can set RPL locking allowed: %w", err)
	}
	if response.Error != "" {
		return api.CanSetRplLockingAllowedResponse{}, fmt.Errorf("Could not set RPL locking allowed: %s", response.Error)
	}
	return response, nil
}

// Sets the allow state for the node to lock RPL
func (c *Client) SetRPLLockingAllowed(allowed bool) (api.SetRplLockingAllowedResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/set-rpl-locking-allowed", url.Values{"allowed": {strconv.FormatBool(allowed)}})
	if err != nil {
		return api.SetRplLockingAllowedResponse{}, fmt.Errorf("Could not set RPL locking allowed: %w", err)
	}
	var response api.SetRplLockingAllowedResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetRplLockingAllowedResponse{}, fmt.Errorf("Could not decode set RPL locking allowed response: %w", err)
	}
	if response.Error != "" {
		return api.SetRplLockingAllowedResponse{}, fmt.Errorf("Could not set RPL locking allowed: %s", response.Error)
	}
	return response, nil
}

// Checks if the node operator can set RPL stake for allowed
func (c *Client) CanSetStakeRPLForAllowed(caller common.Address, allowed bool) (api.CanSetStakeRplForAllowedResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-set-stake-rpl-for-allowed", url.Values{
		"caller":  {caller.Hex()},
		"allowed": {strconv.FormatBool(allowed)},
	})
	if err != nil {
		return api.CanSetStakeRplForAllowedResponse{}, fmt.Errorf("Could not get can set stake RPL for allowed: %w", err)
	}
	var response api.CanSetStakeRplForAllowedResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanSetStakeRplForAllowedResponse{}, fmt.Errorf("Could not decode can set stake RPL for allowed: %w", err)
	}
	if response.Error != "" {
		return api.CanSetStakeRplForAllowedResponse{}, fmt.Errorf("Could not set stake RPL for allowed: %s", response.Error)
	}
	return response, nil
}

// Sets the allow state of another address staking on behalf of the node
func (c *Client) SetStakeRPLForAllowed(caller common.Address, allowed bool) (api.SetStakeRplForAllowedResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/set-stake-rpl-for-allowed", url.Values{
		"caller":  {caller.Hex()},
		"allowed": {strconv.FormatBool(allowed)},
	})
	if err != nil {
		return api.SetStakeRplForAllowedResponse{}, fmt.Errorf("Could not set stake RPL for allowed: %w", err)
	}
	var response api.SetStakeRplForAllowedResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetStakeRplForAllowedResponse{}, fmt.Errorf("Could not decode set stake RPL for allowed response: %w", err)
	}
	if response.Error != "" {
		return api.SetStakeRplForAllowedResponse{}, fmt.Errorf("Could not set stake RPL for allowed: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can withdraw RPL
func (c *Client) CanNodeWithdrawRpl() (api.CanNodeWithdrawRplResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-withdraw-rpl", nil)
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
func (c *Client) NodeWithdrawRpl() (api.NodeWithdrawRplResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/withdraw-rpl", nil)
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

// Check whether the node can unstake legacy RPL
func (c *Client) CanNodeUnstakeLegacyRpl(amountWei *big.Int) (api.CanNodeUnstakeLegacyRplResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-unstake-legacy-rpl", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.CanNodeUnstakeLegacyRplResponse{}, fmt.Errorf("Could not get can node unstake legacy RPL status: %w", err)
	}
	var response api.CanNodeUnstakeLegacyRplResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeUnstakeLegacyRplResponse{}, fmt.Errorf("Could not decode can node unstake legacy RPL response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeUnstakeLegacyRplResponse{}, fmt.Errorf("Could not get can node unstake legacy RPL status: %s", response.Error)
	}
	return response, nil
}

// Unstake legacy RPL staked against the node
func (c *Client) NodeUnstakeLegacyRpl(amountWei *big.Int) (api.NodeUnstakeLegacyRplResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/unstake-legacy-rpl", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.NodeUnstakeLegacyRplResponse{}, fmt.Errorf("Could not unstake node legacy RPL: %w", err)
	}
	var response api.NodeUnstakeLegacyRplResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeUnstakeLegacyRplResponse{}, fmt.Errorf("Could not decode unstake node legacy RPL response: %w", err)
	}
	if response.Error != "" {
		return api.NodeUnstakeLegacyRplResponse{}, fmt.Errorf("Could not unstake node legacy RPL: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can withdraw RPL
// Used if saturn is not deployed (v1.3.1)
func (c *Client) CanNodeWithdrawRplV1_3_1(amountWei *big.Int) (api.CanNodeWithdrawRplv1_3_1Response, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-withdraw-rpl-v131", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.CanNodeWithdrawRplv1_3_1Response{}, fmt.Errorf("Could not get can node withdraw RPL status: %w", err)
	}
	var response api.CanNodeWithdrawRplv1_3_1Response
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeWithdrawRplv1_3_1Response{}, fmt.Errorf("Could not decode can node withdraw RPL response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeWithdrawRplv1_3_1Response{}, fmt.Errorf("Could not get can node withdraw RPL status: %s", response.Error)
	}
	return response, nil
}

// Withdraw RPL staked against the node
// Used if saturn is not deployed (v1.3.1)
func (c *Client) NodeWithdrawRplV1_3_1(amountWei *big.Int) (api.NodeWithdrawRplResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/withdraw-rpl-v131", url.Values{"amountWei": {amountWei.String()}})
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

// Check whether the node can unstake RPL
func (c *Client) CanNodeUnstakeRpl(amountWei *big.Int) (api.CanNodeUnstakeRplResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-unstake-rpl", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.CanNodeUnstakeRplResponse{}, fmt.Errorf("Could not get can node unstake RPL status: %w", err)
	}
	var response api.CanNodeUnstakeRplResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeUnstakeRplResponse{}, fmt.Errorf("Could not decode can node unstake RPL response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeUnstakeRplResponse{}, fmt.Errorf("Could not get can node unstake RPL status: %s", response.Error)
	}
	return response, nil
}

// Unstake RPL staked against the node
func (c *Client) NodeUnstakeRpl(amountWei *big.Int) (api.NodeUnstakeRplResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/unstake-rpl", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.NodeUnstakeRplResponse{}, fmt.Errorf("Could not unstake node RPL: %w", err)
	}
	var response api.NodeUnstakeRplResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeUnstakeRplResponse{}, fmt.Errorf("Could not decode unstake node RPL response: %w", err)
	}
	if response.Error != "" {
		return api.NodeUnstakeRplResponse{}, fmt.Errorf("Could not unstake node RPL: %s", response.Error)
	}
	return response, nil
}

// Check whether we can withdraw ETH staked on behalf of the node
func (c *Client) CanNodeWithdrawEth(amountWei *big.Int) (api.CanNodeWithdrawEthResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-withdraw-eth", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.CanNodeWithdrawEthResponse{}, fmt.Errorf("Could not get can node withdraw ETH status: %w", err)
	}
	var response api.CanNodeWithdrawEthResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeWithdrawEthResponse{}, fmt.Errorf("Could not decode can node withdraw ETH response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeWithdrawEthResponse{}, fmt.Errorf("Could not get can node withdraw ETH status: %s", response.Error)
	}
	return response, nil
}

// Withdraw ETH staked on behalf of the node
func (c *Client) NodeWithdrawEth(amountWei *big.Int) (api.NodeWithdrawEthResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/withdraw-eth", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.NodeWithdrawEthResponse{}, fmt.Errorf("Could not withdraw node ETH: %w", err)
	}
	var response api.NodeWithdrawEthResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeWithdrawEthResponse{}, fmt.Errorf("Could not decode withdraw node ETH response: %w", err)
	}
	if response.Error != "" {
		return api.NodeWithdrawEthResponse{}, fmt.Errorf("Could not withdraw node ETH: %s", response.Error)
	}
	return response, nil
}

// Check whether we can withdraw credit from the node
func (c *Client) CanNodeWithdrawCredit(amountWei *big.Int) (api.CanNodeWithdrawCreditResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-withdraw-credit", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.CanNodeWithdrawCreditResponse{}, fmt.Errorf("Could not get can node withdraw credit status: %w", err)
	}
	var response api.CanNodeWithdrawCreditResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeWithdrawCreditResponse{}, fmt.Errorf("Could not decode can node withdraw credit response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeWithdrawCreditResponse{}, fmt.Errorf("Could not get can node withdraw credit status: %s", response.Error)
	}
	return response, nil
}

// Withdraw credit from the node as rETH
func (c *Client) NodeWithdrawCredit(amountWei *big.Int) (api.NodeWithdrawCreditResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/withdraw-credit", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.NodeWithdrawCreditResponse{}, fmt.Errorf("Could not withdraw credit: %w", err)
	}
	var response api.NodeWithdrawCreditResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeWithdrawCreditResponse{}, fmt.Errorf("Could not decode withdraw credit response: %w", err)
	}
	if response.Error != "" {
		return api.NodeWithdrawCreditResponse{}, fmt.Errorf("Could not withdraw credit: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can make multiple deposits
func (c *Client) CanNodeDeposits(count uint64, amountWei *big.Int, minFee float64, salt *big.Int, expressTickets uint64) (api.CanNodeDepositsResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-deposit", url.Values{
		"count":          {strconv.FormatUint(count, 10)},
		"amountWei":      {amountWei.String()},
		"minFee":         {strconv.FormatFloat(minFee, 'f', -1, 64)},
		"salt":           {salt.String()},
		"expressTickets": {strconv.FormatUint(expressTickets, 10)},
	})
	if err != nil {
		return api.CanNodeDepositsResponse{}, fmt.Errorf("Could not get can node deposits status: %w", err)
	}
	var response api.CanNodeDepositsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeDepositsResponse{}, fmt.Errorf("Could not decode can node deposits response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeDepositsResponse{}, fmt.Errorf("Could not get can node deposits status: %s", response.Error)
	}
	return response, nil
}

// Make multiple node deposits
func (c *Client) NodeDeposits(count uint64, amountWei *big.Int, minFee float64, salt *big.Int, useCreditBalance bool, expressTickets uint64, submit bool) (api.NodeDepositsResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/deposit", url.Values{
		"count":            {strconv.FormatUint(count, 10)},
		"amountWei":        {amountWei.String()},
		"minFee":           {strconv.FormatFloat(minFee, 'f', -1, 64)},
		"salt":             {salt.String()},
		"expressTickets":   {strconv.FormatUint(expressTickets, 10)},
		"useCreditBalance": {strconv.FormatBool(useCreditBalance)},
		"submit":           {strconv.FormatBool(submit)},
	})
	if err != nil {
		return api.NodeDepositsResponse{}, fmt.Errorf("Could not make node deposits: %w", err)
	}
	var response api.NodeDepositsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeDepositsResponse{}, fmt.Errorf("Could not decode node deposits response: %w", err)
	}
	if response.Error != "" {
		return api.NodeDepositsResponse{}, fmt.Errorf("Could not make node deposits: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can send tokens
func (c *Client) CanNodeSend(amountRaw float64, token string, toAddress common.Address) (api.CanNodeSendResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-send", url.Values{
		"amountRaw": {strconv.FormatFloat(amountRaw, 'f', 10, 64)},
		"token":     {token},
		"to":        {toAddress.Hex()},
	})
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
func (c *Client) NodeSend(amountRaw float64, token string, toAddress common.Address) (api.NodeSendResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/send", url.Values{
		"amountRaw": {strconv.FormatFloat(amountRaw, 'f', 10, 64)},
		"token":     {token},
		"to":        {toAddress.Hex()},
	})
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

// Send all tokens of the given type from the node to an address.
// Uses the exact on-chain *big.Int balance to avoid float64 rounding errors.
func (c *Client) NodeSendAll(token string, toAddress common.Address) (api.NodeSendResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/send-all", url.Values{
		"token": {token},
		"to":    {toAddress.Hex()},
	})
	if err != nil {
		return api.NodeSendResponse{}, fmt.Errorf("Could not send tokens from node: %w", err)
	}
	var response api.NodeSendResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeSendResponse{}, fmt.Errorf("Could not decode node send-all response: %w", err)
	}
	if response.Error != "" {
		return api.NodeSendResponse{}, fmt.Errorf("Could not send tokens from node: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can burn tokens
func (c *Client) CanNodeBurn(amountWei *big.Int, token string) (api.CanNodeBurnResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-burn", url.Values{
		"amountWei": {amountWei.String()},
		"token":     {token},
	})
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
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/burn", url.Values{
		"amountWei": {amountWei.String()},
		"token":     {token},
	})
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

// Get node sync progress
func (c *Client) NodeSync() (api.NodeSyncProgressResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/sync", nil)
	if err != nil {
		return api.NodeSyncProgressResponse{}, fmt.Errorf("Could not get node sync: %w", err)
	}
	var response api.NodeSyncProgressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeSyncProgressResponse{}, fmt.Errorf("Could not decode node sync response: %w", err)
	}
	if response.Error != "" {
		return api.NodeSyncProgressResponse{}, fmt.Errorf("Could not get node sync: %s", response.Error)
	}
	return response, nil
}

// Check whether the node has RPL rewards available to claim
func (c *Client) CanNodeClaimRpl() (api.CanNodeClaimRplResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-claim-rpl-rewards", nil)
	if err != nil {
		return api.CanNodeClaimRplResponse{}, fmt.Errorf("Could not get can node claim rpl rewards status: %w", err)
	}
	var response api.CanNodeClaimRplResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeClaimRplResponse{}, fmt.Errorf("Could not decode can node claim rpl rewards response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeClaimRplResponse{}, fmt.Errorf("Could not get can node claim rpl rewards status: %s", response.Error)
	}
	return response, nil
}

// Claim available RPL rewards
func (c *Client) NodeClaimRpl() (api.NodeClaimRplResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/claim-rpl-rewards", nil)
	if err != nil {
		return api.NodeClaimRplResponse{}, fmt.Errorf("Could not claim rpl rewards: %w", err)
	}
	var response api.NodeClaimRplResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeClaimRplResponse{}, fmt.Errorf("Could not decode node claim rpl rewards response: %w", err)
	}
	if response.Error != "" {
		return api.NodeClaimRplResponse{}, fmt.Errorf("Could not claim rpl rewards: %s", response.Error)
	}
	return response, nil
}

// Get node RPL rewards status
func (c *Client) NodeRewards() (api.NodeRewardsResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/rewards", nil)
	if err != nil {
		return api.NodeRewardsResponse{}, fmt.Errorf("Could not get node rewards: %w", err)
	}
	var response api.NodeRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeRewardsResponse{}, fmt.Errorf("Could not decode node rewards response: %w", err)
	}
	if response.Error != "" {
		return api.NodeRewardsResponse{}, fmt.Errorf("Could not get node rewards: %s", response.Error)
	}
	return response, nil
}

// Get the deposit contract info for Rocket Pool and the Beacon Client
func (c *Client) DepositContractInfo() (api.DepositContractInfoResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/deposit-contract-info", nil)
	if err != nil {
		return api.DepositContractInfoResponse{}, fmt.Errorf("Could not get deposit contract info: %w", err)
	}
	var response api.DepositContractInfoResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.DepositContractInfoResponse{}, fmt.Errorf("Could not decode deposit contract info response: %w", err)
	}
	if response.Error != "" {
		return api.DepositContractInfoResponse{}, fmt.Errorf("Could not get deposit contract info: %s", response.Error)
	}
	return response, nil
}

// Get the initialization status of the fee distributor contract
func (c *Client) IsFeeDistributorInitialized() (api.NodeIsFeeDistributorInitializedResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/is-fee-distributor-initialized", nil)
	if err != nil {
		return api.NodeIsFeeDistributorInitializedResponse{}, fmt.Errorf("Could not get fee distributor initialization status: %w", err)
	}
	var response api.NodeIsFeeDistributorInitializedResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeIsFeeDistributorInitializedResponse{}, fmt.Errorf("Could not decode fee distributor initialization status response: %w", err)
	}
	if response.Error != "" {
		return api.NodeIsFeeDistributorInitializedResponse{}, fmt.Errorf("Could not get fee distributor initialization status: %s", response.Error)
	}
	return response, nil
}

// Get the gas cost for initializing the fee distributor contract
func (c *Client) GetInitializeFeeDistributorGas() (api.NodeInitializeFeeDistributorGasResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/get-initialize-fee-distributor-gas", nil)
	if err != nil {
		return api.NodeInitializeFeeDistributorGasResponse{}, fmt.Errorf("Could not get initialize fee distributor gas: %w", err)
	}
	var response api.NodeInitializeFeeDistributorGasResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeInitializeFeeDistributorGasResponse{}, fmt.Errorf("Could not decode initialize fee distributor gas response: %w", err)
	}
	if response.Error != "" {
		return api.NodeInitializeFeeDistributorGasResponse{}, fmt.Errorf("Could not get initialize fee distributor gas: %s", response.Error)
	}
	return response, nil
}

// Initialize the fee distributor contract
func (c *Client) InitializeFeeDistributor() (api.NodeInitializeFeeDistributorResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/initialize-fee-distributor", nil)
	if err != nil {
		return api.NodeInitializeFeeDistributorResponse{}, fmt.Errorf("Could not initialize fee distributor: %w", err)
	}
	var response api.NodeInitializeFeeDistributorResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeInitializeFeeDistributorResponse{}, fmt.Errorf("Could not decode initialize fee distributor response: %w", err)
	}
	if response.Error != "" {
		return api.NodeInitializeFeeDistributorResponse{}, fmt.Errorf("Could not initialize fee distributor: %s", response.Error)
	}
	return response, nil
}

// Check if distributing ETH from the node's fee distributor is possible
func (c *Client) CanDistribute() (api.NodeCanDistributeResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-distribute", nil)
	if err != nil {
		return api.NodeCanDistributeResponse{}, fmt.Errorf("Could not get can distribute: %w", err)
	}
	var response api.NodeCanDistributeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeCanDistributeResponse{}, fmt.Errorf("Could not decode can distribute response: %w", err)
	}
	if response.Error != "" {
		return api.NodeCanDistributeResponse{}, fmt.Errorf("Could not get can distribute: %s", response.Error)
	}
	return response, nil
}

// Distribute ETH from the node's fee distributor
func (c *Client) Distribute() (api.NodeDistributeResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/distribute", nil)
	if err != nil {
		return api.NodeDistributeResponse{}, fmt.Errorf("Could not distribute ETH: %w", err)
	}
	var response api.NodeDistributeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeDistributeResponse{}, fmt.Errorf("Could not decode distribute response: %w", err)
	}
	if response.Error != "" {
		return api.NodeDistributeResponse{}, fmt.Errorf("Could not distribute ETH: %s", response.Error)
	}
	return response, nil
}

// Get info about your eligible rewards periods, including balances and Merkle proofs
func (c *Client) GetRewardsInfo() (api.NodeGetRewardsInfoResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/get-rewards-info", nil)
	if err != nil {
		return api.NodeGetRewardsInfoResponse{}, fmt.Errorf("Could not get rewards info: %w", err)
	}
	var response api.NodeGetRewardsInfoResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeGetRewardsInfoResponse{}, fmt.Errorf("Could not decode get rewards info response: %w", err)
	}
	if response.Error != "" {
		return api.NodeGetRewardsInfoResponse{}, fmt.Errorf("Could not get rewards info: %s", response.Error)
	}
	return response, nil
}

// Check if the rewards for the given intervals can be claimed
func (c *Client) CanNodeClaimRewards(indices []uint64) (api.CanNodeClaimRewardsResponse, error) {
	indexStrings := make([]string, len(indices))
	for i, idx := range indices {
		indexStrings[i] = strconv.FormatUint(idx, 10)
	}
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-claim-rewards", url.Values{"indices": {strings.Join(indexStrings, ",")}})
	if err != nil {
		return api.CanNodeClaimRewardsResponse{}, fmt.Errorf("Could not check if can claim rewards: %w", err)
	}
	var response api.CanNodeClaimRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeClaimRewardsResponse{}, fmt.Errorf("Could not decode can claim rewards response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeClaimRewardsResponse{}, fmt.Errorf("Could not check if can claim rewards: %s", response.Error)
	}
	return response, nil
}

// Claim rewards for the given reward intervals
func (c *Client) NodeClaimRewards(indices []uint64) (api.NodeClaimRewardsResponse, error) {
	indexStrings := make([]string, len(indices))
	for i, idx := range indices {
		indexStrings[i] = strconv.FormatUint(idx, 10)
	}
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/claim-rewards", url.Values{"indices": {strings.Join(indexStrings, ",")}})
	if err != nil {
		return api.NodeClaimRewardsResponse{}, fmt.Errorf("Could not claim rewards: %w", err)
	}
	var response api.NodeClaimRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeClaimRewardsResponse{}, fmt.Errorf("Could not decode claim rewards response: %w", err)
	}
	if response.Error != "" {
		return api.NodeClaimRewardsResponse{}, fmt.Errorf("Could not claim rewards: %s", response.Error)
	}
	return response, nil
}

// Check if the rewards for the given intervals can be claimed, and RPL restaked automatically
func (c *Client) CanNodeClaimAndStakeRewards(indices []uint64, stakeAmountWei *big.Int) (api.CanNodeClaimAndStakeRewardsResponse, error) {
	indexStrings := make([]string, len(indices))
	for i, idx := range indices {
		indexStrings[i] = strconv.FormatUint(idx, 10)
	}
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-claim-and-stake-rewards", url.Values{
		"indices":     {strings.Join(indexStrings, ",")},
		"stakeAmount": {stakeAmountWei.String()},
	})
	if err != nil {
		return api.CanNodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not check if can claim and stake rewards: %w", err)
	}
	var response api.CanNodeClaimAndStakeRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not decode can claim and stake rewards response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not check if can claim and stake rewards: %s", response.Error)
	}
	return response, nil
}

// Claim rewards for the given reward intervals and restake RPL automatically
func (c *Client) NodeClaimAndStakeRewards(indices []uint64, stakeAmountWei *big.Int) (api.NodeClaimAndStakeRewardsResponse, error) {
	indexStrings := make([]string, len(indices))
	for i, idx := range indices {
		indexStrings[i] = strconv.FormatUint(idx, 10)
	}
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/claim-and-stake-rewards", url.Values{
		"indices":     {strings.Join(indexStrings, ",")},
		"stakeAmount": {stakeAmountWei.String()},
	})
	if err != nil {
		return api.NodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not claim and stake rewards: %w", err)
	}
	var response api.NodeClaimAndStakeRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not decode claim and stake rewards response: %w", err)
	}
	if response.Error != "" {
		return api.NodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not claim and stake rewards: %s", response.Error)
	}
	return response, nil
}

// Check whether or not the node is opted into the Smoothing Pool
func (c *Client) NodeGetSmoothingPoolRegistrationStatus() (api.GetSmoothingPoolRegistrationStatusResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/get-smoothing-pool-registration-status", nil)
	if err != nil {
		return api.GetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not get smoothing pool registration status: %w", err)
	}
	var response api.GetSmoothingPoolRegistrationStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not decode smoothing pool registration status response: %w", err)
	}
	if response.Error != "" {
		return api.GetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not get smoothing pool registration status: %s", response.Error)
	}
	return response, nil
}

// Check if the node's Smoothing Pool status can be changed
func (c *Client) CanNodeSetSmoothingPoolStatus(status bool) (api.CanSetSmoothingPoolRegistrationStatusResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-set-smoothing-pool-status", url.Values{"status": {strconv.FormatBool(status)}})
	if err != nil {
		return api.CanSetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not get can-set-smoothing-pool-status: %w", err)
	}
	var response api.CanSetSmoothingPoolRegistrationStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanSetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not decode can-set-smoothing-pool-status response: %w", err)
	}
	if response.Error != "" {
		return api.CanSetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not get can-set-smoothing-pool-status: %s", response.Error)
	}
	return response, nil
}

// Sets the node's Smoothing Pool opt-in status
func (c *Client) NodeSetSmoothingPoolStatus(status bool) (api.SetSmoothingPoolRegistrationStatusResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/set-smoothing-pool-status", url.Values{"status": {strconv.FormatBool(status)}})
	if err != nil {
		return api.SetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not set smoothing pool status: %w", err)
	}
	var response api.SetSmoothingPoolRegistrationStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not decode set-smoothing-pool-status response: %w", err)
	}
	if response.Error != "" {
		return api.SetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not set smoothing pool status: %s", response.Error)
	}
	return response, nil
}

func (c *Client) ResolveEnsName(name string) (api.ResolveEnsNameResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/resolve-ens-name", url.Values{"name": {name}})
	if err != nil {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not resolve ENS name: %w", err)
	}
	var response api.ResolveEnsNameResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not decode resolve-ens-name: %w", err)
	}
	if response.Error != "" {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not resolve ENS name: %s", response.Error)
	}
	return response, nil
}

func (c *Client) ReverseResolveEnsName(name string) (api.ResolveEnsNameResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/reverse-resolve-ens-name", url.Values{"address": {name}})
	if err != nil {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not reverse resolve ENS name: %w", err)
	}
	var response api.ResolveEnsNameResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not decode reverse-resolve-ens-name: %w", err)
	}
	if response.Error != "" {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not reverse resolve ENS name: %s", response.Error)
	}
	return response, nil
}

// Use the node private key to sign an arbitrary message
func (c *Client) SignMessage(message string) (api.NodeSignResponse, error) {
	// Ignore sync status so we can sign messages even without ready clients
	c.ignoreSyncCheck = true
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/sign-message", url.Values{"message": {message}})
	if err != nil {
		return api.NodeSignResponse{}, fmt.Errorf("Could not sign message: %w", err)
	}

	var response api.NodeSignResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeSignResponse{}, fmt.Errorf("Could not decode node sign response: %w", err)
	}
	if response.Error != "" {
		return api.NodeSignResponse{}, fmt.Errorf("Could not sign message: %s", response.Error)
	}
	return response, nil
}

// Check whether a vacant minipool can be created for solo staker migration
func (c *Client) CanCreateVacantMinipool(amountWei *big.Int, minFee float64, salt *big.Int, pubkey types.ValidatorPubkey) (api.CanCreateVacantMinipoolResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-create-vacant-minipool", url.Values{
		"amountWei": {amountWei.String()},
		"minFee":    {strconv.FormatFloat(minFee, 'f', -1, 64)},
		"salt":      {salt.String()},
		"pubkey":    {pubkey.Hex()},
	})
	if err != nil {
		return api.CanCreateVacantMinipoolResponse{}, fmt.Errorf("Could not get can create vacant minipool status: %w", err)
	}
	var response api.CanCreateVacantMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanCreateVacantMinipoolResponse{}, fmt.Errorf("Could not decode can create vacant minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CanCreateVacantMinipoolResponse{}, fmt.Errorf("Could not get can create vacant minipool status: %s", response.Error)
	}
	return response, nil
}

// Create a vacant minipool, which can be used to migrate a solo staker
func (c *Client) CreateVacantMinipool(amountWei *big.Int, minFee float64, salt *big.Int, pubkey types.ValidatorPubkey) (api.CreateVacantMinipoolResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/create-vacant-minipool", url.Values{
		"amountWei": {amountWei.String()},
		"minFee":    {strconv.FormatFloat(minFee, 'f', -1, 64)},
		"salt":      {salt.String()},
		"pubkey":    {pubkey.Hex()},
	})
	if err != nil {
		return api.CreateVacantMinipoolResponse{}, fmt.Errorf("Could not get create vacant minipool status: %w", err)
	}
	var response api.CreateVacantMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CreateVacantMinipoolResponse{}, fmt.Errorf("Could not decode create vacant minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CreateVacantMinipoolResponse{}, fmt.Errorf("Could not get create vacant minipool status: %s", response.Error)
	}
	return response, nil
}

// Get the node's collateral info, including pending bond reductions
func (c *Client) CheckCollateral() (api.CheckCollateralResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/check-collateral", nil)
	if err != nil {
		return api.CheckCollateralResponse{}, fmt.Errorf("Could not get check-collateral status: %w", err)
	}
	var response api.CheckCollateralResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CheckCollateralResponse{}, fmt.Errorf("Could not decode check-collateral response: %w", err)
	}
	if response.Error != "" {
		return api.CheckCollateralResponse{}, fmt.Errorf("Could not get check-collateral status: %s", response.Error)
	}
	return response, nil
}

// Get the ETH balance of the node address
func (c *Client) GetEthBalance() (api.NodeEthBalanceResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/get-eth-balance", nil)
	if err != nil {
		return api.NodeEthBalanceResponse{}, fmt.Errorf("Could not get get-eth-balance status: %w", err)
	}
	var response api.NodeEthBalanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeEthBalanceResponse{}, fmt.Errorf("Could not decode get-eth-balance response: %w", err)
	}
	if response.Error != "" {
		return api.NodeEthBalanceResponse{}, fmt.Errorf("Could not get get-eth-balance status: %s", response.Error)
	}
	return response, nil
}

// Estimates the gas for sending a zero-value message with a payload
func (c *Client) CanSendMessage(address common.Address, message []byte) (api.CanNodeSendMessageResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-send-message", url.Values{
		"address": {address.Hex()},
		"message": {hex.EncodeToString(message)},
	})
	if err != nil {
		return api.CanNodeSendMessageResponse{}, fmt.Errorf("Could not get can-send-message response: %w", err)
	}
	var response api.CanNodeSendMessageResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeSendMessageResponse{}, fmt.Errorf("Could not decode can-send-message response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeSendMessageResponse{}, fmt.Errorf("Could not get can-send-message response: %s", response.Error)
	}
	return response, nil
}

// Sends a zero-value message with a payload
func (c *Client) SendMessage(address common.Address, message []byte) (api.NodeSendMessageResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/send-message", url.Values{
		"address": {address.Hex()},
		"message": {hex.EncodeToString(message)},
	})
	if err != nil {
		return api.NodeSendMessageResponse{}, fmt.Errorf("Could not get send-message response: %w", err)
	}
	var response api.NodeSendMessageResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeSendMessageResponse{}, fmt.Errorf("Could not decode send-message response: %w", err)
	}
	if response.Error != "" {
		return api.NodeSendMessageResponse{}, fmt.Errorf("Could not get send-message response: %s", response.Error)
	}
	return response, nil
}

// Get the number of express tickets available for the node
func (c *Client) GetExpressTicketCount() (api.GetExpressTicketCountResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/get-express-ticket-count", nil)
	if err != nil {
		return api.GetExpressTicketCountResponse{}, fmt.Errorf("Could not get express ticket count: %w", err)
	}
	var response api.GetExpressTicketCountResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetExpressTicketCountResponse{}, fmt.Errorf("Could not decode express ticket count response: %w", err)
	}
	if response.Error != "" {
		return api.GetExpressTicketCountResponse{}, fmt.Errorf("Could not get express ticket count: %s", response.Error)
	}
	return response, nil
}

// Check if the node's express tickets have been provisioned
func (c *Client) GetExpressTicketsProvisioned() (api.GetExpressTicketsProvisionedResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/get-express-tickets-provisioned", nil)
	if err != nil {
		return api.GetExpressTicketsProvisionedResponse{}, fmt.Errorf("Could not get express tickets provisioned: %w", err)
	}
	var response api.GetExpressTicketsProvisionedResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetExpressTicketsProvisionedResponse{}, fmt.Errorf("Could not decode express ticket count response: %w", err)
	}
	if response.Error != "" {
		return api.GetExpressTicketsProvisionedResponse{}, fmt.Errorf("Could not get express ticket count: %s", response.Error)
	}
	return response, nil
}

func (c *Client) CanProvisionExpressTickets() (api.CanProvisionExpressTicketsResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-provision-express-tickets", nil)
	if err != nil {
		return api.CanProvisionExpressTicketsResponse{}, fmt.Errorf("Could not get can-provision-express-tickets response: %w", err)
	}
	var response api.CanProvisionExpressTicketsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProvisionExpressTicketsResponse{}, fmt.Errorf("Could not decode can-provision-express-tickets response: %w", err)
	}
	if response.Error != "" {
		return api.CanProvisionExpressTicketsResponse{}, fmt.Errorf("Could not get can-provision-express-tickets response: %s", response.Error)
	}
	return response, nil
}

func (c *Client) ProvisionExpressTickets() (api.ProvisionExpressTicketsResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/provision-express-tickets", nil)
	if err != nil {
		return api.ProvisionExpressTicketsResponse{}, fmt.Errorf("Could not get provision-express-tickets response: %w", err)
	}
	var response api.ProvisionExpressTicketsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProvisionExpressTicketsResponse{}, fmt.Errorf("Could not decode provision-express-tickets response: %w", err)
	}
	if response.Error != "" {
		return api.ProvisionExpressTicketsResponse{}, fmt.Errorf("Could not get provision-express-tickets response: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can claim unclaimed rewards
func (c *Client) CanClaimUnclaimedRewards(nodeAddress common.Address) (api.CanClaimUnclaimedRewardsResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/can-claim-unclaimed-rewards", url.Values{"nodeAddress": {nodeAddress.Hex()}})
	if err != nil {
		return api.CanClaimUnclaimedRewardsResponse{}, fmt.Errorf("Could not get can-claim-unclaimed-rewards response: %w", err)
	}
	var response api.CanClaimUnclaimedRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanClaimUnclaimedRewardsResponse{}, fmt.Errorf("Could not decode can-claim-unclaimed-rewards response: %w", err)
	}
	if response.Error != "" {
		return api.CanClaimUnclaimedRewardsResponse{}, fmt.Errorf("Could not get can-claim-unclaimed-rewards response: %s", response.Error)
	}
	return response, nil
}

// Send unclaimed rewards to a node operator's withdrawal address
func (c *Client) ClaimUnclaimedRewards(nodeAddress common.Address) (api.ClaimUnclaimedRewardsResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/node/claim-unclaimed-rewards", url.Values{"nodeAddress": {nodeAddress.Hex()}})
	if err != nil {
		return api.ClaimUnclaimedRewardsResponse{}, fmt.Errorf("Could not get claim-unclaimed-rewards response: %w", err)
	}
	var response api.ClaimUnclaimedRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ClaimUnclaimedRewardsResponse{}, fmt.Errorf("Could not decode claim-unclaimed-rewards response: %w", err)
	}
	if response.Error != "" {
		return api.ClaimUnclaimedRewardsResponse{}, fmt.Errorf("Could not get claim-unclaimed-rewards response: %s", response.Error)
	}
	return response, nil
}

// Get the bond requirement for a number of validators
func (c *Client) GetBondRequirement(numValidators uint64) (api.GetBondRequirementResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/node/get-bond-requirement", url.Values{"numValidators": {strconv.FormatUint(numValidators, 10)}})
	if err != nil {
		return api.GetBondRequirementResponse{}, fmt.Errorf("Could not get get-bond-requirement response: %w", err)
	}
	var response api.GetBondRequirementResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetBondRequirementResponse{}, fmt.Errorf("Could not decode get-bond-requirement response: %w", err)
	}
	return response, nil
}
