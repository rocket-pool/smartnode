package rocketpool

import (
	"math/big"
	"net/url"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get oracle DAO status
func (c *Client) TNDAOStatus() (api.TNDAOStatusResponse, error) {
	return c.callAPI[api.TNDAOStatusResponse]("GET", "/api/odao/status", nil, "Could not get oracle DAO status")
}

// Get oracle DAO members
func (c *Client) TNDAOMembers() (api.TNDAOMembersResponse, error) {
	response, err := c.callAPI[api.TNDAOMembersResponse]("GET", "/api/odao/members", nil, "Could not get oracle DAO members")
	if err != nil {
		return response, err
	}
	for i := 0; i < len(response.Members); i++ {
		member := &response.Members[i]
		if member.RPLBondAmount == nil {
			member.RPLBondAmount = big.NewInt(0)
		}
	}
	return response, nil
}

// Get oracle DAO proposals
func (c *Client) TNDAOProposals() (api.TNDAOProposalsResponse, error) {
	return c.callAPI[api.TNDAOProposalsResponse]("GET", "/api/odao/proposals", nil, "Could not get oracle DAO proposals")
}

// Get a single oracle DAO proposal
func (c *Client) TNDAOProposal(id uint64) (api.TNDAOProposalResponse, error) {
	return c.callAPI[api.TNDAOProposalResponse]("GET", "/api/odao/proposal-details", url.Values{"id": {strconv.FormatUint(id, 10)}}, "Could not get oracle DAO proposal")
}

// Check whether the node can propose inviting a new member
func (c *Client) CanProposeInviteToTNDAO(memberAddress common.Address, memberId, memberUrl string) (api.CanProposeTNDAOInviteResponse, error) {
	return c.callAPI[api.CanProposeTNDAOInviteResponse]("GET", "/api/odao/can-propose-invite", url.Values{
		"address":   {memberAddress.Hex()},
		"memberId":  {memberId},
		"memberUrl": {memberUrl},
	}, "Could not get can propose oracle DAO invite status")
}

// Propose inviting a new member
func (c *Client) ProposeInviteToTNDAO(memberAddress common.Address, memberId, memberUrl string) (api.ProposeTNDAOInviteResponse, error) {
	return c.callAPI[api.ProposeTNDAOInviteResponse]("POST", "/api/odao/propose-invite", url.Values{
		"address":   {memberAddress.Hex()},
		"memberId":  {memberId},
		"memberUrl": {memberUrl},
	}, "Could not propose oracle DAO invite")
}

// Check whether the node can propose leaving the oracle DAO
func (c *Client) CanProposeLeaveTNDAO() (api.CanProposeTNDAOLeaveResponse, error) {
	return c.callAPI[api.CanProposeTNDAOLeaveResponse]("GET", "/api/odao/can-propose-leave", nil, "Could not get can propose leaving oracle DAO status")
}

// Propose leaving the oracle DAO
func (c *Client) ProposeLeaveTNDAO() (api.ProposeTNDAOLeaveResponse, error) {
	return c.callAPI[api.ProposeTNDAOLeaveResponse]("POST", "/api/odao/propose-leave", nil, "Could not propose leaving oracle DAO")
}

// Check whether the node can propose kicking a member
func (c *Client) CanProposeKickFromTNDAO(memberAddress common.Address, fineAmountWei *big.Int) (api.CanProposeTNDAOKickResponse, error) {
	return c.callAPI[api.CanProposeTNDAOKickResponse]("GET", "/api/odao/can-propose-kick", url.Values{
		"address":       {memberAddress.Hex()},
		"fineAmountWei": {fineAmountWei.String()},
	}, "Could not get can propose kicking oracle DAO member status")
}

// Propose kicking a member
func (c *Client) ProposeKickFromTNDAO(memberAddress common.Address, fineAmountWei *big.Int) (api.ProposeTNDAOKickResponse, error) {
	return c.callAPI[api.ProposeTNDAOKickResponse]("POST", "/api/odao/propose-kick", url.Values{
		"address":       {memberAddress.Hex()},
		"fineAmountWei": {fineAmountWei.String()},
	}, "Could not propose kicking oracle DAO member")
}

// Check whether the node can cancel a proposal
func (c *Client) CanCancelTNDAOProposal(proposalId uint64) (api.CanCancelTNDAOProposalResponse, error) {
	return c.callAPI[api.CanCancelTNDAOProposalResponse]("GET", "/api/odao/can-cancel-proposal", url.Values{"id": {strconv.FormatUint(proposalId, 10)}}, "Could not get can cancel oracle DAO proposal status")
}

// Cancel a proposal made by the node
func (c *Client) CancelTNDAOProposal(proposalId uint64) (api.CancelTNDAOProposalResponse, error) {
	return c.callAPI[api.CancelTNDAOProposalResponse]("POST", "/api/odao/cancel-proposal", url.Values{"id": {strconv.FormatUint(proposalId, 10)}}, "Could not cancel oracle DAO proposal")
}

// Check whether the node can vote on a proposal
func (c *Client) CanVoteOnTNDAOProposal(proposalId uint64) (api.CanVoteOnTNDAOProposalResponse, error) {
	return c.callAPI[api.CanVoteOnTNDAOProposalResponse]("GET", "/api/odao/can-vote-proposal", url.Values{"id": {strconv.FormatUint(proposalId, 10)}}, "Could not get can vote on oracle DAO proposal status")
}

// Vote on a proposal
func (c *Client) VoteOnTNDAOProposal(proposalId uint64, support bool) (api.VoteOnTNDAOProposalResponse, error) {
	supportStr := "false"
	if support {
		supportStr = "true"
	}
	return c.callAPI[api.VoteOnTNDAOProposalResponse]("POST", "/api/odao/vote-proposal", url.Values{
		"id":      {strconv.FormatUint(proposalId, 10)},
		"support": {supportStr},
	}, "Could not vote on oracle DAO proposal")
}

// Check whether the node can execute a proposal
func (c *Client) CanExecuteTNDAOProposal(proposalId uint64) (api.CanExecuteTNDAOProposalResponse, error) {
	return c.callAPI[api.CanExecuteTNDAOProposalResponse]("GET", "/api/odao/can-execute-proposal", url.Values{"id": {strconv.FormatUint(proposalId, 10)}}, "Could not get can execute oracle DAO proposal status")
}

// Execute a proposal
func (c *Client) ExecuteTNDAOProposal(proposalId uint64) (api.ExecuteTNDAOProposalResponse, error) {
	return c.callAPI[api.ExecuteTNDAOProposalResponse]("POST", "/api/odao/execute-proposal", url.Values{"id": {strconv.FormatUint(proposalId, 10)}}, "Could not execute oracle DAO proposal")
}

// Check whether the node can join the oracle DAO
func (c *Client) CanJoinTNDAO() (api.CanJoinTNDAOResponse, error) {
	return c.callAPI[api.CanJoinTNDAOResponse]("GET", "/api/odao/can-join", nil, "Could not get can join oracle DAO status")
}

// Approve RPL for joining the oracle DAO
func (c *Client) ApproveRPLToJoinTNDAO() (api.JoinTNDAOApproveResponse, error) {
	return c.callAPI[api.JoinTNDAOApproveResponse]("POST", "/api/odao/join-approve-rpl", nil, "Could not approve RPL for joining oracle DAO")
}

// Join the oracle DAO (requires an executed invite proposal)
func (c *Client) JoinTNDAO(approvalTxHash common.Hash) (api.JoinTNDAOJoinResponse, error) {
	return c.callAPI[api.JoinTNDAOJoinResponse]("POST", "/api/odao/join", url.Values{"approvalTxHash": {approvalTxHash.String()}}, "Could not join oracle DAO")
}

// Check whether the node can leave the oracle DAO
func (c *Client) CanLeaveTNDAO() (api.CanLeaveTNDAOResponse, error) {
	return c.callAPI[api.CanLeaveTNDAOResponse]("GET", "/api/odao/can-leave", nil, "Could not get can leave oracle DAO status")
}

// Leave the oracle DAO (requires an executed leave proposal)
func (c *Client) LeaveTNDAO(bondRefundAddress common.Address) (api.LeaveTNDAOResponse, error) {
	return c.callAPI[api.LeaveTNDAOResponse]("POST", "/api/odao/leave", url.Values{"bondRefundAddress": {bondRefundAddress.Hex()}}, "Could not leave oracle DAO")
}

func (c *Client) CanProposeTNDAOSettingMembersQuorum(quorum float64) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-members-quorum", url.Values{"quorum": {strconv.FormatFloat(quorum, 'f', -1, 64)}}, "Could not get can propose setting members.quorum")
}

