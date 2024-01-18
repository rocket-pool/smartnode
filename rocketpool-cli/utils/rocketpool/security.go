package rocketpool

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type SecurityRequester struct {
	client *http.Client
}

func NewSecurityRequester(client *http.Client) *SecurityRequester {
	return &SecurityRequester{
		client: client,
	}
}

func (r *SecurityRequester) GetName() string {
	return "Security Council"
}
func (r *SecurityRequester) GetRoute() string {
	return "security"
}
func (r *SecurityRequester) GetClient() *http.Client {
	return r.client
}

// Join the security council after being invited
func (r *SecurityRequester) Join() (*api.ApiResponse[api.SecurityJoinData], error) {
	return sendGetRequest[api.SecurityJoinData](r, "join", "Join", nil)
}

// Leave the security council after the proposal to leave has passed
func (r *SecurityRequester) Leave() (*api.ApiResponse[api.SecurityLeaveData], error) {
	return sendGetRequest[api.SecurityLeaveData](r, "leave", "Leave", nil)
}

// Get info about the security council members
func (r *SecurityRequester) Members() (*api.ApiResponse[api.SecurityMembersData], error) {
	return sendGetRequest[api.SecurityMembersData](r, "members", "Members", nil)
}

// Cancel a proposal made by the node
func (r *SecurityRequester) CancelProposal(id uint64) (*api.ApiResponse[api.SecurityCancelProposalData], error) {
	args := map[string]string{
		"id": fmt.Sprint(id),
	}
	return sendGetRequest[api.SecurityCancelProposalData](r, "proposal/cancel", "CancelProposal", args)
}

// Execute a proposal
func (r *SecurityRequester) ExecuteProposals(ids []uint64) (*api.ApiResponse[api.DataBatch[api.SecurityExecuteProposalData]], error) {
	args := map[string]string{
		"ids": makeBatchArg(ids),
	}
	return sendGetRequest[api.DataBatch[api.SecurityExecuteProposalData]](r, "proposal/execute", "ExecuteProposals", args)
}

// Vote on a proposal
func (r *SecurityRequester) VoteOnProposal(id uint64, support bool) (*api.ApiResponse[api.SecurityVoteOnProposalData], error) {
	args := map[string]string{
		"id":      fmt.Sprint(id),
		"support": fmt.Sprint(support),
	}
	return sendGetRequest[api.SecurityVoteOnProposalData](r, "proposal/vote", "VoteOnProposal", args)
}

// Get info about the security council proposals
func (r *SecurityRequester) Proposals() (*api.ApiResponse[api.SecurityProposalsData], error) {
	return sendGetRequest[api.SecurityProposalsData](r, "proposals", "Proposals", nil)
}

// Invite a new member to the security council
func (r *SecurityRequester) ProposeInvite(id string, address common.Address) (*api.ApiResponse[api.SecurityProposeInviteData], error) {
	args := map[string]string{
		"id":      id,
		"address": address.Hex(),
	}
	return sendGetRequest[api.SecurityProposeInviteData](r, "propose-invite", "ProposeInvite", args)
}

// Request leaving the security council
func (r *SecurityRequester) ProposeLeave() (*api.ApiResponse[api.TxInfoData], error) {
	return sendGetRequest[api.TxInfoData](r, "propose-leave", "ProposeLeave", nil)
}

// Kick a member from the security council
func (r *SecurityRequester) ProposeKick(address common.Address) (*api.ApiResponse[api.SecurityProposeKickData], error) {
	args := map[string]string{
		"address": address.Hex(),
	}
	return sendGetRequest[api.SecurityProposeKickData](r, "propose-kick", "ProposeKick", args)
}

// Kick multiple members of the security council
func (r *SecurityRequester) ProposeKickMulti(addresses []common.Address) (*api.ApiResponse[api.SecurityProposeKickMultiData], error) {
	args := map[string]string{
		"addresses": makeBatchArg(addresses),
	}
	return sendGetRequest[api.SecurityProposeKickMultiData](r, "propose-kick-multi", "ProposeKickMulti", args)
}

// Replace a member of the security council with a new member
func (r *SecurityRequester) ProposeReplace(existingAddress common.Address, newID string, newAddress common.Address) (*api.ApiResponse[api.SecurityProposeReplaceData], error) {
	args := map[string]string{
		"existing-address": existingAddress.Hex(),
		"new-id":           newID,
		"new-address":      newAddress.Hex(),
	}
	return sendGetRequest[api.SecurityProposeReplaceData](r, "propose-replace", "ProposeReplace", args)
}

// Propose a Protocol DAO (security council) setting update
func (r *SecurityRequester) ProposeSetting(contractName rocketpool.ContractName, settingName protocol.SettingName, value string) (*api.ApiResponse[api.SecurityProposeSettingData], error) {
	args := map[string]string{
		"contract": string(contractName),
		"setting":  string(settingName),
		"value":    value,
	}
	return sendGetRequest[api.SecurityProposeSettingData](r, "setting/propose", "ProposeSetting", args)
}

// Get info about the security council
func (r *SecurityRequester) Status() (*api.ApiResponse[api.SecurityStatusData], error) {
	return sendGetRequest[api.SecurityStatusData](r, "status", "Status", nil)
}
