package network

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

type NetworkHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewNetworkHandler(logger *log.Logger, ctx context.Context, serviceProvider *services.ServiceProvider) *NetworkHandler {
	h := &NetworkHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
	}
	h.factories = []server.IContextFactory{
		&networkProposalContextFactory{h},
		&networkDelegateContextFactory{h},
		&networkDepositInfoContextFactory{h},
		&networkDownloadRewardsContextFactory{h},
		&networkRewardsFileContextFactory{h},
		&networkGenerateRewardsContextFactory{h},
		&networkFeeContextFactory{h},
		&networkPriceContextFactory{h},
		&networkStatsContextFactory{h},
		&networkTimezoneContextFactory{h},
	}
	return h
}

func (h *NetworkHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/network").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