func (c *Client) CanProposeTNDAOSettingMembersRplBond(bondAmountWei *big.Int) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-members-rplbond", url.Values{"bondAmountWei": {bondAmountWei.String()}}, "Could not get can propose setting members.rplbond")
}

func (c *Client) CanProposeTNDAOSettingProposalCooldown(proposalCooldownTimespan uint64) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-proposal-cooldown", url.Values{"value": {strconv.FormatUint(proposalCooldownTimespan, 10)}}, "Could not get can propose setting proposal.cooldown.time")
}

func (c *Client) CanProposeTNDAOSettingProposalVoteTimespan(proposalVoteTimespan uint64) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-proposal-vote-timespan", url.Values{"value": {strconv.FormatUint(proposalVoteTimespan, 10)}}, "Could not get can propose setting proposal.vote.time")
}

func (c *Client) CanProposeTNDAOSettingProposalVoteDelayTimespan(proposalDelayTimespan uint64) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-proposal-vote-delay-timespan", url.Values{"value": {strconv.FormatUint(proposalDelayTimespan, 10)}}, "Could not get can propose setting proposal.vote.delay.time")
}

func (c *Client) CanProposeTNDAOSettingProposalExecuteTimespan(proposalExecuteTimespan uint64) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-proposal-execute-timespan", url.Values{"value": {strconv.FormatUint(proposalExecuteTimespan, 10)}}, "Could not get can propose setting proposal.execute.time")
}

