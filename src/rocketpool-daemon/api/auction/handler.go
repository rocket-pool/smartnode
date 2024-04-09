package auction

import (
	"context"

	"github.com/gorilla/mux"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/log"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/services"
)

type AuctionHandler struct {
	logger          *log.Logger
	ctx             context.Context
	serviceProvider *services.ServiceProvider
	factories       []server.IContextFactory
}

func NewAuctionHandler(logger *log.Logger, ctx context.Context, serviceProvider *services.ServiceProvider) *AuctionHandler {
	h := &AuctionHandler{
		logger:          logger,
		ctx:             ctx,
		serviceProvider: serviceProvider,
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
