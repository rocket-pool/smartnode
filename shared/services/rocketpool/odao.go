package rocketpool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get oracle DAO status
func (c *Client) TNDAOStatus() (api.TNDAOStatusResponse, error) {
	responseBytes, err := c.callAPI("odao status")
	if err != nil {
		return api.TNDAOStatusResponse{}, fmt.Errorf("Could not get oracle DAO status: %w", err)
	}
	var response api.TNDAOStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.TNDAOStatusResponse{}, fmt.Errorf("Could not decode oracle DAO stats response: %w", err)
	}
	if response.Error != "" {
		return api.TNDAOStatusResponse{}, fmt.Errorf("Could not get oracle DAO status: %s", response.Error)
	}
	return response, nil
}

// Get oracle DAO members
func (c *Client) TNDAOMembers() (api.TNDAOMembersResponse, error) {
	responseBytes, err := c.callAPI("odao members")
	if err != nil {
		return api.TNDAOMembersResponse{}, fmt.Errorf("Could not get oracle DAO members: %w", err)
	}
	var response api.TNDAOMembersResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.TNDAOMembersResponse{}, fmt.Errorf("Could not decode oracle DAO members response: %w", err)
	}
	if response.Error != "" {
		return api.TNDAOMembersResponse{}, fmt.Errorf("Could not get oracle DAO members: %s", response.Error)
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
	responseBytes, err := c.callAPI("odao proposals")
	if err != nil {
		return api.TNDAOProposalsResponse{}, fmt.Errorf("Could not get oracle DAO proposals: %w", err)
	}
	var response api.TNDAOProposalsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.TNDAOProposalsResponse{}, fmt.Errorf("Could not decode oracle DAO proposals response: %w", err)
	}
	if response.Error != "" {
		return api.TNDAOProposalsResponse{}, fmt.Errorf("Could not get oracle DAO proposals: %s", response.Error)
	}
	return response, nil
}

