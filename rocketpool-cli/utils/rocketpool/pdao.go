package rocketpool

import (
	"net/http"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type PDaoRequester struct {
	client *http.Client
}

func NewPDaoRequester(client *http.Client) *PDaoRequester {
	return &PDaoRequester{
		client: client,
	}
}

func (r *PDaoRequester) GetName() string {
	return "PDAO"
}
func (r *PDaoRequester) GetRoute() string {
	return "pdao"
}
func (r *PDaoRequester) GetClient() *http.Client {
	return r.client
}

// Claim / unlock bonds from a proposal
func (r *PDaoRequester) ClaimBonds() (*api.ApiResponse[api.ProtocolDaoClaimBondsData], error) {
	return sendGetRequest[api.ProtocolDaoClaimBondsData](r, "claim-bonds", "Claim Bonds", nil)
}
