package service

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/validator"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
	server.RegisterQuerylessGet[*serviceRestartVcContext, api.SuccessData](
		router, "restart-vc", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type serviceRestartVcContext struct {
	handler *ServiceHandler
}

func (c *serviceRestartVcContext) PrepareData(data *api.SuccessData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	cfg := sp.GetConfig()
	bc := sp.GetBeaconClient()
	d := sp.GetDocker()

	err := validator.RestartValidator(cfg, bc, nil, d)
	if err != nil {
		return fmt.Errorf("error restarting validator client: %w", err)
	}

	data.Success = true
	return nil
}
