package rocketpool

import (
    "encoding/json"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Get trusted node DAO status
func (c *Client) TNDAOStatus() (api.TNDAOStatusResponse, error) {
    responseBytes, err := c.callAPI("tndao status")
    if err != nil {
        return api.TNDAOStatusResponse{}, fmt.Errorf("Could not get trusted node DAO status: %w", err)
    }
    var response api.TNDAOStatusResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.TNDAOStatusResponse{}, fmt.Errorf("Could not decode trusted node DAO stats response: %w", err)
    }
    if response.Error != "" {
        return api.TNDAOStatusResponse{}, fmt.Errorf("Could not get trusted node DAO status: %s", response.Error)
    }
    return response, nil
}


// Get trusted node DAO members
func (c *Client) TNDAOMembers() (api.TNDAOMembersResponse, error) {
    responseBytes, err := c.callAPI("tndao members")
    if err != nil {
        return api.TNDAOMembersResponse{}, fmt.Errorf("Could not get trusted node DAO members: %w", err)
    }
    var response api.TNDAOMembersResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.TNDAOMembersResponse{}, fmt.Errorf("Could not decode trusted node DAO members response: %w", err)
    }
    if response.Error != "" {
        return api.TNDAOMembersResponse{}, fmt.Errorf("Could not get trusted node DAO members: %s", response.Error)
    }
    return response, nil
}


