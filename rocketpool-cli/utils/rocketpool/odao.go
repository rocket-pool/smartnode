package rocketpool

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/dao/oracle"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type ODaoRequester struct {
	client *http.Client
}

func NewODaoRequester(client *http.Client) *ODaoRequester {
	return &ODaoRequester{
		client: client,
	}
}

func (r *ODaoRequester) GetName() string {
	return "ODAO"
}
func (r *ODaoRequester) GetRoute() string {
	return "odao"
}
func (r *ODaoRequester) GetClient() *http.Client {
	return r.client
}

// Join the oracle DAO (requires an executed invite proposal)
func (r *ODaoRequester) Join() (*api.ApiResponse[api.OracleDaoJoinData], error) {
	return sendGetRequest[api.OracleDaoJoinData](r, "join", "Join", nil)
}

// Leave the oracle DAO (requires an executed leave proposal)
func (r *ODaoRequester) Leave(bondRefundAddress common.Address) (*api.ApiResponse[api.OracleDaoLeaveData], error) {
	args := map[string]string{
		"bondRefundAddress": bondRefundAddress.Hex(),
	}
	return sendGetRequest[api.OracleDaoLeaveData](r, "leave", "Leave", args)
}

// Get oracle DAO members
func (r *ODaoRequester) Members() (*api.ApiResponse[api.OracleDaoMembersData], error) {
	return sendGetRequest[api.OracleDaoMembersData](r, "members", "Members", nil)
}

// Cancel a proposal made by the node
func (r *ODaoRequester) CancelProposal(id uint64) (*api.ApiResponse[api.OracleDaoCancelProposalData], error) {
	args := map[string]string{
		"id": fmt.Sprint(id),
	}
	return sendGetRequest[api.OracleDaoCancelProposalData](r, "proposal/cancel", "CancelProposal", args)
}

// Execute a proposal
func (r *ODaoRequester) ExecuteProposals(ids []uint64) (*api.ApiResponse[api.DataBatch[api.OracleDaoExecuteProposalData]], error) {
	args := map[string]string{
		"ids": makeBatchArg(ids),
	}
	return sendGetRequest[api.DataBatch[api.OracleDaoExecuteProposalData]](r, "proposal/execute", "ExecuteProposals", args)
}

// Vote on a proposal
func (r *ODaoRequester) Vote(id uint64, support bool) (*api.ApiResponse[api.OracleDaoVoteOnProposalData], error) {
	args := map[string]string{
		"id":      fmt.Sprint(id),
		"support": fmt.Sprint(support),
	}
	return sendGetRequest[api.OracleDaoVoteOnProposalData](r, "proposal/vote", "Vote", args)
}

// Get oracle DAO proposals
func (r *ODaoRequester) Proposals() (*api.ApiResponse[api.OracleDaoProposalsData], error) {
	return sendGetRequest[api.OracleDaoProposalsData](r, "proposals", "Proposals", nil)
}

// Propose inviting a new member
func (r *ODaoRequester) ProposeInvite(memberAddress common.Address, memberId string, memberUrl string) (*api.ApiResponse[api.OracleDaoProposeInviteData], error) {
	args := map[string]string{
		"address": memberAddress.Hex(),
		"id":      memberId,
		"url":     memberUrl,
	}
	return sendGetRequest[api.OracleDaoProposeInviteData](r, "propose-invite", "ProposeInvite", args)
}

// Propose kicking a member
func (r *ODaoRequester) ProposeKick(memberAddress common.Address, fineAmount *big.Int) (*api.ApiResponse[api.OracleDaoProposeKickData], error) {
	args := map[string]string{
		"address":    memberAddress.Hex(),
		"fineAmount": fineAmount.String(),
	}
	return sendGetRequest[api.OracleDaoProposeKickData](r, "propose-kick", "ProposeKick", args)
}

// Propose leaving the oracle DAO
func (r *ODaoRequester) ProposeLeave() (*api.ApiResponse[api.OracleDaoProposeLeaveData], error) {
	return sendGetRequest[api.OracleDaoProposeLeaveData](r, "propose-leave", "ProposeLeave", nil)
}

// Propose an Oracle DAO setting update
func (r *ODaoRequester) ProposeSetting(contractName rocketpool.ContractName, settingName oracle.SettingName, value string) (*api.ApiResponse[api.OracleDaoProposeSettingData], error) {
	args := map[string]string{
		"contract": string(contractName),
		"setting":  string(settingName),
		"value":    value,
	}
	return sendGetRequest[api.OracleDaoProposeSettingData](r, "setting/propose", "ProposeSetting", args)
}

// Get oracle DAO settings
func (r *ODaoRequester) Settings() (*api.ApiResponse[api.OracleDaoSettingsData], error) {
	return sendGetRequest[api.OracleDaoSettingsData](r, "settings", "Settings", nil)
}

// Get oracle DAO status
func (r *ODaoRequester) Status() (*api.ApiResponse[api.OracleDaoStatusData], error) {
	return sendGetRequest[api.OracleDaoStatusData](r, "status", "Status", nil)
}
