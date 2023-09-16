package faucet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/common/contracts"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	wtypes "github.com/rocket-pool/smartnode/shared/types/wallet"
)

// ===============
// === Handler ===
// ===============

type FaucetHandler struct {
	serviceProvider *services.ServiceProvider
	statusFactory   server.IContextFactory[*faucetStatusContext, api.FaucetStatusData, commonContext]
	withdrawFactory server.IContextFactory[*faucetWithdrawContext, api.FaucetWithdrawRplData, commonContext]
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
	server.RegisterSingleStageRoute(router, "status", h.statusFactory)
	server.RegisterSingleStageRoute(router, "withdraw-rpl", h.withdrawFactory)
}

// ==============
// === Common ===
// ==============

// Context with services and common bindings for calls
type commonContext struct {
	w           *wallet.LocalWallet
	rp          *rocketpool.RocketPool
	f           *contracts.RplFaucet
	opts        *bind.TransactOpts
	nodeAddress common.Address
}

// Create a scaffolded generic call handler, with caller-specific functionality where applicable
func runFaucetCall[dataType any](h server.ISingleStageCallContext[dataType, commonContext]) (*api.ApiResponse[dataType], error) {
	// Get services
	if err := services.RequireNodeRegistered(); err != nil {
		return nil, fmt.Errorf("error checking if node is registered: %w", err)
	}
	sp := services.GetServiceProvider()
	w := sp.GetWallet()
	rp := sp.GetRocketPool()
	f := sp.GetRplFaucet()
	address, _ := w.GetAddress()

	// Make sure the faucet is available
	if f == nil {
		return nil, fmt.Errorf("the RPL faucet is not present on this netowrk")
	}

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
		f:           f,
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
