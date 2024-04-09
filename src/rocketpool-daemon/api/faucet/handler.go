package faucet

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

type FaucetHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewFaucetHandler(logger *log.Logger, ctx context.Context, serviceProvider *services.ServiceProvider) *FaucetHandler {
	h := &FaucetHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&faucetStatusContextFactory{h},
		&faucetWithdrawContextFactory{h},
	}
	return h
}

func (h *FaucetHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/faucet").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