// Get a single oracle DAO proposal
func (c *Client) TNDAOProposal(id uint64) (api.TNDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao proposal-details %d", id))
	if err != nil {
		return api.TNDAOProposalResponse{}, fmt.Errorf("Could not get oracle DAO proposal: %w", err)
	}
	var response api.TNDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.TNDAOProposalResponse{}, fmt.Errorf("Could not decode oracle DAO proposal response: %w", err)
	}
	if response.Error != "" {
		return api.TNDAOProposalResponse{}, fmt.Errorf("Could not get oracle DAO proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose inviting a new member
func (c *Client) CanProposeInviteToTNDAO(memberAddress common.Address, memberId, memberUrl string) (api.CanProposeTNDAOInviteResponse, error) {
	responseBytes, err := c.callAPI("odao can-propose-invite", memberAddress.Hex(), memberId, memberUrl)
	if err != nil {
		return api.CanProposeTNDAOInviteResponse{}, fmt.Errorf("Could not get can propose oracle DAO invite status: %w", err)
	}
	var response api.CanProposeTNDAOInviteResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOInviteResponse{}, fmt.Errorf("Could not decode can propose oracle DAO invite response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOInviteResponse{}, fmt.Errorf("Could not get can propose oracle DAO invite status: %s", response.Error)
	}
	return response, nil
}

// Propose inviting a new member
func (c *Client) ProposeInviteToTNDAO(memberAddress common.Address, memberId, memberUrl string) (api.ProposeTNDAOInviteResponse, error) {
	responseBytes, err := c.callAPI("odao propose-invite", memberAddress.Hex(), memberId, memberUrl)
	if err != nil {
		return api.ProposeTNDAOInviteResponse{}, fmt.Errorf("Could not propose oracle DAO invite: %w", err)
	}
	var response api.ProposeTNDAOInviteResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOInviteResponse{}, fmt.Errorf("Could not decode propose oracle DAO invite response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOInviteResponse{}, fmt.Errorf("Could not propose oracle DAO invite: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose leaving the oracle DAO
func (c *Client) CanProposeLeaveTNDAO() (api.CanProposeTNDAOLeaveResponse, error) {
	responseBytes, err := c.callAPI("odao can-propose-leave")
	if err != nil {
		return api.CanProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not get can propose leaving oracle DAO status: %w", err)
	}
	var response api.CanProposeTNDAOLeaveResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not decode can propose leaving oracle DAO response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not get can propose leaving oracle DAO status: %s", response.Error)
	}
	return response, nil
}

// Propose leaving the oracle DAO
func (c *Client) ProposeLeaveTNDAO() (api.ProposeTNDAOLeaveResponse, error) {
	responseBytes, err := c.callAPI("odao propose-leave")
	if err != nil {
		return api.ProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not propose leaving oracle DAO: %w", err)
	}
	var response api.ProposeTNDAOLeaveResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not decode propose leaving oracle DAO response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not propose leaving oracle DAO: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose replacing its position with a new member
func (c *Client) CanProposeReplaceTNDAOMember(memberAddress common.Address, memberId, memberUrl string) (api.CanProposeTNDAOReplaceResponse, error) {
	responseBytes, err := c.callAPI("odao can-propose-replace", memberAddress.Hex(), memberId, memberUrl)
	if err != nil {
		return api.CanProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not get can propose replacing oracle DAO member status: %w", err)
	}
	var response api.CanProposeTNDAOReplaceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not decode can propose replacing oracle DAO member response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not get can propose replacing oracle DAO member status: %s", response.Error)
	}
	return response, nil
}

// Propose replacing the node's position with a new member
func (c *Client) ProposeReplaceTNDAOMember(memberAddress common.Address, memberId, memberUrl string) (api.ProposeTNDAOReplaceResponse, error) {
	responseBytes, err := c.callAPI("odao propose-replace", memberAddress.Hex(), memberId, memberUrl)
	if err != nil {
		return api.ProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not propose replacing oracle DAO member: %w", err)
	}
	var response api.ProposeTNDAOReplaceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not decode propose replacing oracle DAO member response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not propose replacing oracle DAO member: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose kicking a member
func (c *Client) CanProposeKickFromTNDAO(memberAddress common.Address, fineAmountWei *big.Int) (api.CanProposeTNDAOKickResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-kick %s %s", memberAddress.Hex(), fineAmountWei.String()))
	if err != nil {
		return api.CanProposeTNDAOKickResponse{}, fmt.Errorf("Could not get can propose kicking oracle DAO member status: %w", err)
	}
	var response api.CanProposeTNDAOKickResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOKickResponse{}, fmt.Errorf("Could not decode can propose kicking oracle DAO member response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOKickResponse{}, fmt.Errorf("Could not get can propose kicking oracle DAO member status: %s", response.Error)
	}
	return response, nil
}

// Propose kicking a member
func (c *Client) ProposeKickFromTNDAO(memberAddress common.Address, fineAmountWei *big.Int) (api.ProposeTNDAOKickResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-kick %s %s", memberAddress.Hex(), fineAmountWei.String()))
	if err != nil {
		return api.ProposeTNDAOKickResponse{}, fmt.Errorf("Could not propose kicking oracle DAO member: %w", err)
	}
	var response api.ProposeTNDAOKickResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOKickResponse{}, fmt.Errorf("Could not decode propose kicking oracle DAO member response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOKickResponse{}, fmt.Errorf("Could not propose kicking oracle DAO member: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can cancel a proposal
func (c *Client) CanCancelTNDAOProposal(proposalId uint64) (api.CanCancelTNDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-cancel-proposal %d", proposalId))
	if err != nil {
		return api.CanCancelTNDAOProposalResponse{}, fmt.Errorf("Could not get can cancel oracle DAO proposal status: %w", err)
	}
	var response api.CanCancelTNDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanCancelTNDAOProposalResponse{}, fmt.Errorf("Could not decode can cancel oracle DAO proposal response: %w", err)
	}
	if response.Error != "" {
		return api.CanCancelTNDAOProposalResponse{}, fmt.Errorf("Could not get can cancel oracle DAO proposal status: %s", response.Error)
	}
	return response, nil
}

// Cancel a proposal made by the node
func (c *Client) CancelTNDAOProposal(proposalId uint64) (api.CancelTNDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao cancel-proposal %d", proposalId))
	if err != nil {
		return api.CancelTNDAOProposalResponse{}, fmt.Errorf("Could not cancel oracle DAO proposal: %w", err)
	}
	var response api.CancelTNDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CancelTNDAOProposalResponse{}, fmt.Errorf("Could not decode cancel oracle DAO proposal response: %w", err)
	}
	if response.Error != "" {
		return api.CancelTNDAOProposalResponse{}, fmt.Errorf("Could not cancel oracle DAO proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can vote on a proposal
func (c *Client) CanVoteOnTNDAOProposal(proposalId uint64) (api.CanVoteOnTNDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-vote-proposal %d", proposalId))
	if err != nil {
		return api.CanVoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not get can vote on oracle DAO proposal status: %w", err)
	}
	var response api.CanVoteOnTNDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanVoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not decode can vote on oracle DAO proposal response: %w", err)
	}
	if response.Error != "" {
		return api.CanVoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not get can vote on oracle DAO proposal status: %s", response.Error)
	}
	return response, nil
}

// Vote on a proposal
func (c *Client) VoteOnTNDAOProposal(proposalId uint64, support bool) (api.VoteOnTNDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao vote-proposal %d %t", proposalId, support))
	if err != nil {
		return api.VoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not vote on oracle DAO proposal: %w", err)
	}
	var response api.VoteOnTNDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.VoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not decode vote on oracle DAO proposal response: %w", err)
	}
	if response.Error != "" {
		return api.VoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not vote on oracle DAO proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can execute a proposal
func (c *Client) CanExecuteTNDAOProposal(proposalId uint64) (api.CanExecuteTNDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-execute-proposal %d", proposalId))
	if err != nil {
		return api.CanExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not get can execute oracle DAO proposal status: %w", err)
	}
	var response api.CanExecuteTNDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not decode can execute oracle DAO proposal response: %w", err)
	}
	if response.Error != "" {
		return api.CanExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not get can execute oracle DAO proposal status: %s", response.Error)
	}
	return response, nil
}

// Execute a proposal
func (c *Client) ExecuteTNDAOProposal(proposalId uint64) (api.ExecuteTNDAOProposalResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao execute-proposal %d", proposalId))
	if err != nil {
		return api.ExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not execute oracle DAO proposal: %w", err)
	}
	var response api.ExecuteTNDAOProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not decode execute oracle DAO proposal response: %w", err)
	}
	if response.Error != "" {
		return api.ExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not execute oracle DAO proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can join the oracle DAO
func (c *Client) CanJoinTNDAO() (api.CanJoinTNDAOResponse, error) {
	responseBytes, err := c.callAPI("odao can-join")
	if err != nil {
		return api.CanJoinTNDAOResponse{}, fmt.Errorf("Could not get can join oracle DAO status: %w", err)
	}
	var response api.CanJoinTNDAOResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanJoinTNDAOResponse{}, fmt.Errorf("Could not decode can join oracle DAO response: %w", err)
	}
	if response.Error != "" {
		return api.CanJoinTNDAOResponse{}, fmt.Errorf("Could not get can join oracle DAO status: %s", response.Error)
	}
	return response, nil
}

// Join the oracle DAO (requires an executed invite proposal)
func (c *Client) ApproveRPLToJoinTNDAO() (api.JoinTNDAOApproveResponse, error) {
	responseBytes, err := c.callAPI("odao join-approve-rpl")
	if err != nil {
		return api.JoinTNDAOApproveResponse{}, fmt.Errorf("Could not approve RPL for joining oracle DAO: %w", err)
	}
	var response api.JoinTNDAOApproveResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.JoinTNDAOApproveResponse{}, fmt.Errorf("Could not decode approve RPL for joining oracle DAO response: %w", err)
	}
	if response.Error != "" {
		return api.JoinTNDAOApproveResponse{}, fmt.Errorf("Could not approve RPL for joining oracle DAO: %s", response.Error)
	}
	return response, nil
}

// Join the oracle DAO (requires an executed invite proposal)
func (c *Client) JoinTNDAO(approvalTxHash common.Hash) (api.JoinTNDAOJoinResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao join %s", approvalTxHash.String()))
	if err != nil {
		return api.JoinTNDAOJoinResponse{}, fmt.Errorf("Could not join oracle DAO: %w", err)
	}
	var response api.JoinTNDAOJoinResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.JoinTNDAOJoinResponse{}, fmt.Errorf("Could not decode join oracle DAO response: %w", err)
	}
	if response.Error != "" {
		return api.JoinTNDAOJoinResponse{}, fmt.Errorf("Could not join oracle DAO: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can leave the oracle DAO
func (c *Client) CanLeaveTNDAO() (api.CanLeaveTNDAOResponse, error) {
	responseBytes, err := c.callAPI("odao can-leave")
	if err != nil {
		return api.CanLeaveTNDAOResponse{}, fmt.Errorf("Could not get can leave oracle DAO status: %w", err)
	}
	var response api.CanLeaveTNDAOResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanLeaveTNDAOResponse{}, fmt.Errorf("Could not decode can leave oracle DAO response: %w", err)
	}
	if response.Error != "" {
		return api.CanLeaveTNDAOResponse{}, fmt.Errorf("Could not get can leave oracle DAO status: %s", response.Error)
	}
	return response, nil
}

// Leave the oracle DAO (requires an executed leave proposal)
func (c *Client) LeaveTNDAO(bondRefundAddress common.Address) (api.LeaveTNDAOResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao leave %s", bondRefundAddress.Hex()))
	if err != nil {
		return api.LeaveTNDAOResponse{}, fmt.Errorf("Could not leave oracle DAO: %w", err)
	}
	var response api.LeaveTNDAOResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.LeaveTNDAOResponse{}, fmt.Errorf("Could not decode leave oracle DAO response: %w", err)
	}
	if response.Error != "" {
		return api.LeaveTNDAOResponse{}, fmt.Errorf("Could not leave oracle DAO: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can replace its position in the oracle DAO
func (c *Client) CanReplaceTNDAOMember() (api.CanReplaceTNDAOPositionResponse, error) {
	responseBytes, err := c.callAPI("odao can-replace")
	if err != nil {
		return api.CanReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not get can replace oracle DAO member status: %w", err)
	}
	var response api.CanReplaceTNDAOPositionResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not decode can replace oracle DAO member response: %w", err)
	}
	if response.Error != "" {
		return api.CanReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not get can replace oracle DAO member status: %s", response.Error)
	}
	return response, nil
}

// Replace the node's position in the oracle DAO (requires an executed replace proposal)
func (c *Client) ReplaceTNDAOMember() (api.ReplaceTNDAOPositionResponse, error) {
	responseBytes, err := c.callAPI("odao replace")
	if err != nil {
		return api.ReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not replace oracle DAO member: %w", err)
	}
	var response api.ReplaceTNDAOPositionResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not decode replace oracle DAO member response: %w", err)
	}
	if response.Error != "" {
		return api.ReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not replace oracle DAO member: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose a setting update
func (c *Client) CanProposeTNDAOSetting() (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI("odao can-propose-setting")
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting status: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting status: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingMembersQuorum(quorum float64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-members-quorum %f", quorum))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting members.quorum: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting members.quorum response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting members.quorum: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingMembersRplBond(bondAmountWei *big.Int) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-members-rplbond %s", bondAmountWei.String()))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting members.rplbond: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting members.rplbond response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting members.rplbond: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingMinipoolUnbondedMax(unbondedMinipoolMax uint64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-members-minipool-unbonded-max %d", unbondedMinipoolMax))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting members.minipool.unbonded.max: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting members.minipool.unbonded.max response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting members.minipool.unbonded.max: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingProposalCooldown(proposalCooldownTimespan uint64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-proposal-cooldown %d", proposalCooldownTimespan))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting proposal.cooldown.time: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting proposal.cooldown.time response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting proposal.cooldown.time: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingProposalVoteTimespan(proposalVoteTimespan uint64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-proposal-vote-timespan %d", proposalVoteTimespan))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting proposal.vote.time: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting proposal.vote.time response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting proposal.vote.time: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingProposalVoteDelayTimespan(proposalDelayTimespan uint64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-proposal-vote-delay-timespan %d", proposalDelayTimespan))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting proposal.vote.delay.time: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting proposal.vote.delay.time response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting proposal.vote.delay.time: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingProposalExecuteTimespan(proposalExecuteTimespan uint64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-proposal-execute-timespan %d", proposalExecuteTimespan))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting proposal.execute.time: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting proposal.execute.time response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting proposal.execute.time: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingProposalActionTimespan(proposalActionTimespan uint64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-proposal-action-timespan %d", proposalActionTimespan))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting proposal.action.time: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting proposal.action.time response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting proposal.action.time: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingScrubPeriod(scrubPeriod uint64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-scrub-period %d", scrubPeriod))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting minipool.scrub.period: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting minipool.scrub.period response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting minipool.scrub.period: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingPromotionScrubPeriod(scrubPeriod uint64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-promotion-scrub-period %d", scrubPeriod))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting minipool.promotion.scrub.period: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting minipool.promotion.scrub.period response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting minipool.promotion.scrub.period: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingScrubPenaltyEnabled(enabled bool) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-scrub-penalty-enabled %t", enabled))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting minipool.scrub.penalty.enabled: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting minipool.scrub.penalty.enabled response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting minipool.scrub.penalty.enabled: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingBondReductionWindowStart(windowStart uint64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-bond-reduction-window-start %d", windowStart))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting minipool.bond.reduction.window.start: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting minipool.bond.reduction.window.start response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting minipool.bond.reduction.window.start: %s", response.Error)
	}
	return response, nil
}
func (c *Client) CanProposeTNDAOSettingBondReductionWindowLength(windowLength uint64) (api.CanProposeTNDAOSettingResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao can-propose-bond-reduction-window-length %d", windowLength))
	if err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting minipool.bond.reduction.window.length: %w", err)
	}
	var response api.CanProposeTNDAOSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not decode can propose setting minipool.bond.reduction.window.length response: %w", err)
	}
	if response.Error != "" {
		return api.CanProposeTNDAOSettingResponse{}, fmt.Errorf("Could not get can propose setting minipool.bond.reduction.window.length: %s", response.Error)
	}
	return response, nil
}

