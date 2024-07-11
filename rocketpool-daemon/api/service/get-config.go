package service

import (
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type serviceGetConfigContextFactory struct {
	handler *ServiceHandler
}

func (f *serviceGetConfigContextFactory) Create(args url.Values) (*serviceGetConfigContext, error) {
	c := &serviceGetConfigContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *serviceGetConfigContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*serviceGetConfigContext, api.ServiceGetConfigData](
		router, "get-config", f, f.handler.logger.Logger, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type serviceGetConfigContext struct {
	handler *ServiceHandler
}

func (c *serviceGetConfigContext) PrepareData(data *api.ServiceGetConfigData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()

	data.Config = cfg.Serialize()
	return types.ResponseStatus_Success, nil
}
