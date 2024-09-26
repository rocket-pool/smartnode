package network

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkHotfixDeployedContextFactory struct {
	handler *NetworkHandler
}

func (f *networkHotfixDeployedContextFactory) Create(args url.Values) (*networkHotfixDeployedContext, error) {
	c := &networkHotfixDeployedContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkHotfixDeployedContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*networkHotfixDeployedContext, api.NetworkHotfixDeployedData](
		router, "is-hotfix-deployed", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkHotfixDeployedContext struct {
	handler *NetworkHandler
}

func (c *networkHotfixDeployedContext) PrepareData(data *api.NetworkHotfixDeployedData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	houstonHotfixDeployed, err := state.IsHoustonHotfixDeployed(rp, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting the protocol version: %w", err)
	}
	data.IsHoustonHotfixDeployed = houstonHotfixDeployed

	return types.ResponseStatus_Success, nil
}
