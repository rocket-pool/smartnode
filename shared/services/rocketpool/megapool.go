package rocketpool

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get megapool status
func (c *Client) MegapoolStatus(finalizedState bool) (api.MegapoolStatusResponse, error) {
	finalizedStr := "false"
	if finalizedState {
		finalizedStr = "true"
	}
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/status", url.Values{"finalizedState": {finalizedStr}})
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

// Get a map of the node's validators and beacon balances
func (c *Client) GetValidatorMapAndBalances() (api.MegapoolValidatorMapAndRewardsResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/validator-map-and-balances", nil)
	if err != nil {
		return api.MegapoolValidatorMapAndRewardsResponse{}, fmt.Errorf("Could not get megapool validator-map-and-balances: %w", err)
	}
	var response api.MegapoolValidatorMapAndRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolValidatorMapAndRewardsResponse{}, fmt.Errorf("Could not decode megapool validator-map-and-balances response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolValidatorMapAndRewardsResponse{}, fmt.Errorf("Could not get megapool validator-map-and-balances: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can claim a megapool refund
func (c *Client) CanClaimMegapoolRefund() (api.CanClaimRefundResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-claim-refund", nil)
	if err != nil {
		return api.CanClaimRefundResponse{}, fmt.Errorf("Could not get can claim refund status: %w", err)
	}
	var response api.CanClaimRefundResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanClaimRefundResponse{}, fmt.Errorf("Could not decode can claim refund response: %w", err)
	}
	if response.Error != "" {
		return api.CanClaimRefundResponse{}, fmt.Errorf("Could not get can claim refund status: %s", response.Error)
	}
	return response, nil
}

// Claim megapool refund
func (c *Client) ClaimMegapoolRefund() (api.ClaimRefundResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/claim-refund", nil)
	if err != nil {
		return api.ClaimRefundResponse{}, fmt.Errorf("Could not claim refund: %w", err)
	}
	var response api.ClaimRefundResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ClaimRefundResponse{}, fmt.Errorf("Could not decode claim refund response: %w", err)
	}
	if response.Error != "" {
		return api.ClaimRefundResponse{}, fmt.Errorf("Could not get claim refund status: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can repay megapool debt
func (c *Client) CanRepayDebt(amountWei *big.Int) (api.CanRepayDebtResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-repay-debt", url.Values{"amountWei": {amountWei.String()}})
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
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/repay-debt", url.Values{"amountWei": {amountWei.String()}})
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

// Check whether the node can reduce the megapool bond
func (c *Client) CanReduceBond(amountWei *big.Int) (api.CanReduceBondResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-reduce-bond", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.CanReduceBondResponse{}, fmt.Errorf("Could not get can reduce bond status: %w", err)
	}
	var response api.CanReduceBondResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanReduceBondResponse{}, fmt.Errorf("Could not decode can reduce bond response: %w", err)
	}
	if response.Error != "" {
		return api.CanReduceBondResponse{}, fmt.Errorf("Could not get can reduce bond status: %s", response.Error)
	}
	return response, nil
}

// Reduce megapool bond
func (c *Client) ReduceBond(amountWei *big.Int) (api.ReduceBondResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/reduce-bond", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.ReduceBondResponse{}, fmt.Errorf("Could not reduce bond: %w", err)
	}
	var response api.ReduceBondResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ReduceBondResponse{}, fmt.Errorf("Could not decode reduce bond response: %w", err)
	}
	if response.Error != "" {
		return api.ReduceBondResponse{}, fmt.Errorf("Could not reduce bond: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can stake a megapool validator
func (c *Client) CanStake(validatorId uint64) (api.CanStakeResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-stake", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}})
	if err != nil {
		return api.CanStakeResponse{}, fmt.Errorf("Could not get can stake status: %w", err)
	}
	var response api.CanStakeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanStakeResponse{}, fmt.Errorf("Could not decode can stake response: %w", err)
	}
	if response.Error != "" {
		return api.CanStakeResponse{}, fmt.Errorf("Could not get can stake status: %s", response.Error)
	}
	return response, nil
}

// Stake a megapool validator
func (c *Client) Stake(validatorId uint64) (api.StakeResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/stake", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}})
	if err != nil {
		return api.StakeResponse{}, fmt.Errorf("Could not stake megapool validator: %w", err)
	}
	var response api.StakeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.StakeResponse{}, fmt.Errorf("Could not decode stake response: %w", err)
	}
	if response.Error != "" {
		return api.StakeResponse{}, fmt.Errorf("Could not stake megapool validator: %s", response.Error)
	}
	return response, nil
}

// Check whether a megapool validator can be dissolved
func (c *Client) CanDissolveValidator(validatorId uint64) (api.CanDissolveValidatorResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-dissolve-validator", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}})
	if err != nil {
		return api.CanDissolveValidatorResponse{}, fmt.Errorf("Could not get can dissolve validator status: %w", err)
	}
	var response api.CanDissolveValidatorResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanDissolveValidatorResponse{}, fmt.Errorf("Could not decode can dissolve-validator response: %w", err)
	}
	if response.Error != "" {
		return api.CanDissolveValidatorResponse{}, fmt.Errorf("Could not get can dissolve status: %s", response.Error)
	}
	return response, nil
}

// Dissolve a megapool validator
func (c *Client) DissolveValidator(validatorId uint64) (api.DissolveValidatorResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/dissolve-validator", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}})
	if err != nil {
		return api.DissolveValidatorResponse{}, fmt.Errorf("Could not dissolve megapool validator: %w", err)
	}
	var response api.DissolveValidatorResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.DissolveValidatorResponse{}, fmt.Errorf("Could not decode dissolve response: %w", err)
	}
	if response.Error != "" {
		return api.DissolveValidatorResponse{}, fmt.Errorf("Could not dissolve megapool validator: %s", response.Error)
	}
	return response, nil
}

