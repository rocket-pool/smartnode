package rocketpool

import (
	"fmt"
	"net/url"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get security council status
func (c *Client) SecurityStatus() (api.SecurityStatusResponse, error) {
	return c.callAPI[api.SecurityStatusResponse]("GET", "/api/security/status", nil, "Could not get security council status")
}

// Get the security council members
func (c *Client) SecurityMembers() (api.SecurityMembersResponse, error) {
	return c.callAPI[api.SecurityMembersResponse]("GET", "/api/security/members", nil, "Could not get security council members")
}

// Get the security council proposals
func (c *Client) SecurityProposals() (api.SecurityProposalsResponse, error) {
	return c.callAPI[api.SecurityProposalsResponse]("GET", "/api/security/proposals", nil, "Could not get security council proposals")
}

// Get details of a proposal
func (c *Client) SecurityProposal(id uint64) (api.SecurityProposalResponse, error) {
	return c.callAPI[api.SecurityProposalResponse]("GET", "/api/security/proposal-details", url.Values{"id": {fmt.Sprintf("%d", id)}}, "Could not get security council proposal")
}

// Check whether the node can propose to leave the security council
func (c *Client) SecurityProposeLeave() (api.SecurityProposeLeaveResponse, error) {
	return c.callAPI[api.SecurityProposeLeaveResponse]("POST", "/api/security/propose-leave", nil, "Could not get security-propose-leave status")
}

// Check whether the node can propose leaving the security council
func (c *Client) SecurityCanProposeLeave() (api.SecurityCanProposeLeaveResponse, error) {
	return c.callAPI[api.SecurityCanProposeLeaveResponse]("GET", "/api/security/can-propose-leave", nil, "Could not get security-can-propose-leave status")
}

// Check whether the node can cancel a proposal
func (c *Client) SecurityCanCancelProposal(proposalId uint64) (api.SecurityCanCancelProposalResponse, error) {
	return c.callAPI[api.SecurityCanCancelProposalResponse]("GET", "/api/security/can-cancel-proposal", url.Values{"id": {fmt.Sprintf("%d", proposalId)}}, "Could not get security-can-cancel-proposal status")
}

// Cancel a proposal made by the node
func (c *Client) SecurityCancelProposal(proposalId uint64) (api.SecurityCancelProposalResponse, error) {
	return c.callAPI[api.SecurityCancelProposalResponse]("POST", "/api/security/cancel-proposal", url.Values{"id": {fmt.Sprintf("%d", proposalId)}}, "Could not cancel security council proposal")
}

// Check whether the node can vote on a proposal
func (c *Client) SecurityCanVoteOnProposal(proposalId uint64) (api.SecurityCanVoteOnProposalResponse, error) {
	return c.callAPI[api.SecurityCanVoteOnProposalResponse]("GET", "/api/security/can-vote-proposal", url.Values{"id": {fmt.Sprintf("%d", proposalId)}}, "Could not get security-can-vote-on-proposal status")
}

// Vote on a proposal
func (c *Client) SecurityVoteOnProposal(proposalId uint64, support bool) (api.SecurityVoteOnProposalResponse, error) {
	supportStr := "false"
	if support {
		supportStr = "true"
	}
	return c.callAPI[api.SecurityVoteOnProposalResponse]("POST", "/api/security/vote-proposal", url.Values{
		"id":      {fmt.Sprintf("%d", proposalId)},
		"support": {supportStr},
	}, "Could not vote on security council proposal")
}

// Check whether the node can execute a proposal
func (c *Client) SecurityCanExecuteProposal(proposalId uint64) (api.SecurityCanExecuteProposalResponse, error) {
	return c.callAPI[api.SecurityCanExecuteProposalResponse]("GET", "/api/security/can-execute-proposal", url.Values{"id": {fmt.Sprintf("%d", proposalId)}}, "Could not get security-can-execute-proposal status")
}

// Execute a proposal
func (c *Client) SecurityExecuteProposal(proposalId uint64) (api.SecurityExecuteProposalResponse, error) {
	return c.callAPI[api.SecurityExecuteProposalResponse]("POST", "/api/security/execute-proposal", url.Values{"id": {fmt.Sprintf("%d", proposalId)}}, "Could not execute security council proposal")
}

// Check whether the node can join the security council
func (c *Client) SecurityCanJoin() (api.SecurityCanJoinResponse, error) {
	return c.callAPI[api.SecurityCanJoinResponse]("GET", "/api/security/can-join", nil, "Could not get security-can-join status")
}

// Join the security council (requires an executed invite proposal)
func (c *Client) SecurityJoin() (api.SecurityJoinResponse, error) {
	return c.callAPI[api.SecurityJoinResponse]("POST", "/api/security/join", nil, "Could not join security council")
}

// Check whether the node can leave the security council
func (c *Client) SecurityCanLeave() (api.SecurityCanLeaveResponse, error) {
	return c.callAPI[api.SecurityCanLeaveResponse]("GET", "/api/security/can-leave", nil, "Could not get security-can-leave status")
}

// Leave the security council (requires an executed leave proposal)
func (c *Client) SecurityLeave() (api.SecurityLeaveResponse, error) {
	return c.callAPI[api.SecurityLeaveResponse]("POST", "/api/security/leave", nil, "Could not leave security council")
}

// Check whether the node can propose updating a PDAO setting
func (c *Client) SecurityCanProposeSetting(contract string, setting string, value string) (api.SecurityCanProposeSettingResponse, error) {
	return c.callAPI[api.SecurityCanProposeSettingResponse]("GET", "/api/security/can-propose-setting", url.Values{
		"contractName": {contract},
		"settingName":  {setting},
		"value":        {value},
	}, "Could not get security-can-propose-setting")
}

// Propose updating a PDAO setting
func (c *Client) SecurityProposeSetting(contract string, setting string, value string) (api.SecurityProposeSettingResponse, error) {
	return c.callAPI[api.SecurityProposeSettingResponse]("POST", "/api/security/propose-setting", url.Values{
		"contractName": {contract},
		"settingName":  {setting},
		"value":        {value},
	}, "Could not propose security council setting")
}
