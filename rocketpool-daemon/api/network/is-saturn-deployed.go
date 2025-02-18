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

type saturnDeployedContextFactory struct {
	handler *NetworkHandler
}

func (f *saturnDeployedContextFactory) Create(args url.Values) (*saturnDeployedContext, error) {
	c := &saturnDeployedContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *saturnDeployedContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*saturnDeployedContext, api.NetworkSaturnDeployedData](
		router, "is-hotfix-deployed", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type saturnDeployedContext struct {
	handler *NetworkHandler
}

func (c *saturnDeployedContext) PrepareData(data *api.NetworkSaturnDeployedData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	isSaturnDeployed, err := state.IsHoustonHotfixDeployed(rp, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting the protocol version: %w", err)
	}
	data.IsSaturnDeployed = isSaturnDeployed

	return types.ResponseStatus_Success, nil
}