// Check whether a megapool validator can be dissolved with proof
func (c *Client) CanDissolveWithProof(validatorId uint64) (api.CanDissolveWithProofResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-dissolve-with-proof", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}})
	if err != nil {
		return api.CanDissolveWithProofResponse{}, fmt.Errorf("Could not get can dissolve-with-proof status: %w", err)
	}
	var response api.CanDissolveWithProofResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanDissolveWithProofResponse{}, fmt.Errorf("Could not decode can dissolve-with-proof response: %w", err)
	}
	if response.Error != "" {
		return api.CanDissolveWithProofResponse{}, fmt.Errorf("Could not get can dissolve-with-proof status: %s", response.Error)
	}
	return response, nil
}

// Dissolve a megapool validator with proof
func (c *Client) DissolveWithProof(validatorId uint64) (api.DissolveWithProofResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/dissolve-with-proof", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}})
	if err != nil {
		return api.DissolveWithProofResponse{}, fmt.Errorf("Could not dissolve megapool validator with proof: %w", err)
	}
	var response api.DissolveWithProofResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.DissolveWithProofResponse{}, fmt.Errorf("Could not decode dissolve-with-proof response: %w", err)
	}
	if response.Error != "" {
		return api.DissolveWithProofResponse{}, fmt.Errorf("Could not dissolve megapool validator with proof: %s", response.Error)
	}
	return response, nil
}

// Check whether a megapool validator can be exited
func (c *Client) CanExitValidator(validatorId uint64) (api.CanExitValidatorResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-exit-validator", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}})
	if err != nil {
		return api.CanExitValidatorResponse{}, fmt.Errorf("Could not get can exit validator status: %w", err)
	}
	var response api.CanExitValidatorResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanExitValidatorResponse{}, fmt.Errorf("Could not decode can exit-validator response: %w", err)
	}
	if response.Error != "" {
		return api.CanExitValidatorResponse{}, fmt.Errorf("Could not get can exit status: %s", response.Error)
	}
	return response, nil
}

