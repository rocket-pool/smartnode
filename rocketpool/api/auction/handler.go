package auction

import (
	"github.com/gorilla/mux"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type AuctionHandler struct {
	serviceProvider *services.ServiceProvider
	bidFactory      server.ISingleStageCallContextFactory[*auctionBidContext, api.BidOnLotData]
	claimFactory    server.ISingleStageCallContextFactory[*auctionClaimContext, api.ClaimFromLotData]
	createFactory   server.ISingleStageCallContextFactory[*auctionCreateContext, api.CreateLotData]
	lotsFactory     server.ISingleStageCallContextFactory[*auctionLotContext, api.AuctionLotsData]
	recoverFactory  server.ISingleStageCallContextFactory[*auctionRecoverContext, api.RecoverRplFromLotData]
	statusFactory   server.ISingleStageCallContextFactory[*auctionStatusContext, api.AuctionStatusData]
}

func NewAuctionHandler(serviceProvider *services.ServiceProvider) *AuctionHandler {
	h := &AuctionHandler{
		serviceProvider: serviceProvider,
	}
	h.bidFactory = &auctionBidContextFactory{h}
	h.claimFactory = &auctionClaimContextFactory{h}
	h.createFactory = &auctionCreateContextFactory{h}
	h.lotsFactory = &auctionLotContextFactory{h}
	h.recoverFactory = &auctionRecoverContextFactory{h}
	h.statusFactory = &auctionStatusContextFactory{h}
	return h
}

func (h *AuctionHandler) RegisterRoutes(router *mux.Router) {
	server.RegisterSingleStageRoute(router, "bid-lot", h.bidFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "claim-lot", h.claimFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "create-lot", h.createFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "lots", h.lotsFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "recover-lot", h.recoverFactory, h.serviceProvider)
	server.RegisterSingleStageRoute(router, "status", h.statusFactory, h.serviceProvider)
}
