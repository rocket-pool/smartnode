package client

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type PDaoRequester struct {
	context *client.RequesterContext
}

func NewPDaoRequester(context *client.RequesterContext) *PDaoRequester {
	return &PDaoRequester{
		context: context,
	}
}

func (r *PDaoRequester) GetName() string {
	return "PDAO"
}
func (r *PDaoRequester) GetRoute() string {
	return "pdao"
}
func (r *PDaoRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Claim / unlock bonds from a proposal
func (r *PDaoRequester) ClaimBonds(claims []api.ProtocolDaoClaimBonds) (*types.ApiResponse[types.DataBatch[api.ProtocolDaoClaimBondsData]], error) {
	body := api.ProtocolDaoClaimBondsBody{
		Claims: claims,
	}
	return client.SendPostRequest[types.DataBatch[api.ProtocolDaoClaimBondsData]](r, "claim-bonds", "ClaimBonds", body)
}

// Get the list of proposals with claimable / rewardable bonds, and the relevant indices for each one
func (r *PDaoRequester) GetClaimableBonds() (*types.ApiResponse[api.ProtocolDaoGetClaimableBondsData], error) {
	return client.SendGetRequest[api.ProtocolDaoGetClaimableBondsData](r, "get-claimable-bonds", "GetClaimableBonds", nil)
}

// Propose a one-time spend of the Protocol DAO's treasury
func (r *PDaoRequester) OneTimeSpend(invoiceID string, recipient common.Address, amount *big.Int) (*types.ApiResponse[api.ProtocolDaoGeneralProposeData], error) {
	args := map[string]string{
		"invoice-id": invoiceID,
		"recipient":  recipient.Hex(),
		"amount":     amount.String(),
	}
	return client.SendGetRequest[api.ProtocolDaoGeneralProposeData](r, "one-time-spend", "OneTimeSpend", args)
}

// Defeat a proposal if it still has an challenge after voting has started
func (r *PDaoRequester) DefeatProposal(proposalID uint64, index uint64) (*types.ApiResponse[api.ProtocolDaoDefeatProposalData], error) {
	args := map[string]string{
		"id":    fmt.Sprint(proposalID),
		"index": fmt.Sprint(index),
	}
	return client.SendGetRequest[api.ProtocolDaoDefeatProposalData](r, "proposal/defeat", "DefeatProposal", args)
}

// Execute one or more proposals
func (r *PDaoRequester) ExecuteProposals(ids []uint64) (*types.ApiResponse[types.DataBatch[api.ProtocolDaoExecuteProposalData]], error) {
	args := map[string]string{
		"ids": client.MakeBatchArg(ids),
	}
	return client.SendGetRequest[types.DataBatch[api.ProtocolDaoExecuteProposalData]](r, "proposal/execute", "ExecuteProposals", args)
}

// Finalize a proposal if it's been vetoed by burning the proposer's bond
func (r *PDaoRequester) FinalizeProposal(proposalID uint64) (*types.ApiResponse[api.ProtocolDaoFinalizeProposalData], error) {
	args := map[string]string{
		"id": fmt.Sprint(proposalID),
	}
	return client.SendGetRequest[api.ProtocolDaoFinalizeProposalData](r, "proposal/finalize", "FinalizeProposal", args)
}

// Override a delegate's vote on a proposal
func (r *PDaoRequester) OverrideVoteOnProposal(proposalID uint64, voteDirection rptypes.VoteDirection) (*types.ApiResponse[api.ProtocolDaoVoteOnProposalData], error) {
	args := map[string]string{
		"id":   fmt.Sprint(proposalID),
		"vote": api.VoteDirectionNameMap[voteDirection],
	}
	return client.SendGetRequest[api.ProtocolDaoVoteOnProposalData](r, "proposal/override-vote", "OverrideVoteOnProposal", args)
}

// Vote on a proposal
func (r *PDaoRequester) VoteOnProposal(proposalID uint64, voteDirection rptypes.VoteDirection) (*types.ApiResponse[api.ProtocolDaoVoteOnProposalData], error) {
	args := map[string]string{
		"id":   fmt.Sprint(proposalID),
		"vote": api.VoteDirectionNameMap[voteDirection],
	}
	return client.SendGetRequest[api.ProtocolDaoVoteOnProposalData](r, "proposal/vote", "VoteOnProposal", args)
}

// Get the Protocol DAO proposals
func (r *PDaoRequester) Proposals() (*types.ApiResponse[api.ProtocolDaoProposalsData], error) {
	return client.SendGetRequest[api.ProtocolDaoProposalsData](r, "proposals", "Proposals", nil)
}

// Propose a recurring spend of the Protocol DAO's treasury
func (r *PDaoRequester) RecurringSpend(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64) (*types.ApiResponse[api.ProtocolDaoGeneralProposeData], error) {
	args := map[string]string{
		"contract-name":     contractName,
		"recipient":         recipient.Hex(),
		"amount-per-period": amountPerPeriod.String(),
		"period-length":     periodLength.String(),
		"start-time":        startTime.Format(time.RFC3339),
		"num-periods":       fmt.Sprint(numberOfPeriods),
	}
	return client.SendGetRequest[api.ProtocolDaoGeneralProposeData](r, "recurring-spend", "RecurringSpend", args)
}

// Propose updating an existing recurring spend of the Protocol DAO's treasury
func (r *PDaoRequester) RecurringSpendUpdate(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64) (*types.ApiResponse[api.ProtocolDaoGeneralProposeData], error) {
	args := map[string]string{
		"contract-name":     contractName,
		"recipient":         recipient.Hex(),
		"amount-per-period": amountPerPeriod.String(),
		"period-length":     periodLength.String(),
		"num-periods":       fmt.Sprint(numberOfPeriods),
	}
	return client.SendGetRequest[api.ProtocolDaoGeneralProposeData](r, "recurring-spend-update", "RecurringSpendUpdate", args)
}

// Get the amount of minted RPL (and percentages) provided to node operators, the oDAO, and the pDAO at each rewards period
func (r *PDaoRequester) RewardsPercentages() (*types.ApiResponse[api.ProtocolDaoRewardsPercentagesData], error) {
	return client.SendGetRequest[api.ProtocolDaoRewardsPercentagesData](r, "rewards-percentages", "RewardsPercentages", nil)
}

// Propose new RPL rewards percentages for node operators, the oDAO, and the pDAO at each rewards period
func (r *PDaoRequester) ProposeRewardsPercentages(node *big.Int, odao *big.Int, pdao *big.Int) (*types.ApiResponse[api.ProtocolDaoGeneralProposeData], error) {
	args := map[string]string{
		"node": node.String(),
		"odao": odao.String(),
		"pdao": pdao.String(),
	}
	return client.SendGetRequest[api.ProtocolDaoGeneralProposeData](r, "rewards-percentages/proposee", "ProposeRewardsPercentages", args)
}

// Propose inviting someone to the security council
func (r *PDaoRequester) InviteToSecurityCouncil(id string, address common.Address) (*types.ApiResponse[api.ProtocolDaoProposeInviteToSecurityCouncilData], error) {
	args := map[string]string{
		"id":      id,
		"address": address.Hex(),
	}
	return client.SendGetRequest[api.ProtocolDaoProposeInviteToSecurityCouncilData](r, "security/invite", "InviteToSecurityCouncil", args)
}

// Propose kicking someone from the security council
func (r *PDaoRequester) KickFromSecurityCouncil(address common.Address) (*types.ApiResponse[api.ProtocolDaoProposeKickFromSecurityCouncilData], error) {
	args := map[string]string{
		"address": address.Hex(),
	}
	return client.SendGetRequest[api.ProtocolDaoProposeKickFromSecurityCouncilData](r, "security/kick", "KickFromSecurityCouncil", args)
}

// Propose kicking multiple members from the security council
func (r *PDaoRequester) KickMultiFromSecurityCouncil(addresses []common.Address) (*types.ApiResponse[api.ProtocolDaoProposeKickMultiFromSecurityCouncilData], error) {
	args := map[string]string{
		"addresses": client.MakeBatchArg(addresses),
	}
	return client.SendGetRequest[api.ProtocolDaoProposeKickMultiFromSecurityCouncilData](r, "security/kick-multi", "KickMultiFromSecurityCouncil", args)
}

// Propose replacing someone on the security council with a new member to invite
func (r *PDaoRequester) ReplaceMemberOfSecurityCouncil(existingAddress common.Address, newID string, newAddress common.Address) (*types.ApiResponse[api.ProtocolDaoProposeReplaceMemberOfSecurityCouncilData], error) {
	args := map[string]string{
		"existing-address": existingAddress.Hex(),
		"new-id":           newID,
		"new-address":      newAddress.Hex(),
	}
	return client.SendGetRequest[api.ProtocolDaoProposeReplaceMemberOfSecurityCouncilData](r, "security/replace", "ReplaceMemberOfSecurityCouncil", args)
}

// Get the Protocol DAO settings
func (r *PDaoRequester) Settings() (*types.ApiResponse[api.ProtocolDaoSettingsData], error) {
	return client.SendGetRequest[api.ProtocolDaoSettingsData](r, "settings", "Settings", nil)
}

// Propose updating one of the Protocol DAO settings
func (r *PDaoRequester) ProposeSetting(contractName rocketpool.ContractName, setting protocol.SettingName, value string) (*types.ApiResponse[api.ProtocolDaoProposeSettingData], error) {
	args := map[string]string{
		"contract": string(contractName),
		"setting":  string(setting),
		"value":    value,
	}
	return client.SendGetRequest[api.ProtocolDaoProposeSettingData](r, "setting/propose", "ProposeSetting", args)
}

// Initialize voting so the node can vote on Protocol DAO proposals
func (r *PDaoRequester) InitializeVoting() (*types.ApiResponse[api.ProtocolDaoInitializeVotingData], error) {
	return client.SendGetRequest[api.ProtocolDaoInitializeVotingData](r, "initialize-voting", "InitializeVoting", nil)
}

// Set the delegate for voting on Protocol DAO proposals
func (r *PDaoRequester) SetVotingDelegate(delegate common.Address) (*types.ApiResponse[types.TxInfoData], error) {
	args := map[string]string{
		"delegate": delegate.Hex(),
	}
	return client.SendGetRequest[types.TxInfoData](r, "voting-delegate/set", "SetVotingDelegate", args)
}

// Get the address that's assigned as the delegate for voting on Protocol DAO proposals
func (r *PDaoRequester) GetCurrentVotingDelegate() (*types.ApiResponse[api.ProtocolDaoCurrentVotingDelegateData], error) {
	return client.SendGetRequest[api.ProtocolDaoCurrentVotingDelegateData](r, "voting-delegate", "GetCurrentVotingDelegate", nil)
}
