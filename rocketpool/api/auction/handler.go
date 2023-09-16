package auction

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	wtypes "github.com/rocket-pool/smartnode/shared/types/wallet"
)

// ===============
// === Handler ===
// ===============

type AuctionHandler struct {
	serviceProvider *services.ServiceProvider
	bidFactory      server.IContextFactory[*auctionBidContext, api.BidOnLotData, commonContext]
	claimFactory    server.IContextFactory[*auctionClaimContext, api.ClaimFromLotData, commonContext]
	createFactory   server.IContextFactory[*auctionCreateContext, api.CreateLotData, commonContext]
	lotsFactory     server.IContextFactory[*auctionLotContext, api.AuctionLotsData, commonContext]
	recoverFactory  server.IContextFactory[*auctionRecoverContext, api.RecoverRplFromLotData, commonContext]
	statusFactory   server.IContextFactory[*auctionStatusContext, api.AuctionStatusData, commonContext]
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
	server.RegisterSingleStageRoute(router, "bid-lot", h.bidFactory)
	server.RegisterSingleStageRoute(router, "claim-lot", h.claimFactory)
	server.RegisterSingleStageRoute(router, "create-lot", h.createFactory)
	server.RegisterSingleStageRoute(router, "lots", h.lotsFactory)
	server.RegisterSingleStageRoute(router, "recover-lot", h.recoverFactory)
	server.RegisterSingleStageRoute(router, "status", h.statusFactory)
}

// ==============
// === Common ===
// ==============

// Context with services and common bindings for calls
type commonContext struct {
	w           *wallet.LocalWallet
	rp          *rocketpool.RocketPool
	opts        *bind.TransactOpts
	nodeAddress common.Address
}

// Create a scaffolded generic call handler, with caller-specific functionality where applicable
func runAuctionCall[dataType any](h server.ISingleStageCallContext[dataType, commonContext]) (*api.ApiResponse[dataType], error) {
	// Get services
	if err := services.RequireNodeRegistered(); err != nil {
		return nil, fmt.Errorf("error checking if node is registered: %w", err)
	}
	sp := services.GetServiceProvider()
	w := sp.GetWallet()
	rp := sp.GetRocketPool()
	address, _ := w.GetAddress()

	// Get the transact opts if this node is ready for transaction
	var opts *bind.TransactOpts
	walletStatus := w.GetStatus()
	if walletStatus == wtypes.WalletStatus_Ready {
		var err error
		opts, err = w.GetTransactor()
		if err != nil {
			return nil, fmt.Errorf("error getting node account transactor: %w", err)
		}
	}

	// Response
	data := new(dataType)
	response := &api.ApiResponse[dataType]{
		WalletStatus: walletStatus,
		Data:         data,
	}

	// Create the context
	context := &commonContext{
		w:           w,
		rp:          rp,
		opts:        opts,
		nodeAddress: address,
	}

	// Supplemental function-specific bindings
	err := h.CreateBindings(context)
	if err != nil {
		return nil, err
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		h.GetState(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Supplemental function-specific response construction
	err = h.PrepareData(data)
	if err != nil {
		return nil, err
	}

	// Return
	return response, nil
}
