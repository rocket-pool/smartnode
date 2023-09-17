package faucet

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
)

type FaucetHandler struct {
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewFaucetHandler(serviceProvider *services.ServiceProvider) *FaucetHandler {
	h := &FaucetHandler{
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&faucetStatusContextFactory{h},
		&faucetWithdrawContextFactory{h},
	}
	return h
}

func (h *FaucetHandler) RegisterRoutes(router *mux.Router) {
	for _, factory := range h.factories {
		factory.RegisterRoute(router)
	}
}