// Get trusted node DAO proposals
func (c *Client) TNDAOProposals() (api.TNDAOProposalsResponse, error) {
    responseBytes, err := c.callAPI("tndao proposals")
    if err != nil {
        return api.TNDAOProposalsResponse{}, fmt.Errorf("Could not get trusted node DAO proposals: %w", err)
    }
    var response api.TNDAOProposalsResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.TNDAOProposalsResponse{}, fmt.Errorf("Could not decode trusted node DAO proposals response: %w", err)
    }
    if response.Error != "" {
        return api.TNDAOProposalsResponse{}, fmt.Errorf("Could not get trusted node DAO proposals: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can propose inviting a new member
func (c *Client) CanProposeInviteToTNDAO(memberAddress common.Address) (api.CanProposeTNDAOInviteResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao can-propose-invite %s", memberAddress.Hex()))
    if err != nil {
        return api.CanProposeTNDAOInviteResponse{}, fmt.Errorf("Could not get can propose trusted node DAO invite status: %w", err)
    }
    var response api.CanProposeTNDAOInviteResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanProposeTNDAOInviteResponse{}, fmt.Errorf("Could not decode can propose trusted node DAO invite response: %w", err)
    }
    if response.Error != "" {
        return api.CanProposeTNDAOInviteResponse{}, fmt.Errorf("Could not get can propose trusted node DAO invite status: %s", response.Error)
    }
    return response, nil
}


// Propose inviting a new member
func (c *Client) ProposeInviteToTNDAO(memberAddress common.Address, memberId, memberEmail string) (api.ProposeTNDAOInviteResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao propose-invite %s \"%s\" \"%s\"", memberAddress.Hex(), memberId, memberEmail))
    if err != nil {
        return api.ProposeTNDAOInviteResponse{}, fmt.Errorf("Could not propose trusted node DAO invite: %w", err)
    }
    var response api.ProposeTNDAOInviteResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.ProposeTNDAOInviteResponse{}, fmt.Errorf("Could not decode propose trusted node DAO invite response: %w", err)
    }
    if response.Error != "" {
        return api.ProposeTNDAOInviteResponse{}, fmt.Errorf("Could not propose trusted node DAO invite: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can propose leaving the trusted node DAO
func (c *Client) CanProposeLeaveTNDAO() (api.CanProposeTNDAOLeaveResponse, error) {
    responseBytes, err := c.callAPI("tndao can-propose-leave")
    if err != nil {
        return api.CanProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not get can propose leaving trusted node DAO status: %w", err)
    }
    var response api.CanProposeTNDAOLeaveResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not decode can propose leaving trusted node DAO response: %w", err)
    }
    if response.Error != "" {
        return api.CanProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not get can propose leaving trusted node DAO status: %s", response.Error)
    }
    return response, nil
}


// Propose leaving the trusted node DAO
func (c *Client) ProposeLeaveTNDAO() (api.ProposeTNDAOLeaveResponse, error) {
    responseBytes, err := c.callAPI("tndao propose-leave")
    if err != nil {
        return api.ProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not propose leaving trusted node DAO: %w", err)
    }
    var response api.ProposeTNDAOLeaveResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.ProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not decode propose leaving trusted node DAO response: %w", err)
    }
    if response.Error != "" {
        return api.ProposeTNDAOLeaveResponse{}, fmt.Errorf("Could not propose leaving trusted node DAO: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can propose replacing its position with a new member
func (c *Client) CanProposeReplaceTNDAOMember(memberAddress common.Address) (api.CanProposeTNDAOReplaceResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao can-propose-replace %s", memberAddress.Hex()))
    if err != nil {
        return api.CanProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not get can propose replacing trusted node DAO member status: %w", err)
    }
    var response api.CanProposeTNDAOReplaceResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not decode can propose replacing trusted node DAO member response: %w", err)
    }
    if response.Error != "" {
        return api.CanProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not get can propose replacing trusted node DAO member status: %s", response.Error)
    }
    return response, nil
}


// Propose replacing the node's position with a new member
func (c *Client) ProposeReplaceTNDAOMember(memberAddress common.Address, memberId, memberEmail string) (api.ProposeTNDAOReplaceResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao propose-replace %s \"%s\" \"%s\"", memberAddress.Hex(), memberId, memberEmail))
    if err != nil {
        return api.ProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not propose replacing trusted node DAO member: %w", err)
    }
    var response api.ProposeTNDAOReplaceResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.ProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not decode propose replacing trusted node DAO member response: %w", err)
    }
    if response.Error != "" {
        return api.ProposeTNDAOReplaceResponse{}, fmt.Errorf("Could not propose replacing trusted node DAO member: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can propose kicking a member
func (c *Client) CanProposeKickFromTNDAO(memberAddress common.Address, fineAmountWei *big.Int) (api.CanProposeTNDAOKickResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao can-propose-kick %s %s", memberAddress.Hex(), fineAmountWei.String()))
    if err != nil {
        return api.CanProposeTNDAOKickResponse{}, fmt.Errorf("Could not get can propose kicking trusted node DAO member status: %w", err)
    }
    var response api.CanProposeTNDAOKickResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanProposeTNDAOKickResponse{}, fmt.Errorf("Could not decode can propose kicking trusted node DAO member response: %w", err)
    }
    if response.Error != "" {
        return api.CanProposeTNDAOKickResponse{}, fmt.Errorf("Could not get can propose kicking trusted node DAO member status: %s", response.Error)
    }
    return response, nil
}


// Propose kicking a member
func (c *Client) ProposeKickFromTNDAO(memberAddress common.Address, fineAmountWei *big.Int) (api.ProposeTNDAOKickResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao propose-kick %s %s", memberAddress.Hex(), fineAmountWei.String()))
    if err != nil {
        return api.ProposeTNDAOKickResponse{}, fmt.Errorf("Could not propose kicking trusted node DAO member: %w", err)
    }
    var response api.ProposeTNDAOKickResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.ProposeTNDAOKickResponse{}, fmt.Errorf("Could not decode propose kicking trusted node DAO member response: %w", err)
    }
    if response.Error != "" {
        return api.ProposeTNDAOKickResponse{}, fmt.Errorf("Could not propose kicking trusted node DAO member: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can cancel a proposal
func (c *Client) CanCancelTNDAOProposal(proposalId uint64) (api.CanCancelTNDAOProposalResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao can-cancel-proposal %d", proposalId))
    if err != nil {
        return api.CanCancelTNDAOProposalResponse{}, fmt.Errorf("Could not get can cancel trusted node DAO proposal status: %w", err)
    }
    var response api.CanCancelTNDAOProposalResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanCancelTNDAOProposalResponse{}, fmt.Errorf("Could not decode can cancel trusted node DAO proposal response: %w", err)
    }
    if response.Error != "" {
        return api.CanCancelTNDAOProposalResponse{}, fmt.Errorf("Could not get can cancel trusted node DAO proposal status: %s", response.Error)
    }
    return response, nil
}


// Cancel a proposal made by the node
func (c *Client) CancelTNDAOProposal(proposalId uint64) (api.CancelTNDAOProposalResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao cancel-proposal %d", proposalId))
    if err != nil {
        return api.CancelTNDAOProposalResponse{}, fmt.Errorf("Could not cancel trusted node DAO proposal: %w", err)
    }
    var response api.CancelTNDAOProposalResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CancelTNDAOProposalResponse{}, fmt.Errorf("Could not decode cancel trusted node DAO proposal response: %w", err)
    }
    if response.Error != "" {
        return api.CancelTNDAOProposalResponse{}, fmt.Errorf("Could not cancel trusted node DAO proposal: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can vote on a proposal
func (c *Client) CanVoteOnTNDAOProposal(proposalId uint64) (api.CanVoteOnTNDAOProposalResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao can-vote-proposal %d", proposalId))
    if err != nil {
        return api.CanVoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not get can vote on trusted node DAO proposal status: %w", err)
    }
    var response api.CanVoteOnTNDAOProposalResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanVoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not decode can vote on trusted node DAO proposal response: %w", err)
    }
    if response.Error != "" {
        return api.CanVoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not get can vote on trusted node DAO proposal status: %s", response.Error)
    }
    return response, nil
}


// Vote on a proposal
func (c *Client) VoteOnTNDAOProposal(proposalId uint64, support bool) (api.VoteOnTNDAOProposalResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao vote-proposal %d %t", proposalId, support))
    if err != nil {
        return api.VoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not vote on trusted node DAO proposal: %w", err)
    }
    var response api.VoteOnTNDAOProposalResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.VoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not decode vote on trusted node DAO proposal response: %w", err)
    }
    if response.Error != "" {
        return api.VoteOnTNDAOProposalResponse{}, fmt.Errorf("Could not vote on trusted node DAO proposal: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can execute a proposal
func (c *Client) CanExecuteTNDAOProposal(proposalId uint64) (api.CanExecuteTNDAOProposalResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao can-execute-proposal %d", proposalId))
    if err != nil {
        return api.CanExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not get can execute trusted node DAO proposal status: %w", err)
    }
    var response api.CanExecuteTNDAOProposalResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not decode can execute trusted node DAO proposal response: %w", err)
    }
    if response.Error != "" {
        return api.CanExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not get can execute trusted node DAO proposal status: %s", response.Error)
    }
    return response, nil
}


// Execute a proposal
func (c *Client) ExecuteTNDAOProposal(proposalId uint64) (api.ExecuteTNDAOProposalResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao execute-proposal %d", proposalId))
    if err != nil {
        return api.ExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not execute trusted node DAO proposal: %w", err)
    }
    var response api.ExecuteTNDAOProposalResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.ExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not decode execute trusted node DAO proposal response: %w", err)
    }
    if response.Error != "" {
        return api.ExecuteTNDAOProposalResponse{}, fmt.Errorf("Could not execute trusted node DAO proposal: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can join the trusted node DAO
func (c *Client) CanJoinTNDAO() (api.CanJoinTNDAOResponse, error) {
    responseBytes, err := c.callAPI("tndao can-join")
    if err != nil {
        return api.CanJoinTNDAOResponse{}, fmt.Errorf("Could not get can join trusted node DAO status: %w", err)
    }
    var response api.CanJoinTNDAOResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanJoinTNDAOResponse{}, fmt.Errorf("Could not decode can join trusted node DAO response: %w", err)
    }
    if response.Error != "" {
        return api.CanJoinTNDAOResponse{}, fmt.Errorf("Could not get can join trusted node DAO status: %s", response.Error)
    }
    return response, nil
}


// Join the trusted node DAO (requires an executed invite proposal)
func (c *Client) JoinTNDAO() (api.JoinTNDAOResponse, error) {
    responseBytes, err := c.callAPI("tndao join")
    if err != nil {
        return api.JoinTNDAOResponse{}, fmt.Errorf("Could not join trusted node DAO: %w", err)
    }
    var response api.JoinTNDAOResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.JoinTNDAOResponse{}, fmt.Errorf("Could not decode join trusted node DAO response: %w", err)
    }
    if response.Error != "" {
        return api.JoinTNDAOResponse{}, fmt.Errorf("Could not join trusted node DAO: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can leave the trusted node DAO
func (c *Client) CanLeaveTNDAO() (api.CanLeaveTNDAOResponse, error) {
    responseBytes, err := c.callAPI("tndao can-leave")
    if err != nil {
        return api.CanLeaveTNDAOResponse{}, fmt.Errorf("Could not get can leave trusted node DAO status: %w", err)
    }
    var response api.CanLeaveTNDAOResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanLeaveTNDAOResponse{}, fmt.Errorf("Could not decode can leave trusted node DAO response: %w", err)
    }
    if response.Error != "" {
        return api.CanLeaveTNDAOResponse{}, fmt.Errorf("Could not get can leave trusted node DAO status: %s", response.Error)
    }
    return response, nil
}


// Leave the trusted node DAO (requires an executed leave proposal)
func (c *Client) LeaveTNDAO(bondRefundAddress common.Address) (api.LeaveTNDAOResponse, error) {
    responseBytes, err := c.callAPI(fmt.Sprintf("tndao leave %s", bondRefundAddress.Hex()))
    if err != nil {
        return api.LeaveTNDAOResponse{}, fmt.Errorf("Could not leave trusted node DAO: %w", err)
    }
    var response api.LeaveTNDAOResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.LeaveTNDAOResponse{}, fmt.Errorf("Could not decode leave trusted node DAO response: %w", err)
    }
    if response.Error != "" {
        return api.LeaveTNDAOResponse{}, fmt.Errorf("Could not leave trusted node DAO: %s", response.Error)
    }
    return response, nil
}


// Check whether the node can replace its position in the trusted node DAO
func (c *Client) CanReplaceTNDAOMember() (api.CanReplaceTNDAOPositionResponse, error) {
    responseBytes, err := c.callAPI("tndao can-replace")
    if err != nil {
        return api.CanReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not get can replace trusted node DAO member status: %w", err)
    }
    var response api.CanReplaceTNDAOPositionResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.CanReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not decode can replace trusted node DAO member response: %w", err)
    }
    if response.Error != "" {
        return api.CanReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not get can replace trusted node DAO member status: %s", response.Error)
    }
    return response, nil
}


// Replace the node's position in the trusted node DAO (requires an executed replace proposal)
func (c *Client) ReplaceTNDAOMember() (api.ReplaceTNDAOPositionResponse, error) {
    responseBytes, err := c.callAPI("tndao replace")
    if err != nil {
        return api.ReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not replace trusted node DAO member: %w", err)
    }
    var response api.ReplaceTNDAOPositionResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        return api.ReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not decode replace trusted node DAO member response: %w", err)
    }
    if response.Error != "" {
        return api.ReplaceTNDAOPositionResponse{}, fmt.Errorf("Could not replace trusted node DAO member: %s", response.Error)
    }
    return response, nil
}

