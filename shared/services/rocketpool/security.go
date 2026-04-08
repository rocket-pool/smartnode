package rocketpool

import (
	"fmt"
	"net/url"

	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get security council status
func (c *Client) SecurityStatus() (api.SecurityStatusResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/status", nil)
	if err != nil {
		return api.SecurityStatusResponse{}, fmt.Errorf("Could not get security council status: %w", err)
	}
	var response api.SecurityStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityStatusResponse{}, fmt.Errorf("Could not decode security council stats response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityStatusResponse{}, fmt.Errorf("Could not get security council status: %s", response.Error)
	}
	return response, nil
}

// Get the security council members
func (c *Client) SecurityMembers() (api.SecurityMembersResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/members", nil)
	if err != nil {
		return api.SecurityMembersResponse{}, fmt.Errorf("Could not get security council members: %w", err)
	}
	var response api.SecurityMembersResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityMembersResponse{}, fmt.Errorf("Could not decode security council members response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityMembersResponse{}, fmt.Errorf("Could not get security council members: %s", response.Error)
	}
	return response, nil
}

// Get the security council proposals
func (c *Client) SecurityProposals() (api.SecurityProposalsResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/proposals", nil)
	if err != nil {
		return api.SecurityProposalsResponse{}, fmt.Errorf("Could not get security council proposals: %w", err)
	}
	var response api.SecurityProposalsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityProposalsResponse{}, fmt.Errorf("Could not decode security council proposals response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityProposalsResponse{}, fmt.Errorf("Could not get security council proposals: %s", response.Error)
	}
	return response, nil
}

// Get details of a proposal
func (c *Client) SecurityProposal(id uint64) (api.SecurityProposalResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/proposal-details", url.Values{"id": {fmt.Sprintf("%d", id)}})
	if err != nil {
		return api.SecurityProposalResponse{}, fmt.Errorf("Could not get security council proposal: %w", err)
	}
	var response api.SecurityProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityProposalResponse{}, fmt.Errorf("Could not decode security council proposal response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityProposalResponse{}, fmt.Errorf("Could not get security council proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose to leave the security council
func (c *Client) SecurityProposeLeave() (api.SecurityProposeLeaveResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/security/propose-leave", nil)
	if err != nil {
		return api.SecurityProposeLeaveResponse{}, fmt.Errorf("Could not get security-propose-leave status: %w", err)
	}
	var response api.SecurityProposeLeaveResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityProposeLeaveResponse{}, fmt.Errorf("Could not decode security-propose-leave response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityProposeLeaveResponse{}, fmt.Errorf("Could not get security-propose-leave status: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose leaving the security council
func (c *Client) SecurityCanProposeLeave() (api.SecurityCanProposeLeaveResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/can-propose-leave", nil)
	if err != nil {
		return api.SecurityCanProposeLeaveResponse{}, fmt.Errorf("Could not get security-can-propose-leave status: %w", err)
	}
	var response api.SecurityCanProposeLeaveResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityCanProposeLeaveResponse{}, fmt.Errorf("Could not decode security-can-propose-leave response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityCanProposeLeaveResponse{}, fmt.Errorf("Could not get security-can-propose-leave status: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can cancel a proposal
func (c *Client) SecurityCanCancelProposal(proposalId uint64) (api.SecurityCanCancelProposalResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/can-cancel-proposal", url.Values{"id": {fmt.Sprintf("%d", proposalId)}})
	if err != nil {
		return api.SecurityCanCancelProposalResponse{}, fmt.Errorf("Could not get security-can-cancel-proposal status: %w", err)
	}
	var response api.SecurityCanCancelProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityCanCancelProposalResponse{}, fmt.Errorf("Could not decode security-can-cancel-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityCanCancelProposalResponse{}, fmt.Errorf("Could not get security-can-cancel-proposal status: %s", response.Error)
	}
	return response, nil
}

// Cancel a proposal made by the node
func (c *Client) SecurityCancelProposal(proposalId uint64) (api.SecurityCancelProposalResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/security/cancel-proposal", url.Values{"id": {fmt.Sprintf("%d", proposalId)}})
	if err != nil {
		return api.SecurityCancelProposalResponse{}, fmt.Errorf("Could not cancel security council proposal: %w", err)
	}
	var response api.SecurityCancelProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityCancelProposalResponse{}, fmt.Errorf("Could not decode cancel security council proposal response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityCancelProposalResponse{}, fmt.Errorf("Could not cancel security council proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can vote on a proposal
func (c *Client) SecurityCanVoteOnProposal(proposalId uint64) (api.SecurityCanVoteOnProposalResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/can-vote-proposal", url.Values{"id": {fmt.Sprintf("%d", proposalId)}})
	if err != nil {
		return api.SecurityCanVoteOnProposalResponse{}, fmt.Errorf("Could not get security-can-vote-on-proposal status: %w", err)
	}
	var response api.SecurityCanVoteOnProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityCanVoteOnProposalResponse{}, fmt.Errorf("Could not decode security-can-vote-on-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityCanVoteOnProposalResponse{}, fmt.Errorf("Could not get security-can-vote-on-proposal status: %s", response.Error)
	}
	return response, nil
}

// Vote on a proposal
func (c *Client) SecurityVoteOnProposal(proposalId uint64, support bool) (api.SecurityVoteOnProposalResponse, error) {
	supportStr := "false"
	if support {
		supportStr = "true"
	}
	responseBytes, err := c.callHTTPAPI("POST", "/api/security/vote-proposal", url.Values{
		"id":      {fmt.Sprintf("%d", proposalId)},
		"support": {supportStr},
	})
	if err != nil {
		return api.SecurityVoteOnProposalResponse{}, fmt.Errorf("Could not vote on security council proposal: %w", err)
	}
	var response api.SecurityVoteOnProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityVoteOnProposalResponse{}, fmt.Errorf("Could not decode vote on security council proposal response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityVoteOnProposalResponse{}, fmt.Errorf("Could not vote on security council proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can execute a proposal
func (c *Client) SecurityCanExecuteProposal(proposalId uint64) (api.SecurityCanExecuteProposalResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/can-execute-proposal", url.Values{"id": {fmt.Sprintf("%d", proposalId)}})
	if err != nil {
		return api.SecurityCanExecuteProposalResponse{}, fmt.Errorf("Could not get security-can-execute-proposal status: %w", err)
	}
	var response api.SecurityCanExecuteProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityCanExecuteProposalResponse{}, fmt.Errorf("Could not decode security-can-execute-proposal response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityCanExecuteProposalResponse{}, fmt.Errorf("Could not get security-can-execute-proposal status: %s", response.Error)
	}
	return response, nil
}

// Execute a proposal
func (c *Client) SecurityExecuteProposal(proposalId uint64) (api.SecurityExecuteProposalResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/security/execute-proposal", url.Values{"id": {fmt.Sprintf("%d", proposalId)}})
	if err != nil {
		return api.SecurityExecuteProposalResponse{}, fmt.Errorf("Could not execute security council proposal: %w", err)
	}
	var response api.SecurityExecuteProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityExecuteProposalResponse{}, fmt.Errorf("Could not decode execute security council proposal response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityExecuteProposalResponse{}, fmt.Errorf("Could not execute security council proposal: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can join the security council
func (c *Client) SecurityCanJoin() (api.SecurityCanJoinResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/can-join", nil)
	if err != nil {
		return api.SecurityCanJoinResponse{}, fmt.Errorf("Could not get security-can-join status: %w", err)
	}
	var response api.SecurityCanJoinResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityCanJoinResponse{}, fmt.Errorf("Could not decode security-can-join response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityCanJoinResponse{}, fmt.Errorf("Could not get security-can-join status: %s", response.Error)
	}
	return response, nil
}

// Join the security council (requires an executed invite proposal)
func (c *Client) SecurityJoin() (api.SecurityJoinResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/security/join", nil)
	if err != nil {
		return api.SecurityJoinResponse{}, fmt.Errorf("Could not join security council: %w", err)
	}
	var response api.SecurityJoinResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityJoinResponse{}, fmt.Errorf("Could not decode join security council response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityJoinResponse{}, fmt.Errorf("Could not join security council: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can leave the security council
func (c *Client) SecurityCanLeave() (api.SecurityCanLeaveResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/can-leave", nil)
	if err != nil {
		return api.SecurityCanLeaveResponse{}, fmt.Errorf("Could not get security-can-leave status: %w", err)
	}
	var response api.SecurityCanLeaveResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityCanLeaveResponse{}, fmt.Errorf("Could not decode security-can-leave response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityCanLeaveResponse{}, fmt.Errorf("Could not get security-can-leave status: %s", response.Error)
	}
	return response, nil
}

// Leave the security council (requires an executed leave proposal)
func (c *Client) SecurityLeave() (api.SecurityLeaveResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/security/leave", nil)
	if err != nil {
		return api.SecurityLeaveResponse{}, fmt.Errorf("Could not leave security council: %w", err)
	}
	var response api.SecurityLeaveResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityLeaveResponse{}, fmt.Errorf("Could not decode leave security council response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityLeaveResponse{}, fmt.Errorf("Could not leave security council: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can propose updating a PDAO setting
func (c *Client) SecurityCanProposeSetting(contract string, setting string, value string) (api.SecurityCanProposeSettingResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/security/can-propose-setting", url.Values{
		"contractName": {contract},
		"settingName":  {setting},
		"value":        {value},
	})
	if err != nil {
		return api.SecurityCanProposeSettingResponse{}, fmt.Errorf("Could not get security-can-propose-setting: %w", err)
	}
	var response api.SecurityCanProposeSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityCanProposeSettingResponse{}, fmt.Errorf("Could not decode security-can-propose-setting response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityCanProposeSettingResponse{}, fmt.Errorf("Could not get security-can-propose-setting: %s", response.Error)
	}
	return response, nil
}

// Propose updating a PDAO setting
func (c *Client) SecurityProposeSetting(contract string, setting string, value string) (api.SecurityProposeSettingResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/security/propose-setting", url.Values{
		"contractName": {contract},
		"settingName":  {setting},
		"value":        {value},
	})
	if err != nil {
		return api.SecurityProposeSettingResponse{}, fmt.Errorf("Could not propose security council setting: %w", err)
	}
	var response api.SecurityProposeSettingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SecurityProposeSettingResponse{}, fmt.Errorf("Could not decode propose security council setting response: %w", err)
	}
	if response.Error != "" {
		return api.SecurityProposeSettingResponse{}, fmt.Errorf("Could not propose security council setting: %s", response.Error)
	}
	return response, nil
}
