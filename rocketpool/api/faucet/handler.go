package faucet

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type FaucetHandler struct {
	serviceProvider *services.ServiceProvider
	statusFactory   server.ISingleStageCallContextFactory[*faucetStatusContext, api.FaucetStatusData]
	withdrawFactory server.ISingleStageCallContextFactory[*faucetWithdrawContext, api.FaucetWithdrawRplData]
}

func NewFaucetHandler(serviceProvider *services.ServiceProvider) *FaucetHandler {
	h := &FaucetHandler{
		serviceProvider: serviceProvider,
	}
	h.statusFactory = &faucetStatusContextFactory{h}
	h.withdrawFactory = &faucetWithdrawContextFactory{h}
	return h
}

func (h *FaucetHandler) RegisterRoutes(router *mux.Router) {
	server.RegisterSingleStageRoute(router, "status", h.statusFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "withdraw-rpl", h.withdrawFactory, h.serviceProvider)
}