func (c *Client) CanProposeTNDAOSettingProposalActionTimespan(proposalActionTimespan uint64) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-proposal-action-timespan", url.Values{"value": {strconv.FormatUint(proposalActionTimespan, 10)}}, "Could not get can propose setting proposal.action.time")
}

func (c *Client) CanProposeTNDAOSettingScrubPeriod(scrubPeriod uint64) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-scrub-period", url.Values{"value": {strconv.FormatUint(scrubPeriod, 10)}}, "Could not get can propose setting minipool.scrub.period")
}

func (c *Client) CanProposeTNDAOSettingPromotionScrubPeriod(scrubPeriod uint64) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-promotion-scrub-period", url.Values{"value": {strconv.FormatUint(scrubPeriod, 10)}}, "Could not get can propose setting minipool.promotion.scrub.period")
}

func (c *Client) CanProposeTNDAOSettingScrubPenaltyEnabled(enabled bool) (api.CanProposeTNDAOSettingResponse, error) {
	enabledStr := "false"
	if enabled {
		enabledStr = "true"
	}
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-scrub-penalty-enabled", url.Values{"enabled": {enabledStr}}, "Could not get can propose setting minipool.scrub.penalty.enabled")
}

func (c *Client) CanProposeTNDAOSettingBondReductionWindowStart(windowStart uint64) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-bond-reduction-window-start", url.Values{"value": {strconv.FormatUint(windowStart, 10)}}, "Could not get can propose setting minipool.bond.reduction.window.start")
}

