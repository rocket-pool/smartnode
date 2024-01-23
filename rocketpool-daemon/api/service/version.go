package service

import (
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
		router, "version", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type serviceVersionContext struct {
	handler *ServiceHandler
}

func (c *serviceVersionContext) PrepareData(data *api.ServiceVersionData, opts *bind.TransactOpts) error {
	data.Version = shared.RocketPoolVersion
	return nil
}
