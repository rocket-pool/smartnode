package rocketpool

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// VerifyMegapoolValidatorPerformance computes RPIP-73 target-vote performance
// over [startEpoch, endEpoch] for one or more validators of a megapool. If
// megapoolAddress is the zero address the daemon resolves the node's own
// megapool. targets is either the literal "all" or a comma-separated list of
// validator IDs. This call has no client-side deadline because each epoch
// requires fetching attestation data that can take a while per epoch on an
// archival beacon node.
func (c *Client) VerifyMegapoolValidatorPerformance(megapoolAddress common.Address, targets string, startEpoch, endEpoch uint64) (api.VerifyPerformanceBatchResponse, error) {
	values := url.Values{
		"targets":    {targets},
		"startEpoch": {strconv.FormatUint(startEpoch, 10)},
		"endEpoch":   {strconv.FormatUint(endEpoch, 10)},
	}
	if (megapoolAddress != common.Address{}) {
		values.Set("megapoolAddress", megapoolAddress.Hex())
	}
	responseBytes, err := c.callHTTPAPICtx(context.Background(), "GET", "/api/megapool/verify-performance", values)
	if err != nil {
		return api.VerifyPerformanceBatchResponse{}, fmt.Errorf("Could not verify megapool validator performance: %w", err)
	}
	var response api.VerifyPerformanceBatchResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.VerifyPerformanceBatchResponse{}, fmt.Errorf("Could not decode verify-performance response: %w", err)
	}
	if response.Error != "" {
		return api.VerifyPerformanceBatchResponse{}, fmt.Errorf("Could not verify megapool validator performance: %s", response.Error)
	}
	return response, nil
}

// challengePerformanceValues builds the shared parameters of the
// can-challenge-performance and challenge-performance calls. validatorIds and
// participation are serialized as comma-separated decimal strings.
func challengePerformanceValues(megapoolAddress common.Address, validatorIds []uint32, startEpoch uint64, participation []*big.Int) url.Values {
	ids := make([]string, len(validatorIds))
	for i, id := range validatorIds {
		ids[i] = strconv.FormatUint(uint64(id), 10)
	}
	words := make([]string, len(participation))
	for i, word := range participation {
		words[i] = word.String()
	}
	values := url.Values{
		"validatorIds":  {strings.Join(ids, ",")},
		"startEpoch":    {strconv.FormatUint(startEpoch, 10)},
		"participation": {strings.Join(words, ",")},
	}
	if (megapoolAddress != common.Address{}) {
		values.Set("megapoolAddress", megapoolAddress.Hex())
	}
	return values
}

// CanChallengeMegapoolPerformance checks whether the node can challenge the
// target-vote performance of a group of megapool validators, returning the
// challenge bond, the node's RPL balance, and the gas estimate. If
// megapoolAddress is the zero address the daemon resolves the node's own
// megapool. This call has no client-side deadline because the gas estimate
// requires downloading a beacon state to build the slot proof.
func (c *Client) CanChallengeMegapoolPerformance(megapoolAddress common.Address, validatorIds []uint32, startEpoch uint64, participation []*big.Int) (api.CanChallengeMegapoolPerformanceResponse, error) {
	values := challengePerformanceValues(megapoolAddress, validatorIds, startEpoch, participation)
	responseBytes, err := c.callHTTPAPICtx(context.Background(), "GET", "/api/megapool/can-challenge-performance", values)
	if err != nil {
		return api.CanChallengeMegapoolPerformanceResponse{}, fmt.Errorf("Could not get can challenge megapool performance status: %w", err)
	}
	var response api.CanChallengeMegapoolPerformanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanChallengeMegapoolPerformanceResponse{}, fmt.Errorf("Could not decode can challenge megapool performance response: %w", err)
	}
	if response.Error != "" {
		return api.CanChallengeMegapoolPerformanceResponse{}, fmt.Errorf("Could not get can challenge megapool performance status: %s", response.Error)
	}
	return response, nil
}

// ChallengeMegapoolPerformance submits a target-vote performance challenge
// against a group of megapool validators. This call has no client-side
// deadline because the transaction requires downloading a beacon state to
// build the slot proof.
func (c *Client) ChallengeMegapoolPerformance(megapoolAddress common.Address, validatorIds []uint32, startEpoch uint64, participation []*big.Int) (api.ChallengeMegapoolPerformanceResponse, error) {
	values := challengePerformanceValues(megapoolAddress, validatorIds, startEpoch, participation)
	responseBytes, err := c.callHTTPAPICtx(context.Background(), "POST", "/api/megapool/challenge-performance", values)
	if err != nil {
		return api.ChallengeMegapoolPerformanceResponse{}, fmt.Errorf("Could not challenge megapool performance: %w", err)
	}
	var response api.ChallengeMegapoolPerformanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ChallengeMegapoolPerformanceResponse{}, fmt.Errorf("Could not decode challenge megapool performance response: %w", err)
	}
	if response.Error != "" {
		return api.ChallengeMegapoolPerformanceResponse{}, fmt.Errorf("Could not challenge megapool performance: %s", response.Error)
	}
	return response, nil
}

