package auction

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/services"
)

type AuctionHandler struct {
	serviceProvider *services.ServiceProvider
	context         context.Context
	factories       []server.IContextFactory
}

func NewAuctionHandler(context context.Context, serviceProvider *services.ServiceProvider) *AuctionHandler {
	h := &AuctionHandler{
		serviceProvider: serviceProvider,
		context:         context,
	}
	h.factories = []server.IContextFactory{
		&auctionBidContextFactory{h},
		&auctionClaimContextFactory{h},
		&auctionCreateContextFactory{h},
		&auctionLotContextFactory{h},
		&auctionRecoverContextFactory{h},
		&auctionStatusContextFactory{h},
	}
	return h
}

func (h *AuctionHandler) RegisterRoutes(router *mux.Router) {
	subrouter := router.PathPrefix("/auction").Subrouter()
	for _, factory := range h.factories {
		factory.RegisterRoute(subrouter)
	}
}
