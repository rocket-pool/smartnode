package rocketpool

import (
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
func (r *PDaoRequester) ClaimBonds(proposalID uint64, indices []uint64) (*api.ApiResponse[api.ProtocolDaoClaimBondsData], error) {
	indicesStrings := make([]string, len(indices))
	for i, index := range indices {
		indicesStrings[i] = fmt.Sprint(index)
	}
	args := map[string]string{
		"proposal-id": fmt.Sprint(proposalID),
		"indices":     strings.Join(indicesStrings, ","),
	}
	return sendGetRequest[api.ProtocolDaoClaimBondsData](r, "claim-bonds", "ClaimBonds", args)
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

// Execute a proposal
func (r *PDaoRequester) ExecuteProposal(proposalID uint64) (*api.ApiResponse[api.ProtocolDaoExecuteProposalData], error) {
	args := map[string]string{
		"id": fmt.Sprint(proposalID),
	}
	return sendGetRequest[api.ProtocolDaoExecuteProposalData](r, "proposal/execute", "ExecuteProposal", args)
}

// Finalize a proposal if it's been vetoed by burning the proposer's bond
func (r *PDaoRequester) FinalizeProposal(proposalID uint64) (*api.ApiResponse[api.ProtocolDaoFinalizeProposalData], error) {
	args := map[string]string{
		"id": fmt.Sprint(proposalID),
	}
	return sendGetRequest[api.ProtocolDaoFinalizeProposalData](r, "proposal/finalize", "FinalizeProposal", args)
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

// Get the Protocol DAO settings
func (r *PDaoRequester) Settings() (*api.ApiResponse[api.ProtocolDaoSettingsData], error) {
	return sendGetRequest[api.ProtocolDaoSettingsData](r, "settings", "Settings", nil)
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
	addressStrings := make([]string, len(addresses))
	for i, address := range addresses {
		addressStrings[i] = address.Hex()
	}
	args := map[string]string{
		"addresses": strings.Join(addressStrings, ","),
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