// Exit a megapool validator
func (c *Client) ExitValidator(validatorId uint64) (api.ExitValidatorResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/exit-validator", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}})
	if err != nil {
		return api.ExitValidatorResponse{}, fmt.Errorf("Could not exit megapool validator: %w", err)
	}
	var response api.ExitValidatorResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ExitValidatorResponse{}, fmt.Errorf("Could not decode exit response: %w", err)
	}
	if response.Error != "" {
		return api.ExitValidatorResponse{}, fmt.Errorf("Could not exit megapool validator: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can notify validator exit
func (c *Client) CanNotifyValidatorExit(validatorId uint64) (api.CanNotifyValidatorExitResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-notify-validator-exit", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}})
	if err != nil {
		return api.CanNotifyValidatorExitResponse{}, fmt.Errorf("Could not get can notify validator exit status: %w", err)
	}
	var response api.CanNotifyValidatorExitResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNotifyValidatorExitResponse{}, fmt.Errorf("Could not decode can notify validator exit response: %w", err)
	}
	if response.Error != "" {
		return api.CanNotifyValidatorExitResponse{}, fmt.Errorf("Could not get can notify validator exit status: %s", response.Error)
	}
	return response, nil
}

// Notify the megapool that a validator has exited
func (c *Client) NotifyValidatorExit(validatorId uint64) (api.NotifyValidatorExitResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/notify-validator-exit", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}})
	if err != nil {
		return api.NotifyValidatorExitResponse{}, fmt.Errorf("Could not notify validator exit: %w", err)
	}
	var response api.NotifyValidatorExitResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NotifyValidatorExitResponse{}, fmt.Errorf("Could not decode notify validator exit response: %w", err)
	}
	if response.Error != "" {
		return api.NotifyValidatorExitResponse{}, fmt.Errorf("Could not notify validator exit: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can notify final balance
func (c *Client) CanNotifyFinalBalance(validatorId uint64, slot uint64) (api.CanNotifyFinalBalanceResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-notify-final-balance", url.Values{
		"validatorId": {fmt.Sprintf("%d", validatorId)},
		"slot":        {fmt.Sprintf("%d", slot)},
	})
	if err != nil {
		return api.CanNotifyFinalBalanceResponse{}, fmt.Errorf("Could not get can notify final balance status: %w", err)
	}
	var response api.CanNotifyFinalBalanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNotifyFinalBalanceResponse{}, fmt.Errorf("Could not decode can notify final balance response: %w", err)
	}
	if response.Error != "" {
		return api.CanNotifyFinalBalanceResponse{}, fmt.Errorf("Could not get can notify final balance status: %s", response.Error)
	}
	return response, nil
}

