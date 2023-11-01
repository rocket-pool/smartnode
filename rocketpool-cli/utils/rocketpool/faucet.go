package rocketpool

import (
	"fmt"
	"net/http"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type FaucetRequester struct {
	client *http.Client
	route  string
}

func NewFaucetRequester(client *http.Client) *FaucetRequester {
	return &FaucetRequester{
		client: client,
		route:  "faucet",
	}
}

// Get faucet status
func (r *FaucetRequester) Status() (*api.ApiResponse[api.FaucetStatusData], error) {
	method := "status"
	args := map[string]string{}
	response, err := SendGetRequest[api.FaucetStatusData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Faucet Status request: %w", err)
	}
	return response, nil
}

// Withdraw RPL from the faucet
func (r *FaucetRequester) WithdrawRpl() (*api.ApiResponse[api.FaucetWithdrawRplData], error) {
	method := "withdraw-rpl"
	args := map[string]string{}
	response, err := SendGetRequest[api.FaucetWithdrawRplData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Faucet WithdrawRpl request: %w", err)
	}
	return response, nil
}