// Get megapool status
func (c *Client) MegapoolStatus(finalizedState bool) (api.MegapoolStatusResponse, error) {
	finalizedStr := "false"
	if finalizedState {
		finalizedStr = "true"
	}
	return c.callAPI[api.MegapoolStatusResponse]("GET", "/api/megapool/status", url.Values{"finalizedState": {finalizedStr}}, "Could not get megapool status")
}

// Get a map of the node's validators and beacon balances
func (c *Client) GetValidatorMapAndBalances() (api.MegapoolValidatorMapAndRewardsResponse, error) {
	return c.callAPI[api.MegapoolValidatorMapAndRewardsResponse]("GET", "/api/megapool/validator-map-and-balances", nil, "Could not get megapool validator-map-and-balances")
}

// Check whether the node can claim a megapool refund
func (c *Client) CanClaimMegapoolRefund() (api.CanClaimRefundResponse, error) {
	return c.callAPI[api.CanClaimRefundResponse]("GET", "/api/megapool/can-claim-refund", nil, "Could not get can claim refund status")
}

// Claim megapool refund
func (c *Client) ClaimMegapoolRefund() (api.ClaimRefundResponse, error) {
	return c.callAPI[api.ClaimRefundResponse]("POST", "/api/megapool/claim-refund", nil, "Could not claim refund")
}

// Check whether the node can repay megapool debt
func (c *Client) CanRepayDebt(amountWei *big.Int) (api.CanRepayDebtResponse, error) {
	return c.callAPI[api.CanRepayDebtResponse]("GET", "/api/megapool/can-repay-debt", url.Values{"amountWei": {amountWei.String()}}, "Could not get can repay debt status")
}

// Repay megapool debt
func (c *Client) RepayDebt(amountWei *big.Int) (api.RepayDebtResponse, error) {
	return c.callAPI[api.RepayDebtResponse]("POST", "/api/megapool/repay-debt", url.Values{"amountWei": {amountWei.String()}}, "Could not repay megapool debt")
}

// Check whether the node can reduce the megapool bond
func (c *Client) CanReduceBond(amountWei *big.Int) (api.CanReduceBondResponse, error) {
	return c.callAPI[api.CanReduceBondResponse]("GET", "/api/megapool/can-reduce-bond", url.Values{"amountWei": {amountWei.String()}}, "Could not get can reduce bond status")
}

// Reduce megapool bond
func (c *Client) ReduceBond(amountWei *big.Int) (api.ReduceBondResponse, error) {
	return c.callAPI[api.ReduceBondResponse]("POST", "/api/megapool/reduce-bond", url.Values{"amountWei": {amountWei.String()}}, "Could not reduce bond")
}

// Check whether the node can stake a megapool validator
func (c *Client) CanStake(validatorId uint64) (api.CanStakeResponse, error) {
	return c.callAPI[api.CanStakeResponse]("GET", "/api/megapool/can-stake", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}}, "Could not get can stake status")
}

// Stake a megapool validator
func (c *Client) Stake(validatorId uint64) (api.StakeResponse, error) {
	return c.callAPI[api.StakeResponse]("POST", "/api/megapool/stake", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}}, "Could not stake megapool validator")
}

// Check whether a megapool validator can be dissolved
func (c *Client) CanDissolveValidator(validatorId uint64) (api.CanDissolveValidatorResponse, error) {
	return c.callAPI[api.CanDissolveValidatorResponse]("GET", "/api/megapool/can-dissolve-validator", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}}, "Could not get can dissolve validator status")
}

// Dissolve a megapool validator
func (c *Client) DissolveValidator(validatorId uint64) (api.DissolveValidatorResponse, error) {
	return c.callAPI[api.DissolveValidatorResponse]("POST", "/api/megapool/dissolve-validator", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}}, "Could not dissolve megapool validator")
}

