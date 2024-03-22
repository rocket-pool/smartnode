package service

import (
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type serviceClientStatusContextFactory struct {
	handler *ServiceHandler
}

func (f *serviceClientStatusContextFactory) Create(args url.Values) (*serviceClientStatusContext, error) {
	c := &serviceClientStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *serviceClientStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*serviceClientStatusContext, api.ServiceClientStatusData](
		router, "client-status", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type serviceClientStatusContext struct {
	handler *ServiceHandler
}

func (c *serviceClientStatusContext) PrepareData(data *api.ServiceClientStatusData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	ec := sp.GetEthClient()
	bc := sp.GetBeaconClient()
	ctx := sp.GetContext()

	// Get the EC manager status
	ecMgrStatus := ec.CheckStatus(ctx)
	data.EcManagerStatus = *ecMgrStatus

	// Get the BC manager status
	bcMgrStatus := bc.CheckStatus(ctx)
	data.BcManagerStatus = *bcMgrStatus

	return nil
}
