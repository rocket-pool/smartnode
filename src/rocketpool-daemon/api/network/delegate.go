package network

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkDelegateContextFactory struct {
	handler *NetworkHandler
}

func (f *networkDelegateContextFactory) Create(args url.Values) (*networkDelegateContext, error) {
	c := &networkDelegateContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkDelegateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*networkDelegateContext, api.NetworkLatestDelegateData](
		router, "latest-delegate", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkDelegateContext struct {
	handler *NetworkHandler
}

func (c *networkDelegateContext) PrepareData(data *api.NetworkLatestDelegateData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	delegateContract, err := rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting minipool delegate contract: %w", err)
	}

	data.Address = delegateContract.Address
	return types.ResponseStatus_Success, nil
}
