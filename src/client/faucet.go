package client

import (
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type FaucetRequester struct {
	context *client.RequesterContext
}

func NewFaucetRequester(context *client.RequesterContext) *FaucetRequester {
	return &FaucetRequester{
		context: context,
	}
}

func (r *FaucetRequester) GetName() string {
	return "Faucet"
}
func (r *FaucetRequester) GetRoute() string {
	return "faucet"
}
func (r *FaucetRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Get faucet status
func (r *FaucetRequester) Status() (*types.ApiResponse[api.FaucetStatusData], error) {
	return client.SendGetRequest[api.FaucetStatusData](r, "status", "Status", nil)
}

// Withdraw RPL from the faucet
func (r *FaucetRequester) WithdrawRpl() (*types.ApiResponse[api.FaucetWithdrawRplData], error) {
	return client.SendGetRequest[api.FaucetWithdrawRplData](r, "withdraw-rpl", "WithdrawRpl", nil)
}