// Notify the megapool of a validator's final balance
func (c *Client) NotifyFinalBalance(validatorId uint64, slot uint64) (api.NotifyFinalBalanceResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/notify-final-balance", url.Values{
		"validatorId": {fmt.Sprintf("%d", validatorId)},
		"slot":        {fmt.Sprintf("%d", slot)},
	})
	if err != nil {
		return api.NotifyFinalBalanceResponse{}, fmt.Errorf("Could not notify final balance: %w", err)
	}
	var response api.NotifyFinalBalanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NotifyFinalBalanceResponse{}, fmt.Errorf("Could not decode notify final balance response: %w", err)
	}
	if response.Error != "" {
		return api.NotifyFinalBalanceResponse{}, fmt.Errorf("Could not notify final balance: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can exit the validator queue
func (c *Client) CanExitQueue(validatorIndex uint32) (api.CanExitQueueResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-exit-queue", url.Values{"validatorIndex": {fmt.Sprintf("%d", validatorIndex)}})
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

// Exit the validator queue
func (c *Client) ExitQueue(validatorIndex uint32) (api.ExitQueueResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/exit-queue", url.Values{"validatorIndex": {fmt.Sprintf("%d", validatorIndex)}})
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

// Get the gas info for a megapool delegate upgrade
func (c *Client) CanDelegateUpgradeMegapool(address common.Address) (api.MegapoolCanDelegateUpgradeResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-delegate-upgrade", url.Values{"address": {address.Hex()}})
	if err != nil {
		return api.MegapoolCanDelegateUpgradeResponse{}, fmt.Errorf("Could not get can delegate upgrade megapool status: %w", err)
	}
	var response api.MegapoolCanDelegateUpgradeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolCanDelegateUpgradeResponse{}, fmt.Errorf("Could not decode can delegate upgrade megapool response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolCanDelegateUpgradeResponse{}, fmt.Errorf("Could not get can delegate upgrade megapool status: %s", response.Error)
	}
	return response, nil
}

// Upgrade the megapool delegate
func (c *Client) DelegateUpgradeMegapool(address common.Address) (api.MegapoolDelegateUpgradeResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/delegate-upgrade", url.Values{"address": {address.Hex()}})
	if err != nil {
		return api.MegapoolDelegateUpgradeResponse{}, fmt.Errorf("Could not upgrade megapool delegate: %w", err)
	}
	var response api.MegapoolDelegateUpgradeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolDelegateUpgradeResponse{}, fmt.Errorf("Could not decode megapool delegate upgrade response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolDelegateUpgradeResponse{}, fmt.Errorf("Could not upgrade megapool delegate: %s", response.Error)
	}
	return response, nil
}

// Get the megapool's auto-upgrade setting
func (c *Client) GetUseLatestDelegate(address common.Address) (api.MegapoolGetUseLatestDelegateResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/get-use-latest-delegate", url.Values{"address": {address.Hex()}})
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
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-set-use-latest-delegate", url.Values{"address": {address.Hex()}})
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
	settingStr := "false"
	if setting {
		settingStr = "true"
	}
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/set-use-latest-delegate", url.Values{
		"address": {address.Hex()},
		"setting": {settingStr},
	})
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
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/get-delegate", url.Values{"address": {address.Hex()}})
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
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/get-effective-delegate", url.Values{"address": {address.Hex()}})
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

// Calculate the megapool pending rewards
func (c *Client) CalculatePendingRewards() (api.MegapoolRewardSplitResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/pending-rewards", nil)
	if err != nil {
		return api.MegapoolRewardSplitResponse{}, fmt.Errorf("Could not get pending rewards: %w", err)
	}
	var response api.MegapoolRewardSplitResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolRewardSplitResponse{}, fmt.Errorf("Could not decode pending rewards response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolRewardSplitResponse{}, fmt.Errorf("Could not get pending rewards: %s", response.Error)
	}
	return response, nil
}

// Calculate rewards split given an arbitrary amount
func (c *Client) CalculateRewards(amountWei *big.Int) (api.MegapoolRewardSplitResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/calculate-rewards", url.Values{"amountWei": {amountWei.String()}})
	if err != nil {
		return api.MegapoolRewardSplitResponse{}, fmt.Errorf("Could not calculate rewards: %w", err)
	}
	var response api.MegapoolRewardSplitResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MegapoolRewardSplitResponse{}, fmt.Errorf("Could not decode calculate rewards response: %w", err)
	}
	if response.Error != "" {
		return api.MegapoolRewardSplitResponse{}, fmt.Errorf("Could not get calculate rewards: %s", response.Error)
	}
	return response, nil
}

// Check if the node can distribute megapool rewards
func (c *Client) CanDistributeMegapool() (api.CanDistributeMegapoolResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/can-distribute", nil)
	if err != nil {
		return api.CanDistributeMegapoolResponse{}, fmt.Errorf("Could not get can-distribute-megapool response: %w", err)
	}
	var response api.CanDistributeMegapoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanDistributeMegapoolResponse{}, fmt.Errorf("Could not decode can-distribute-megapool response: %w", err)
	}
	if response.Error != "" {
		return api.CanDistributeMegapoolResponse{}, fmt.Errorf("Could not get can-distribute-megapool response: %s", response.Error)
	}
	return response, nil
}

// Distribute megapool rewards
func (c *Client) DistributeMegapool() (api.DistributeMegapoolResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/megapool/distribute", nil)
	if err != nil {
		return api.DistributeMegapoolResponse{}, fmt.Errorf("Could not get distribute-megapool response: %w", err)
	}
	var response api.DistributeMegapoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.DistributeMegapoolResponse{}, fmt.Errorf("Could not decode distribute-megapool response: %w", err)
	}
	if response.Error != "" {
		return api.DistributeMegapoolResponse{}, fmt.Errorf("Could not get distribute-megapool response: %s", response.Error)
	}
	return response, nil
}

// Get the bond amount required for the megapool's next validator
func (c *Client) GetNewValidatorBondRequirement() (api.GetNewValidatorBondRequirementResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/megapool/get-new-validator-bond-requirement", nil)
	if err != nil {
		return api.GetNewValidatorBondRequirementResponse{}, fmt.Errorf("Could not get new validator bond requirement: %w", err)
	}
	var response api.GetNewValidatorBondRequirementResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetNewValidatorBondRequirementResponse{}, fmt.Errorf("Could not decode new validator bond requirement response: %w", err)
	}
	if response.Error != "" {
		return api.GetNewValidatorBondRequirementResponse{}, fmt.Errorf("Could not get new validator bond requirement: %s", response.Error)
	}
	return response, nil
}

// DissolveWithProof and CanDissolveWithProof client methods added above.
// CanDissolveWithProof / DissolveWithProof (also known as DissolveWithProof) are
// already implemented above.
