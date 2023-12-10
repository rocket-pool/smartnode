package rocketpool

import (
	"fmt"
	"net/http"
	"strings"

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

// Get the list of proposals with claimable / rewardable bonds, and the relevant indices for each one
func (r *PDaoRequester) GetClaimableBonds() (*api.ApiResponse[api.ProtocolDaoGetClaimableBondsData], error) {
	return sendGetRequest[api.ProtocolDaoGetClaimableBondsData](r, "get-claimable-bonds", "GetClaimableBonds", nil)
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
