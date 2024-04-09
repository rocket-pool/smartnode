package client

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type SecurityRequester struct {
	context *client.RequesterContext
}

func NewSecurityRequester(context *client.RequesterContext) *SecurityRequester {
	return &SecurityRequester{
		context: context,
	}
}

func (r *SecurityRequester) GetName() string {
	return "Security Council"
}
func (r *SecurityRequester) GetRoute() string {
	return "security"
}
func (r *SecurityRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Join the security council after being invited
func (r *SecurityRequester) Join() (*types.ApiResponse[api.SecurityJoinData], error) {
	return client.SendGetRequest[api.SecurityJoinData](r, "join", "Join", nil)
}

// Leave the security council after the proposal to leave has passed
func (r *SecurityRequester) Leave() (*types.ApiResponse[api.SecurityLeaveData], error) {
	return client.SendGetRequest[api.SecurityLeaveData](r, "leave", "Leave", nil)
}

// Get info about the security council members
func (r *SecurityRequester) Members() (*types.ApiResponse[api.SecurityMembersData], error) {
	return client.SendGetRequest[api.SecurityMembersData](r, "members", "Members", nil)
}

// Cancel a proposal made by the node
func (r *SecurityRequester) CancelProposal(id uint64) (*types.ApiResponse[api.SecurityCancelProposalData], error) {
	args := map[string]string{
		"id": fmt.Sprint(id),
	}
	return client.SendGetRequest[api.SecurityCancelProposalData](r, "proposal/cancel", "CancelProposal", args)
}

// Execute a proposal
func (r *SecurityRequester) ExecuteProposals(ids []uint64) (*types.ApiResponse[types.DataBatch[api.SecurityExecuteProposalData]], error) {
	args := map[string]string{
		"ids": client.MakeBatchArg(ids),
	}
	return client.SendGetRequest[types.DataBatch[api.SecurityExecuteProposalData]](r, "proposal/execute", "ExecuteProposals", args)
}

// Vote on a proposal
func (r *SecurityRequester) VoteOnProposal(id uint64, support bool) (*types.ApiResponse[api.SecurityVoteOnProposalData], error) {
	args := map[string]string{
		"id":      fmt.Sprint(id),
		"support": fmt.Sprint(support),
	}
	return client.SendGetRequest[api.SecurityVoteOnProposalData](r, "proposal/vote", "VoteOnProposal", args)
}

// Get info about the security council proposals
func (r *SecurityRequester) Proposals() (*types.ApiResponse[api.SecurityProposalsData], error) {
	return client.SendGetRequest[api.SecurityProposalsData](r, "proposals", "Proposals", nil)
}

// Request leaving the security council
func (r *SecurityRequester) ProposeLeave() (*types.ApiResponse[types.TxInfoData], error) {
	return client.SendGetRequest[types.TxInfoData](r, "propose-leave", "ProposeLeave", nil)
}

// Propose a Protocol DAO (security council) setting update
func (r *SecurityRequester) ProposeSetting(contractName rocketpool.ContractName, settingName protocol.SettingName, value string) (*types.ApiResponse[api.SecurityProposeSettingData], error) {
	args := map[string]string{
		"contract": string(contractName),
		"setting":  string(settingName),
		"value":    value,
	}
	return client.SendGetRequest[api.SecurityProposeSettingData](r, "setting/propose", "ProposeSetting", args)
}

// Get info about the security council
func (r *SecurityRequester) Status() (*types.ApiResponse[api.SecurityStatusData], error) {
	return client.SendGetRequest[api.SecurityStatusData](r, "status", "Status", nil)
}
