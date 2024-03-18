package network

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

type NetworkHandler struct {
	serviceProvider *services.ServiceProvider
	context         context.Context
	factories       []server.IContextFactory
}

func NewNetworkHandler(context context.Context, serviceProvider *services.ServiceProvider) *NetworkHandler {
	h := &NetworkHandler{
		serviceProvider: serviceProvider,
		context:         context,
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
		&networkInitializeVotingContextFactory{h},
		&networkSetVotingDelegateContextFactory{h},
		&networkCurrentVotingDelegateContextFactory{h},
	}
	return h
}

func (h *NetworkHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/network").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
