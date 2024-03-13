package faucet

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

type FaucetHandler struct {
	serviceProvider *services.ServiceProvider
	context         context.Context
	factories       []server.IContextFactory
}

func NewFaucetHandler(context context.Context, serviceProvider *services.ServiceProvider) *FaucetHandler {
	h := &FaucetHandler{
		serviceProvider: serviceProvider,
		context:         context,
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