func (c *Client) CanProposeTNDAOSettingBondReductionWindowLength(windowLength uint64) (api.CanProposeTNDAOSettingResponse, error) {
	return c.callAPI[api.CanProposeTNDAOSettingResponse]("GET", "/api/odao/can-propose-bond-reduction-window-length", url.Values{"value": {strconv.FormatUint(windowLength, 10)}}, "Could not get can propose setting minipool.bond.reduction.window.length")
}

// Propose a setting update
func (c *Client) ProposeTNDAOSettingMembersQuorum(quorum float64) (api.ProposeTNDAOSettingMembersQuorumResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingMembersQuorumResponse]("POST", "/api/odao/propose-members-quorum", url.Values{"quorum": {strconv.FormatFloat(quorum, 'f', -1, 64)}}, "Could not propose oracle DAO setting members.quorum")
}

func (c *Client) ProposeTNDAOSettingMembersRplBond(bondAmountWei *big.Int) (api.ProposeTNDAOSettingMembersRplBondResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingMembersRplBondResponse]("POST", "/api/odao/propose-members-rplbond", url.Values{"bondAmountWei": {bondAmountWei.String()}}, "Could not propose oracle DAO setting members.rplbond")
}

func (c *Client) ProposeTNDAOSettingProposalCooldown(proposalCooldownTimespan uint64) (api.ProposeTNDAOSettingProposalCooldownResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingProposalCooldownResponse]("POST", "/api/odao/propose-proposal-cooldown", url.Values{"value": {strconv.FormatUint(proposalCooldownTimespan, 10)}}, "Could not propose oracle DAO setting proposal.cooldown.time")
}

func (c *Client) ProposeTNDAOSettingProposalVoteTimespan(proposalVoteTimespan uint64) (api.ProposeTNDAOSettingProposalVoteTimespanResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingProposalVoteTimespanResponse]("POST", "/api/odao/propose-proposal-vote-timespan", url.Values{"value": {strconv.FormatUint(proposalVoteTimespan, 10)}}, "Could not propose oracle DAO setting proposal.vote.time")
}

func (c *Client) ProposeTNDAOSettingProposalVoteDelayTimespan(proposalDelayTimespan uint64) (api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse]("POST", "/api/odao/propose-proposal-vote-delay-timespan", url.Values{"value": {strconv.FormatUint(proposalDelayTimespan, 10)}}, "Could not propose oracle DAO setting proposal.vote.delay.time")
}

func (c *Client) ProposeTNDAOSettingProposalExecuteTimespan(proposalExecuteTimespan uint64) (api.ProposeTNDAOSettingProposalExecuteTimespanResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingProposalExecuteTimespanResponse]("POST", "/api/odao/propose-proposal-execute-timespan", url.Values{"value": {strconv.FormatUint(proposalExecuteTimespan, 10)}}, "Could not propose oracle DAO setting proposal.execute.time")
}

func (c *Client) ProposeTNDAOSettingProposalActionTimespan(proposalActionTimespan uint64) (api.ProposeTNDAOSettingProposalActionTimespanResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingProposalActionTimespanResponse]("POST", "/api/odao/propose-proposal-action-timespan", url.Values{"value": {strconv.FormatUint(proposalActionTimespan, 10)}}, "Could not propose oracle DAO setting proposal.action.time")
}

func (c *Client) ProposeTNDAOSettingScrubPeriod(scrubPeriod uint64) (api.ProposeTNDAOSettingScrubPeriodResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingScrubPeriodResponse]("POST", "/api/odao/propose-scrub-period", url.Values{"value": {strconv.FormatUint(scrubPeriod, 10)}}, "Could not propose oracle DAO setting minipool.scrub.period")
}

func (c *Client) ProposeTNDAOSettingPromotionScrubPeriod(scrubPeriod uint64) (api.ProposeTNDAOSettingPromotionScrubPeriodResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingPromotionScrubPeriodResponse]("POST", "/api/odao/propose-promotion-scrub-period", url.Values{"value": {strconv.FormatUint(scrubPeriod, 10)}}, "Could not propose oracle DAO setting minipool.promotion.scrub.period")
}

