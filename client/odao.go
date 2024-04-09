package client

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/rocketpool-go/v2/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type ODaoRequester struct {
	context *client.RequesterContext
}

func NewODaoRequester(context *client.RequesterContext) *ODaoRequester {
	return &ODaoRequester{
		context: context,
	}
}

func (r *ODaoRequester) GetName() string {
	return "ODAO"
}
func (r *ODaoRequester) GetRoute() string {
	return "odao"
}
func (r *ODaoRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Join the oracle DAO (requires an executed invite proposal)
func (r *ODaoRequester) Join() (*types.ApiResponse[api.OracleDaoJoinData], error) {
	return client.SendGetRequest[api.OracleDaoJoinData](r, "join", "Join", nil)
}

// Leave the oracle DAO (requires an executed leave proposal)
func (r *ODaoRequester) Leave(bondRefundAddress common.Address) (*types.ApiResponse[api.OracleDaoLeaveData], error) {
	args := map[string]string{
		"bondRefundAddress": bondRefundAddress.Hex(),
	}
	return client.SendGetRequest[api.OracleDaoLeaveData](r, "leave", "Leave", args)
}

// Get oracle DAO members
func (r *ODaoRequester) Members() (*types.ApiResponse[api.OracleDaoMembersData], error) {
	return client.SendGetRequest[api.OracleDaoMembersData](r, "members", "Members", nil)
}

// Cancel a proposal made by the node
func (r *ODaoRequester) CancelProposal(id uint64) (*types.ApiResponse[api.OracleDaoCancelProposalData], error) {
	args := map[string]string{
		"id": fmt.Sprint(id),
	}
	return client.SendGetRequest[api.OracleDaoCancelProposalData](r, "proposal/cancel", "CancelProposal", args)
}

// Execute a proposal
func (r *ODaoRequester) ExecuteProposals(ids []uint64) (*types.ApiResponse[types.DataBatch[api.OracleDaoExecuteProposalData]], error) {
	args := map[string]string{
		"ids": client.MakeBatchArg(ids),
	}
	return client.SendGetRequest[types.DataBatch[api.OracleDaoExecuteProposalData]](r, "proposal/execute", "ExecuteProposals", args)
}

// Vote on a proposal
func (r *ODaoRequester) Vote(id uint64, support bool) (*types.ApiResponse[api.OracleDaoVoteOnProposalData], error) {
	args := map[string]string{
		"id":      fmt.Sprint(id),
		"support": fmt.Sprint(support),
	}
	return client.SendGetRequest[api.OracleDaoVoteOnProposalData](r, "proposal/vote", "Vote", args)
}

// Get oracle DAO proposals
func (r *ODaoRequester) Proposals() (*types.ApiResponse[api.OracleDaoProposalsData], error) {
	return client.SendGetRequest[api.OracleDaoProposalsData](r, "proposals", "Proposals", nil)
}

// Propose inviting a new member
func (r *ODaoRequester) ProposeInvite(memberAddress common.Address, memberId string, memberUrl string) (*types.ApiResponse[api.OracleDaoProposeInviteData], error) {
	args := map[string]string{
		"address": memberAddress.Hex(),
		"id":      memberId,
		"url":     memberUrl,
	}
	return client.SendGetRequest[api.OracleDaoProposeInviteData](r, "propose-invite", "ProposeInvite", args)
}

// Propose kicking a member
func (r *ODaoRequester) ProposeKick(memberAddress common.Address, fineAmount *big.Int) (*types.ApiResponse[api.OracleDaoProposeKickData], error) {
	args := map[string]string{
		"address":    memberAddress.Hex(),
		"fineAmount": fineAmount.String(),
	}
	return client.SendGetRequest[api.OracleDaoProposeKickData](r, "propose-kick", "ProposeKick", args)
}

// Propose leaving the oracle DAO
func (r *ODaoRequester) ProposeLeave() (*types.ApiResponse[api.OracleDaoProposeLeaveData], error) {
	return client.SendGetRequest[api.OracleDaoProposeLeaveData](r, "propose-leave", "ProposeLeave", nil)
}

// Propose an Oracle DAO setting update
func (r *ODaoRequester) ProposeSetting(contractName rocketpool.ContractName, settingName oracle.SettingName, value string) (*types.ApiResponse[api.OracleDaoProposeSettingData], error) {
	args := map[string]string{
		"contract": string(contractName),
		"setting":  string(settingName),
		"value":    value,
	}
	return client.SendGetRequest[api.OracleDaoProposeSettingData](r, "setting/propose", "ProposeSetting", args)
}

// Get oracle DAO settings
func (r *ODaoRequester) Settings() (*types.ApiResponse[api.OracleDaoSettingsData], error) {
	return client.SendGetRequest[api.OracleDaoSettingsData](r, "settings", "Settings", nil)
}

// Get oracle DAO status
func (r *ODaoRequester) Status() (*types.ApiResponse[api.OracleDaoStatusData], error) {
	return client.SendGetRequest[api.OracleDaoStatusData](r, "status", "Status", nil)
}
