package rocketpool

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get upgrade proposals
func (c *Client) TNDAOUpgradeProposals() (api.TNDAOGetUpgradeProposalsResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/upgrade/get-upgrade-proposals", nil)
	if err != nil {
		return api.TNDAOGetUpgradeProposalsResponse{}, fmt.Errorf("Could not get upgrade proposals: %w", err)
	}
	var response api.TNDAOGetUpgradeProposalsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.TNDAOGetUpgradeProposalsResponse{}, fmt.Errorf("Could not decode upgrade proposals response: %w", err)
	}
	if response.Error != "" {
		return api.TNDAOGetUpgradeProposalsResponse{}, fmt.Errorf("Could not get upgrade proposals: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can execute a proposal
func (c *Client) CanExecuteUpgradeProposal(proposalId uint64) (api.CanExecuteUpgradeProposalResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/upgrade/can-execute-upgrade", url.Values{"id": {strconv.FormatUint(proposalId, 10)}})
	if err != nil {
		return api.CanExecuteUpgradeProposalResponse{}, fmt.Errorf("Could not check whether the node can execute upgrade proposal: %w", err)
	}
	var response api.CanExecuteUpgradeProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanExecuteUpgradeProposalResponse{}, fmt.Errorf("Could not decode can execute upgrade proposal response: %w", err)
	}
	if response.Error != "" {
		return api.CanExecuteUpgradeProposalResponse{}, fmt.Errorf("Could not check whether the node can execute upgrade proposal: %s", response.Error)
	}
	return response, nil
}

// Execute a proposal
func (c *Client) ExecuteUpgradeProposal(proposalId uint64) (api.ExecuteUpgradeProposalResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/upgrade/execute-upgrade", url.Values{"id": {strconv.FormatUint(proposalId, 10)}})
	if err != nil {
		return api.ExecuteUpgradeProposalResponse{}, fmt.Errorf("Could not execute upgrade proposal: %w", err)
	}
	var response api.ExecuteUpgradeProposalResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ExecuteUpgradeProposalResponse{}, fmt.Errorf("Could not decode execute upgrade proposal response: %w", err)
	}
	if response.Error != "" {
		return api.ExecuteUpgradeProposalResponse{}, fmt.Errorf("Could not execute upgrade proposal: %s", response.Error)
	}
	return response, nil
}