// Propose a setting update
func (c *Client) ProposeTNDAOSettingMembersQuorum(quorum float64) (api.ProposeTNDAOSettingMembersQuorumResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-members-quorum %f", quorum))
	if err != nil {
		return api.ProposeTNDAOSettingMembersQuorumResponse{}, fmt.Errorf("Could not propose oracle DAO setting members.quorum: %w", err)
	}
	var response api.ProposeTNDAOSettingMembersQuorumResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingMembersQuorumResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting members.quorum response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingMembersQuorumResponse{}, fmt.Errorf("Could not propose oracle DAO setting members.quorum: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingMembersRplBond(bondAmountWei *big.Int) (api.ProposeTNDAOSettingMembersRplBondResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-members-rplbond %s", bondAmountWei.String()))
	if err != nil {
		return api.ProposeTNDAOSettingMembersRplBondResponse{}, fmt.Errorf("Could not propose oracle DAO setting members.rplbond: %w", err)
	}
	var response api.ProposeTNDAOSettingMembersRplBondResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingMembersRplBondResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting members.rplbond response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingMembersRplBondResponse{}, fmt.Errorf("Could not propose oracle DAO setting members.rplbond: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingMinipoolUnbondedMax(unbondedMinipoolMax uint64) (api.ProposeTNDAOSettingMinipoolUnbondedMaxResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-members-minipool-unbonded-max %d", unbondedMinipoolMax))
	if err != nil {
		return api.ProposeTNDAOSettingMinipoolUnbondedMaxResponse{}, fmt.Errorf("Could not propose oracle DAO setting members.minipool.unbonded.max: %w", err)
	}
	var response api.ProposeTNDAOSettingMinipoolUnbondedMaxResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingMinipoolUnbondedMaxResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting members.minipool.unbonded.max response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingMinipoolUnbondedMaxResponse{}, fmt.Errorf("Could not propose oracle DAO setting members.minipool.unbonded.max: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingProposalCooldown(proposalCooldownTimespan uint64) (api.ProposeTNDAOSettingProposalCooldownResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-proposal-cooldown %d", proposalCooldownTimespan))
	if err != nil {
		return api.ProposeTNDAOSettingProposalCooldownResponse{}, fmt.Errorf("Could not propose oracle DAO setting proposal.cooldown.time: %w", err)
	}
	var response api.ProposeTNDAOSettingProposalCooldownResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingProposalCooldownResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting proposal.cooldown.time response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingProposalCooldownResponse{}, fmt.Errorf("Could not propose oracle DAO setting proposal.cooldown.time: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingProposalVoteTimespan(proposalVoteTimespan uint64) (api.ProposeTNDAOSettingProposalVoteTimespanResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-proposal-vote-timespan %d", proposalVoteTimespan))
	if err != nil {
		return api.ProposeTNDAOSettingProposalVoteTimespanResponse{}, fmt.Errorf("Could not propose oracle DAO setting proposal.vote.time: %w", err)
	}
	var response api.ProposeTNDAOSettingProposalVoteTimespanResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingProposalVoteTimespanResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting proposal.vote.time response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingProposalVoteTimespanResponse{}, fmt.Errorf("Could not propose oracle DAO setting proposal.vote.time: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingProposalVoteDelayTimespan(proposalDelayTimespan uint64) (api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-proposal-vote-delay-timespan %d", proposalDelayTimespan))
	if err != nil {
		return api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse{}, fmt.Errorf("Could not propose oracle DAO setting proposal.vote.delay.time: %w", err)
	}
	var response api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting proposal.vote.delay.time response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingProposalVoteDelayTimespanResponse{}, fmt.Errorf("Could not propose oracle DAO setting proposal.vote.delay.time: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingProposalExecuteTimespan(proposalExecuteTimespan uint64) (api.ProposeTNDAOSettingProposalExecuteTimespanResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-proposal-execute-timespan %d", proposalExecuteTimespan))
	if err != nil {
		return api.ProposeTNDAOSettingProposalExecuteTimespanResponse{}, fmt.Errorf("Could not propose oracle DAO setting proposal.execute.time: %w", err)
	}
	var response api.ProposeTNDAOSettingProposalExecuteTimespanResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingProposalExecuteTimespanResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting proposal.execute.time response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingProposalExecuteTimespanResponse{}, fmt.Errorf("Could not propose oracle DAO setting proposal.execute.time: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingProposalActionTimespan(proposalActionTimespan uint64) (api.ProposeTNDAOSettingProposalActionTimespanResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-proposal-action-timespan %d", proposalActionTimespan))
	if err != nil {
		return api.ProposeTNDAOSettingProposalActionTimespanResponse{}, fmt.Errorf("Could not propose oracle DAO setting proposal.action.time: %w", err)
	}
	var response api.ProposeTNDAOSettingProposalActionTimespanResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingProposalActionTimespanResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting proposal.action.time response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingProposalActionTimespanResponse{}, fmt.Errorf("Could not propose oracle DAO setting proposal.action.time: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingScrubPeriod(scrubPeriod uint64) (api.ProposeTNDAOSettingScrubPeriodResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-scrub-period %d", scrubPeriod))
	if err != nil {
		return api.ProposeTNDAOSettingScrubPeriodResponse{}, fmt.Errorf("Could not propose oracle DAO setting minipool.scrub.period: %w", err)
	}
	var response api.ProposeTNDAOSettingScrubPeriodResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingScrubPeriodResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting minipool.scrub.period response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingScrubPeriodResponse{}, fmt.Errorf("Could not propose oracle DAO setting minipool.scrub.period: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingPromotionScrubPeriod(scrubPeriod uint64) (api.ProposeTNDAOSettingPromotionScrubPeriodResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-promotion-scrub-period %d", scrubPeriod))
	if err != nil {
		return api.ProposeTNDAOSettingPromotionScrubPeriodResponse{}, fmt.Errorf("Could not propose oracle DAO setting minipool.promotion.scrub.period: %w", err)
	}
	var response api.ProposeTNDAOSettingPromotionScrubPeriodResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingPromotionScrubPeriodResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting minipool.promotion.scrub.period response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingPromotionScrubPeriodResponse{}, fmt.Errorf("Could not propose oracle DAO setting minipool.promotion.scrub.period: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingScrubPenaltyEnabled(enabled bool) (api.ProposeTNDAOSettingScrubPenaltyEnabledResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-scrub-penalty-enabled %t", enabled))
	if err != nil {
		return api.ProposeTNDAOSettingScrubPenaltyEnabledResponse{}, fmt.Errorf("Could not propose oracle DAO setting minipool.scrub.penalty.enabled: %w", err)
	}
	var response api.ProposeTNDAOSettingScrubPenaltyEnabledResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingScrubPenaltyEnabledResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting minipool.scrub.penalty.enabled response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingScrubPenaltyEnabledResponse{}, fmt.Errorf("Could not propose oracle DAO setting minipool.scrub.penalty.enabled: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingBondReductionWindowStart(windowStart uint64) (api.ProposeTNDAOSettingBondReductionWindowStartResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-bond-reduction-window-start %d", windowStart))
	if err != nil {
		return api.ProposeTNDAOSettingBondReductionWindowStartResponse{}, fmt.Errorf("Could not propose oracle DAO setting minipool.bond.reduction.window.start: %w", err)
	}
	var response api.ProposeTNDAOSettingBondReductionWindowStartResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingBondReductionWindowStartResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting minipool.bond.reduction.window.start response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingBondReductionWindowStartResponse{}, fmt.Errorf("Could not propose oracle DAO setting minipool.bond.reduction.window.start: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ProposeTNDAOSettingBondReductionWindowLength(windowLength uint64) (api.ProposeTNDAOSettingBondReductionWindowLengthResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao propose-bond-reduction-window-length %d", windowLength))
	if err != nil {
		return api.ProposeTNDAOSettingBondReductionWindowLengthResponse{}, fmt.Errorf("Could not propose oracle DAO setting minipool.bond.reduction.window.length: %w", err)
	}
	var response api.ProposeTNDAOSettingBondReductionWindowLengthResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ProposeTNDAOSettingBondReductionWindowLengthResponse{}, fmt.Errorf("Could not decode propose oracle DAO setting minipool.bond.reduction.window.length response: %w", err)
	}
	if response.Error != "" {
		return api.ProposeTNDAOSettingBondReductionWindowLengthResponse{}, fmt.Errorf("Could not propose oracle DAO setting minipool.bond.reduction.window.length: %s", response.Error)
	}
	return response, nil
}

