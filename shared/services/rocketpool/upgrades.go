package rocketpool

import (
	"net/url"
	"strconv"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get upgrade proposals
func (c *Client) TNDAOUpgradeProposals() (api.TNDAOGetUpgradeProposalsResponse, error) {
	return c.callAPI[api.TNDAOGetUpgradeProposalsResponse]("GET", "/api/upgrade/get-upgrade-proposals", nil, "Could not get upgrade proposals")
}

// Check whether the node can execute a proposal
func (c *Client) CanExecuteUpgradeProposal(proposalId uint64) (api.CanExecuteUpgradeProposalResponse, error) {
	return c.callAPI[api.CanExecuteUpgradeProposalResponse]("GET", "/api/upgrade/can-execute-upgrade", url.Values{"id": {strconv.FormatUint(proposalId, 10)}}, "Could not check whether the node can execute upgrade proposal")
}

// Execute a proposal
func (c *Client) ExecuteUpgradeProposal(proposalId uint64) (api.ExecuteUpgradeProposalResponse, error) {
	return c.callAPI[api.ExecuteUpgradeProposalResponse]("POST", "/api/upgrade/execute-upgrade", url.Values{"id": {strconv.FormatUint(proposalId, 10)}}, "Could not execute upgrade proposal")
}
