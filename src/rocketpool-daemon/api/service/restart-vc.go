package service

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/validator"
)

// ===============
// === Factory ===
// ===============

type serviceRestartVcContextFactory struct {
	handler *ServiceHandler
}

func (f *serviceRestartVcContextFactory) Create(args url.Values) (*serviceRestartVcContext, error) {
	c := &serviceRestartVcContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *serviceRestartVcContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*serviceRestartVcContext, types.SuccessData](
		router, "restart-vc", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type serviceRestartVcContext struct {
	handler *ServiceHandler
}

func (c *serviceRestartVcContext) PrepareData(data *types.SuccessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	bc := sp.GetBeaconClient()
	d := sp.GetDocker()

	err := validator.StopValidator(c.handler.ctx, cfg, bc, d, true)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error restarting validator client: %w", err)
	}
	return types.ResponseStatus_Success, nil
}
