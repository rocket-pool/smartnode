package rocketpool

import (
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type PDaoRequester struct {
	client *http.Client
}

func NewPDaoRequester(client *http.Client) *PDaoRequester {
	return &PDaoRequester{
		client: client,
	}
}

func (r *PDaoRequester) GetName() string {
	return "PDAO"
}
func (r *PDaoRequester) GetRoute() string {
	return "pdao"
}
func (r *PDaoRequester) GetClient() *http.Client {
	return r.client
}

// Claim / unlock bonds from a proposal
func (r *PDaoRequester) ClaimBonds(claims []api.ProtocolDaoClaimBonds) (*api.ApiResponse[api.DataBatch[api.ProtocolDaoClaimBondsData]], error) {
	body := api.ProtocolDaoClaimBondsBody{
		Claims: claims,
	}
	return sendPostRequest[api.DataBatch[api.ProtocolDaoClaimBondsData]](r, "claim-bonds", "ClaimBonds", body)
}

// Get the list of proposals with claimable / rewardable bonds, and the relevant indices for each one
func (r *PDaoRequester) GetClaimableBonds() (*api.ApiResponse[api.ProtocolDaoGetClaimableBondsData], error) {
	return sendGetRequest[api.ProtocolDaoGetClaimableBondsData](r, "get-claimable-bonds", "GetClaimableBonds", nil)
}

// Propose a one-time spend of the Protocol DAO's treasury
func (r *PDaoRequester) OneTimeSpend(invoiceID string, recipient common.Address, amount *big.Int) (*api.ApiResponse[api.ProtocolDaoGeneralProposeData], error) {
	args := map[string]string{
		"invoice-id": invoiceID,
		"recipient":  recipient.Hex(),
		"amount":     amount.String(),
	}
	return sendGetRequest[api.ProtocolDaoGeneralProposeData](r, "one-time-spend", "OneTimeSpend", args)
}

// Defeat a proposal if it still has an challenge after voting has started
func (r *PDaoRequester) DefeatProposal(proposalID uint64, index uint64) (*api.ApiResponse[api.ProtocolDaoDefeatProposalData], error) {
	args := map[string]string{
		"id":    fmt.Sprint(proposalID),
		"index": fmt.Sprint(index),
	}
	return sendGetRequest[api.ProtocolDaoDefeatProposalData](r, "proposal/defeat", "DefeatProposal", args)
}

// Execute one or more proposals
func (r *PDaoRequester) ExecuteProposals(ids []uint64) (*api.ApiResponse[api.DataBatch[api.ProtocolDaoExecuteProposalData]], error) {
	args := map[string]string{
		"ids": makeBatchArg(ids),
	}
	return sendGetRequest[api.DataBatch[api.ProtocolDaoExecuteProposalData]](r, "proposal/execute", "ExecuteProposals", args)
}

// Finalize a proposal if it's been vetoed by burning the proposer's bond
func (r *PDaoRequester) FinalizeProposal(proposalID uint64) (*api.ApiResponse[api.ProtocolDaoFinalizeProposalData], error) {
	args := map[string]string{
		"id": fmt.Sprint(proposalID),
	}
	return sendGetRequest[api.ProtocolDaoFinalizeProposalData](r, "proposal/finalize", "FinalizeProposal", args)
}

// Override a delegate's vote on a proposal
func (r *PDaoRequester) OverrideVoteOnProposal(proposalID uint64, voteDirection types.VoteDirection) (*api.ApiResponse[api.ProtocolDaoVoteOnProposalData], error) {
	args := map[string]string{
		"id":   fmt.Sprint(proposalID),
		"vote": api.VoteDirectionNameMap[voteDirection],
	}
	return sendGetRequest[api.ProtocolDaoVoteOnProposalData](r, "proposal/override-vote", "OverrideVoteOnProposal", args)
}

// Vote on a proposal
func (r *PDaoRequester) VoteOnProposal(proposalID uint64, voteDirection types.VoteDirection) (*api.ApiResponse[api.ProtocolDaoVoteOnProposalData], error) {
	args := map[string]string{
		"id":   fmt.Sprint(proposalID),
		"vote": api.VoteDirectionNameMap[voteDirection],
	}
	return sendGetRequest[api.ProtocolDaoVoteOnProposalData](r, "proposal/vote", "VoteOnProposal", args)
}

// Get the Protocol DAO proposals
func (r *PDaoRequester) Proposals() (*api.ApiResponse[api.ProtocolDaoProposalsData], error) {
	return sendGetRequest[api.ProtocolDaoProposalsData](r, "proposals", "Proposals", nil)
}

// Propose a recurring spend of the Protocol DAO's treasury
func (r *PDaoRequester) RecurringSpend(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, startTime time.Time, numberOfPeriods uint64) (*api.ApiResponse[api.ProtocolDaoGeneralProposeData], error) {
	args := map[string]string{
		"contract-name":     contractName,
		"recipient":         recipient.Hex(),
		"amount-per-period": amountPerPeriod.String(),
		"period-length":     periodLength.String(),
		"start-time":        startTime.Format(time.RFC3339),
		"num-periods":       fmt.Sprint(numberOfPeriods),
	}
	return sendGetRequest[api.ProtocolDaoGeneralProposeData](r, "recurring-spend", "RecurringSpend", args)
}

// Propose updating an existing recurring spend of the Protocol DAO's treasury
func (r *PDaoRequester) RecurringSpendUpdate(contractName string, recipient common.Address, amountPerPeriod *big.Int, periodLength time.Duration, numberOfPeriods uint64) (*api.ApiResponse[api.ProtocolDaoGeneralProposeData], error) {
	args := map[string]string{
		"contract-name":     contractName,
		"recipient":         recipient.Hex(),
		"amount-per-period": amountPerPeriod.String(),
		"period-length":     periodLength.String(),
		"num-periods":       fmt.Sprint(numberOfPeriods),
	}
	return sendGetRequest[api.ProtocolDaoGeneralProposeData](r, "recurring-spend-update", "RecurringSpendUpdate", args)
}

// Get the amount of minted RPL (and percentages) provided to node operators, the oDAO, and the pDAO at each rewards period
func (r *PDaoRequester) RewardsPercentages() (*api.ApiResponse[api.ProtocolDaoRewardsPercentagesData], error) {
	return sendGetRequest[api.ProtocolDaoRewardsPercentagesData](r, "rewards-percentages", "RewardsPercentages", nil)
}

// Propose new RPL rewards percentages for node operators, the oDAO, and the pDAO at each rewards period
func (r *PDaoRequester) ProposeRewardsPercentages(node *big.Int, odao *big.Int, pdao *big.Int) (*api.ApiResponse[api.ProtocolDaoGeneralProposeData], error) {
	args := map[string]string{
		"node": node.String(),
		"odao": odao.String(),
		"pdao": pdao.String(),
	}
	return sendGetRequest[api.ProtocolDaoGeneralProposeData](r, "rewards-percentages/proposee", "ProposeRewardsPercentages", args)
}

// Propose inviting someone to the security council
func (r *PDaoRequester) InviteToSecurityCouncil(id string, address common.Address) (*api.ApiResponse[api.ProtocolDaoProposeInviteToSecurityCouncilData], error) {
	args := map[string]string{
		"id":      id,
		"address": address.Hex(),
	}
	return sendGetRequest[api.ProtocolDaoProposeInviteToSecurityCouncilData](r, "security/invite", "InviteToSecurityCouncil", args)
}

// Propose kicking someone from the security council
func (r *PDaoRequester) KickFromSecurityCouncil(address common.Address) (*api.ApiResponse[api.ProtocolDaoProposeKickFromSecurityCouncilData], error) {
	args := map[string]string{
		"address": address.Hex(),
	}
	return sendGetRequest[api.ProtocolDaoProposeKickFromSecurityCouncilData](r, "security/kick", "KickFromSecurityCouncil", args)
}

// Propose kicking multiple members from the security council
func (r *PDaoRequester) KickMultiFromSecurityCouncil(addresses []common.Address) (*api.ApiResponse[api.ProtocolDaoProposeKickMultiFromSecurityCouncilData], error) {
	args := map[string]string{
		"addresses": makeBatchArg(addresses),
	}
	return sendGetRequest[api.ProtocolDaoProposeKickMultiFromSecurityCouncilData](r, "security/kick-multi", "KickMultiFromSecurityCouncil", args)
}

// Propose replacing someone on the security council with a new member to invite
func (r *PDaoRequester) ReplaceMemberOfSecurityCouncil(existingAddress common.Address, newID string, newAddress common.Address) (*api.ApiResponse[api.ProtocolDaoProposeReplaceMemberOfSecurityCouncilData], error) {
	args := map[string]string{
		"existing-address": existingAddress.Hex(),
		"new-id":           newID,
		"new-address":      newAddress.Hex(),
	}
	return sendGetRequest[api.ProtocolDaoProposeReplaceMemberOfSecurityCouncilData](r, "security/replace", "ReplaceMemberOfSecurityCouncil", args)
}

// Get the Protocol DAO settings
func (r *PDaoRequester) Settings() (*api.ApiResponse[api.ProtocolDaoSettingsData], error) {
	return sendGetRequest[api.ProtocolDaoSettingsData](r, "settings", "Settings", nil)
}

// Propose updating one of the Protocol DAO settings
func (r *PDaoRequester) ProposeSetting(contractName rocketpool.ContractName, setting protocol.SettingName, value string) (*api.ApiResponse[api.ProtocolDaoProposeSettingData], error) {
	args := map[string]string{
		"contract": string(contractName),
		"setting":  string(setting),
		"value":    value,
	}
	return sendGetRequest[api.ProtocolDaoProposeSettingData](r, "setting/propose", "ProposeSetting", args)
}