// Get the member settings
func (c *Client) GetTNDAOMemberSettings() (api.GetTNDAOMemberSettingsResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao get-member-settings"))
	if err != nil {
		return api.GetTNDAOMemberSettingsResponse{}, fmt.Errorf("Could not get oracle DAO member settings: %w", err)
	}
	var response api.GetTNDAOMemberSettingsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetTNDAOMemberSettingsResponse{}, fmt.Errorf("Could not decode oracle DAO member settings response: %w", err)
	}
	if response.Error != "" {
		return api.GetTNDAOMemberSettingsResponse{}, fmt.Errorf("Could not get oracle DAO member settings: %s", response.Error)
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
	responseBytes, err := c.callAPI(fmt.Sprintf("odao get-proposal-settings"))
	if err != nil {
		return api.GetTNDAOProposalSettingsResponse{}, fmt.Errorf("Could not get oracle DAO proposal settings: %w", err)
	}
	var response api.GetTNDAOProposalSettingsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetTNDAOProposalSettingsResponse{}, fmt.Errorf("Could not decode oracle DAO proposal settings response: %w", err)
	}
	if response.Error != "" {
		return api.GetTNDAOProposalSettingsResponse{}, fmt.Errorf("Could not get oracle DAO proposal settings: %s", response.Error)
	}
	return response, nil
}

// Get the proposal settings
func (c *Client) GetTNDAOMinipoolSettings() (api.GetTNDAOMinipoolSettingsResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("odao get-minipool-settings"))
	if err != nil {
		return api.GetTNDAOMinipoolSettingsResponse{}, fmt.Errorf("Could not get oracle DAO minipool settings: %w", err)
	}
	var response api.GetTNDAOMinipoolSettingsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetTNDAOMinipoolSettingsResponse{}, fmt.Errorf("Could not decode oracle DAO minipool settings response: %w", err)
	}
	if response.Error != "" {
		return api.GetTNDAOMinipoolSettingsResponse{}, fmt.Errorf("Could not get oracle DAO minipool settings: %s", response.Error)
	}
	return response, nil
}
