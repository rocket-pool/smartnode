package rocketpool

import (
	"net/http"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type FaucetRequester struct {
	client *http.Client
}

func NewFaucetRequester(client *http.Client) *FaucetRequester {
	return &FaucetRequester{
		client: client,
	}
}

func (r *FaucetRequester) GetName() string {
	return "Faucet"
}
func (r *FaucetRequester) GetRoute() string {
	return "faucet"
}
func (r *FaucetRequester) GetClient() *http.Client {
	return r.client
}

// Get faucet status
func (r *FaucetRequester) Status() (*api.ApiResponse[api.FaucetStatusData], error) {
	return sendGetRequest[api.FaucetStatusData](r, "status", "Status", nil)
}

// Withdraw RPL from the faucet
func (r *FaucetRequester) WithdrawRpl() (*api.ApiResponse[api.FaucetWithdrawRplData], error) {
	return sendGetRequest[api.FaucetWithdrawRplData](r, "withdraw-rpl", "WithdrawRpl", nil)
}
