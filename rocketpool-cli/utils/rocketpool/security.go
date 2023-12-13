package rocketpool

import (
	"fmt"
	"net/http"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type SecurityRequester struct {
	client *http.Client
}

func NewSecurityRequester(client *http.Client) *SecurityRequester {
	return &SecurityRequester{
		client: client,
	}
}

func (r *SecurityRequester) GetName() string {
	return "Security Council"
}
func (r *SecurityRequester) GetRoute() string {
	return "security"
}
func (r *SecurityRequester) GetClient() *http.Client {
	return r.client
}

// Cancel a proposal made by the node
func (r *SecurityRequester) CancelProposal(id uint64) (*api.ApiResponse[api.SecurityCancelProposalData], error) {
	args := map[string]string{
		"id": fmt.Sprint(id),
	}
	return sendGetRequest[api.SecurityCancelProposalData](r, "proposal/cancel", "CancelProposal", args)
}

// Execute a proposal
func (r *SecurityRequester) ExecuteProposal(id uint64) (*api.ApiResponse[api.SecurityExecuteProposalData], error) {
	args := map[string]string{
		"id": fmt.Sprint(id),
	}
	return sendGetRequest[api.SecurityExecuteProposalData](r, "proposal/execute", "ExecuteProposal", args)
}

// Join the security council after being invited
func (r *SecurityRequester) Join() (*api.ApiResponse[api.SecurityJoinData], error) {
	return sendGetRequest[api.SecurityJoinData](r, "join", "Join", nil)
}

// Leave the security council after the proposal to leave has passed
func (r *SecurityRequester) Leave() (*api.ApiResponse[api.SecurityLeaveData], error) {
	return sendGetRequest[api.SecurityLeaveData](r, "leave", "Leave", nil)
}