func (c *Client) ProposeTNDAOSettingScrubPenaltyEnabled(enabled bool) (api.ProposeTNDAOSettingScrubPenaltyEnabledResponse, error) {
	enabledStr := "false"
	if enabled {
		enabledStr = "true"
	}
	return c.callAPI[api.ProposeTNDAOSettingScrubPenaltyEnabledResponse]("POST", "/api/odao/propose-scrub-penalty-enabled", url.Values{"enabled": {enabledStr}}, "Could not propose oracle DAO setting minipool.scrub.penalty.enabled")
}

func (c *Client) ProposeTNDAOSettingBondReductionWindowStart(windowStart uint64) (api.ProposeTNDAOSettingBondReductionWindowStartResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingBondReductionWindowStartResponse]("POST", "/api/odao/propose-bond-reduction-window-start", url.Values{"value": {strconv.FormatUint(windowStart, 10)}}, "Could not propose oracle DAO setting minipool.bond.reduction.window.start")
}

func (c *Client) ProposeTNDAOSettingBondReductionWindowLength(windowLength uint64) (api.ProposeTNDAOSettingBondReductionWindowLengthResponse, error) {
	return c.callAPI[api.ProposeTNDAOSettingBondReductionWindowLengthResponse]("POST", "/api/odao/propose-bond-reduction-window-length", url.Values{"value": {strconv.FormatUint(windowLength, 10)}}, "Could not propose oracle DAO setting minipool.bond.reduction.window.length")
}

// Get the member settings
func (c *Client) GetTNDAOMemberSettings() (api.GetTNDAOMemberSettingsResponse, error) {
	response, err := c.callAPI[api.GetTNDAOMemberSettingsResponse]("GET", "/api/odao/get-member-settings", nil, "Could not get oracle DAO member settings")
	if err != nil {
		return response, err
	}
	if response.RPLBond == nil {
		response.RPLBond = big.NewInt(0)
	}
	if response.ChallengeCost == nil {
		response.ChallengeCost = big.NewInt(0)
	}
	return response, nil
}

// Get the proposal settings
func (c *Client) GetTNDAOProposalSettings() (api.GetTNDAOProposalSettingsResponse, error) {
	return c.callAPI[api.GetTNDAOProposalSettingsResponse]("GET", "/api/odao/get-proposal-settings", nil, "Could not get oracle DAO proposal settings")
}

// Get the minipool settings
func (c *Client) GetTNDAOMinipoolSettings() (api.GetTNDAOMinipoolSettingsResponse, error) {
	return c.callAPI[api.GetTNDAOMinipoolSettingsResponse]("GET", "/api/odao/get-minipool-settings", nil, "Could not get oracle DAO minipool settings")
}

// Check whether the node can penalise a megapool
func (c *Client) CanPenaliseMegapool(megapoolAddress common.Address, block *big.Int, amountWei *big.Int) (api.CanPenaliseMegapoolResponse, error) {
	return c.callAPI[api.CanPenaliseMegapoolResponse]("GET", "/api/odao/can-penalise-megapool", url.Values{
		"megapoolAddress": {megapoolAddress.Hex()},
		"block":           {block.String()},
		"amountWei":       {amountWei.String()},
	}, "Could not get can penalise megapool status")
}

// Penalise a megapool
func (c *Client) PenaliseMegapool(megapoolAddress common.Address, block *big.Int, amountWei *big.Int) (api.RepayDebtResponse, error) {
	return c.callAPI[api.RepayDebtResponse]("POST", "/api/odao/penalise-megapool", url.Values{
		"megapoolAddress": {megapoolAddress.Hex()},
		"block":           {block.String()},
		"amountWei":       {amountWei.String()},
	}, "Could not penalise megapool")
}
