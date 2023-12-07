package rocketpool

import (
	"fmt"
	"net/http"
	"strings"

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
func (r *PDaoRequester) ClaimBonds(proposalID uint64, indices []uint64) (*api.ApiResponse[api.ProtocolDaoClaimBondsData], error) {
	indicesStrings := make([]string, len(indices))
	for i, index := range indices {
		indicesStrings[i] = fmt.Sprint(index)
	}
	args := map[string]string{
		"proposal-id": fmt.Sprint(proposalID),
		"indices":     strings.Join(indicesStrings, ","),
	}
	return sendGetRequest[api.ProtocolDaoClaimBondsData](r, "claim-bonds", "Claim Bonds", args)
}

// Claim / unlock bonds from a proposal
func (r *PDaoRequester) GetClaimableBonds() (*api.ApiResponse[api.ProtocolDaoGetClaimableBondsData], error) {
	return sendGetRequest[api.ProtocolDaoGetClaimableBondsData](r, "get-claimable-bonds", "Get Claimable Bonds", nil)
}