// Check whether a megapool validator can be dissolved with proof
func (c *Client) CanDissolveWithProof(validatorId uint64) (api.CanDissolveWithProofResponse, error) {
	return c.callAPI[api.CanDissolveWithProofResponse]("GET", "/api/megapool/can-dissolve-with-proof", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}}, "Could not get can dissolve-with-proof status")
}

// Dissolve a megapool validator with proof
func (c *Client) DissolveWithProof(validatorId uint64) (api.DissolveWithProofResponse, error) {
	return c.callAPI[api.DissolveWithProofResponse]("POST", "/api/megapool/dissolve-with-proof", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}}, "Could not dissolve megapool validator with proof")
}

// Check whether a megapool validator can be exited
func (c *Client) CanExitValidator(validatorId uint64) (api.CanExitValidatorResponse, error) {
	return c.callAPI[api.CanExitValidatorResponse]("GET", "/api/megapool/can-exit-validator", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}}, "Could not get can exit validator status")
}

// Exit a megapool validator
func (c *Client) ExitValidator(validatorId uint64) (api.ExitValidatorResponse, error) {
	return c.callAPI[api.ExitValidatorResponse]("POST", "/api/megapool/exit-validator", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}}, "Could not exit megapool validator")
}

// Check whether the node can notify validator exit
func (c *Client) CanNotifyValidatorExit(validatorId uint64) (api.CanNotifyValidatorExitResponse, error) {
	return c.callAPI[api.CanNotifyValidatorExitResponse]("GET", "/api/megapool/can-notify-validator-exit", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}}, "Could not get can notify validator exit status")
}

// Notify the megapool that a validator has exited
func (c *Client) NotifyValidatorExit(validatorId uint64) (api.NotifyValidatorExitResponse, error) {
	return c.callAPI[api.NotifyValidatorExitResponse]("POST", "/api/megapool/notify-validator-exit", url.Values{"validatorId": {fmt.Sprintf("%d", validatorId)}}, "Could not notify validator exit")
}

// Check whether the node can notify final balance
func (c *Client) CanNotifyFinalBalance(validatorId uint64, slot uint64) (api.CanNotifyFinalBalanceResponse, error) {
	return c.callAPI[api.CanNotifyFinalBalanceResponse]("GET", "/api/megapool/can-notify-final-balance", url.Values{
		"validatorId": {fmt.Sprintf("%d", validatorId)},
		"slot":        {fmt.Sprintf("%d", slot)},
	}, "Could not get can notify final balance status")
}

// Notify the megapool of a validator's final balance
func (c *Client) NotifyFinalBalance(validatorId uint64, slot uint64) (api.NotifyFinalBalanceResponse, error) {
	return c.callAPI[api.NotifyFinalBalanceResponse]("POST", "/api/megapool/notify-final-balance", url.Values{
		"validatorId": {fmt.Sprintf("%d", validatorId)},
		"slot":        {fmt.Sprintf("%d", slot)},
	}, "Could not notify final balance")
}

// Check whether the node can exit the validator queue
func (c *Client) CanExitQueue(validatorIndex uint32) (api.CanExitQueueResponse, error) {
	return c.callAPI[api.CanExitQueueResponse]("GET", "/api/megapool/can-exit-queue", url.Values{"validatorIndex": {fmt.Sprintf("%d", validatorIndex)}}, "Could not get can exit queue status")
}

// Exit the validator queue
func (c *Client) ExitQueue(validatorIndex uint32) (api.ExitQueueResponse, error) {
	return c.callAPI[api.ExitQueueResponse]("POST", "/api/megapool/exit-queue", url.Values{"validatorIndex": {fmt.Sprintf("%d", validatorIndex)}}, "Could not exit queue")
}

// Get the gas info for a megapool delegate upgrade
func (c *Client) CanDelegateUpgradeMegapool(address common.Address) (api.MegapoolCanDelegateUpgradeResponse, error) {
	return c.callAPI[api.MegapoolCanDelegateUpgradeResponse]("GET", "/api/megapool/can-delegate-upgrade", url.Values{"address": {address.Hex()}}, "Could not get can delegate upgrade megapool status")
}

// Upgrade the megapool delegate
func (c *Client) DelegateUpgradeMegapool(address common.Address) (api.MegapoolDelegateUpgradeResponse, error) {
	return c.callAPI[api.MegapoolDelegateUpgradeResponse]("POST", "/api/megapool/delegate-upgrade", url.Values{"address": {address.Hex()}}, "Could not upgrade megapool delegate")
}

