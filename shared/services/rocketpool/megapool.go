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
func (c *Client) CanClaimMegapoolRefund() (api.CanClaimRefundResponse, error) {
	responseBytes, err := c.callAPI("megapool can-claim-refund")
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

// Repay megapool debt
func (c *Client) ClaimMegapoolRefund() (api.ClaimRefundResponse, error) {
	responseBytes, err := c.callAPI("megapool claim-refund")
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

// Check whether the node can reduce the megapool bond
func (c *Client) CanReduceBond(amountWei *big.Int) (api.CanReduceBondResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-reduce-bond %s", amountWei.String()))
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
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool reduce-bond %s", amountWei.String()))
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
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-stake %d", validatorId))
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
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool stake %d", validatorId))
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

// Check whether the megapool validator can be disoolved
func (c *Client) CanDissolveValidator(validatorId uint64) (api.CanDissolveValidatorResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-dissolve-validator %d", validatorId))
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
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool dissolve-validator %d", validatorId))
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

// Check whether the megapool validator can be exited
func (c *Client) CanExitValidator(validatorId uint64) (api.CanExitValidatorResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-exit-validator %d", validatorId))
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
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool exit-validator %d", validatorId))
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

// Check whether we can notify a validator exit
func (c *Client) CanNotifyValidatorExit(validatorId uint64) (api.CanNotifyValidatorExitResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-notify-validator-exit %d", validatorId))
	if err != nil {
		return api.CanNotifyValidatorExitResponse{}, fmt.Errorf("Could not get can notify validator exit status: %w", err)
	}
	var response api.CanNotifyValidatorExitResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNotifyValidatorExitResponse{}, fmt.Errorf("Could not decode can notify-validator-exit response: %w", err)
	}
	if response.Error != "" {
		return api.CanNotifyValidatorExitResponse{}, fmt.Errorf("Could not get can notify validator exit status: %s", response.Error)
	}
	return response, nil
}

// Notify exit of a megapool validator
func (c *Client) NotifyValidatorExit(validatorId uint64) (api.NotifyValidatorExitResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool notify-validator-exit %d", validatorId))
	if err != nil {
		return api.NotifyValidatorExitResponse{}, fmt.Errorf("Could not notify validator exit: %w", err)
	}
	var response api.NotifyValidatorExitResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NotifyValidatorExitResponse{}, fmt.Errorf("Could not decode notify-validator-exit response: %w", err)
	}
	if response.Error != "" {
		return api.NotifyValidatorExitResponse{}, fmt.Errorf("Could not get notify-validator-exit status: %s", response.Error)
	}
	return response, nil
}

// Check whether we can notify a validator's final balance
func (c *Client) CanNotifyFinalBalance(validatorId uint64, slot uint64) (api.CanNotifyFinalBalanceResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-notify-final-balance %d %d", validatorId, slot))
	if err != nil {
		return api.CanNotifyFinalBalanceResponse{}, fmt.Errorf("Could not get can notify validator final balance status: %w", err)
	}
	var response api.CanNotifyFinalBalanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNotifyFinalBalanceResponse{}, fmt.Errorf("Could not decode can notify-final-balance response: %w", err)
	}
	if response.Error != "" {
		return api.CanNotifyFinalBalanceResponse{}, fmt.Errorf("Could not get can notify validator final balance status: %s", response.Error)
	}
	return response, nil
}

// Notify final balance of a megapool validator
func (c *Client) NotifyFinalBalance(validatorId uint64, slot uint64) (api.NotifyFinalBalanceResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool notify-final-balance %d", validatorId, slot))
	if err != nil {
		return api.NotifyFinalBalanceResponse{}, fmt.Errorf("Could not notify final balance: %w", err)
	}
	var response api.NotifyFinalBalanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NotifyFinalBalanceResponse{}, fmt.Errorf("Could not decode notify-final-balance response: %w", err)
	}
	if response.Error != "" {
		return api.NotifyFinalBalanceResponse{}, fmt.Errorf("Could not get notify-final-balance status: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can exit the megapool queue
func (c *Client) CanExitQueue(validatorIndex uint32) (api.CanExitQueueResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-exit-queue %d", validatorIndex))
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
func (c *Client) ExitQueue(validatorIndex uint32) (api.ExitQueueResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool exit-queue %d", validatorIndex))
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
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool can-delegate-upgrade %s", address.Hex()))
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
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool delegate-upgrade %s", address.Hex()))
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

// Calculate the megapool pending rewards
func (c *Client) CalculatePendingRewards() (api.MegapoolRewardSplitResponse, error) {
	responseBytes, err := c.callAPI("megapool pending-rewards")
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

// Calculate Rewards split given an arbitrary amount
func (c *Client) CalculateRewards(amountWei *big.Int) (api.MegapoolRewardSplitResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("megapool calculate-rewards %s", amountWei.String()))
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
	responseBytes, err := c.callAPI("megapool can-distribute-megapool")
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
	responseBytes, err := c.callAPI("megapool distribute-megapool")
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
