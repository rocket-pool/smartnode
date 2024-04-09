package service

import (
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type serviceVersionContextFactory struct {
	handler *ServiceHandler
}

func (f *serviceVersionContextFactory) Create(args url.Values) (*serviceVersionContext, error) {
	c := &serviceVersionContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *serviceVersionContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*serviceVersionContext, api.ServiceVersionData](
		router, "version", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type serviceVersionContext struct {
	handler *ServiceHandler
}

func (c *serviceVersionContext) PrepareData(data *api.ServiceVersionData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.Version = shared.RocketPoolVersion
	return types.ResponseStatus_Success, nil
}