// Get the megapool's use-latest-delegate setting
func (c *Client) GetUseLatestDelegate(address common.Address) (api.MegapoolGetUseLatestDelegateResponse, error) {
	return c.callAPI[api.MegapoolGetUseLatestDelegateResponse]("GET", "/api/megapool/get-use-latest-delegate", url.Values{"address": {address.Hex()}}, "Could not get use latest delegate for megapool")
}

// Check whether a megapool can have its use-latest-delegate setting changed
func (c *Client) CanSetUseLatestDelegateMegapool(address common.Address, useLatest bool) (api.MegapoolCanSetUseLatestDelegateResponse, error) {
	return c.callAPI[api.MegapoolCanSetUseLatestDelegateResponse]("GET", "/api/megapool/can-set-use-latest-delegate", url.Values{"address": {address.Hex()}, "setLatest": {strconv.FormatBool(useLatest)}}, "Could not get can set use latest delegate for megapool status")
}

// Change a megapool's use-latest-delegate setting
func (c *Client) SetUseLatestDelegateMegapool(address common.Address, setting bool) (api.MegapoolSetUseLatestDelegateResponse, error) {
	settingStr := "false"
	if setting {
		settingStr = "true"
	}
	return c.callAPI[api.MegapoolSetUseLatestDelegateResponse]("POST", "/api/megapool/set-use-latest-delegate", url.Values{
		"address": {address.Hex()},
		"setting": {settingStr},
	}, "Could not set use latest delegate for megapool")
}

// Get the megapool's delegate address
func (c *Client) GetDelegate() (api.MegapoolGetDelegateResponse, error) {
	return c.callAPI[api.MegapoolGetDelegateResponse]("GET", "/api/megapool/get-delegate", nil, "Could get delegate for megapool")
}

// Get the megapool's effective delegate address
func (c *Client) GetEffectiveDelegate(address common.Address) (api.MegapoolGetEffectiveDelegateResponse, error) {
	return c.callAPI[api.MegapoolGetEffectiveDelegateResponse]("GET", "/api/megapool/get-effective-delegate", url.Values{"address": {address.Hex()}}, "Could get effective delegate for megapool")
}

// Calculate the megapool pending rewards
func (c *Client) CalculatePendingRewards() (api.MegapoolRewardSplitResponse, error) {
	return c.callAPI[api.MegapoolRewardSplitResponse]("GET", "/api/megapool/pending-rewards", nil, "Could not get pending rewards")
}

// Calculate rewards split given an arbitrary amount
func (c *Client) CalculateRewards(amountWei *big.Int) (api.MegapoolRewardSplitResponse, error) {
	return c.callAPI[api.MegapoolRewardSplitResponse]("GET", "/api/megapool/calculate-rewards", url.Values{"amountWei": {amountWei.String()}}, "Could not calculate rewards")
}

// Check if the node can distribute megapool rewards
func (c *Client) CanDistributeMegapool() (api.CanDistributeMegapoolResponse, error) {
	return c.callAPI[api.CanDistributeMegapoolResponse]("GET", "/api/megapool/can-distribute", nil, "Could not get can-distribute-megapool response")
}

// Distribute megapool rewards
func (c *Client) DistributeMegapool() (api.DistributeMegapoolResponse, error) {
	return c.callAPI[api.DistributeMegapoolResponse]("POST", "/api/megapool/distribute", nil, "Could not get distribute-megapool response")
}

// Get the validator withdrawals processed in the latest beacon block (with execution payload)
func (c *Client) GetLatestBlockWithdrawals() (api.LatestBlockWithdrawalsResponse, error) {
	return c.callAPI[api.LatestBlockWithdrawalsResponse]("GET", "/api/megapool/latest-block-withdrawals", nil, "Could not get latest block withdrawals")
}

// Get an estimate of the beacon chain withdrawal-sweep cycle time
func (c *Client) GetBeaconWithdrawalQueueEstimate() (api.BeaconWithdrawalQueueEstimateResponse, error) {
	return c.callAPI[api.BeaconWithdrawalQueueEstimateResponse]("GET", "/api/megapool/beacon-withdrawal-queue-estimate", nil, "Could not get beacon withdrawal queue estimate")
}

// Get the bond amount required for the megapool's next validator
func (c *Client) GetNewValidatorBondRequirement() (api.GetNewValidatorBondRequirementResponse, error) {
	return c.callAPI[api.GetNewValidatorBondRequirementResponse]("GET", "/api/megapool/get-new-validator-bond-requirement", nil, "Could not get new validator bond requirement")
}

// DissolveWithProof and CanDissolveWithProof client methods added above.
// CanDissolveWithProof / DissolveWithProof (also known as DissolveWithProof) are
// already implemented above.
