package auction

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool/api/handlers"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/services"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
	wtypes "github.com/rocket-pool/smartnode/shared/types/wallet"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Context with services and common bindings for calls
type callContext struct {
	w           *wallet.LocalWallet
	rp          *rocketpool.RocketPool
	opts        *bind.TransactOpts
	nodeAddress common.Address
}

// Register routes
func RegisterRoutes(router *mux.Router, name string) {
	route := "auction"

	// Bid
	server.RegisterSingleStageHandler[api.BidOnLotData](router, route, "bid-lot", []func(*auctionBidHandler, map[string]string) error{
		func(h *auctionBidHandler, vars map[string]string) error {
			return server.ValidateArg("index", vars, cliutils.ValidateUint, &h.lotIndex)
		},
		func(h *auctionBidHandler, vars map[string]string) error {
			return server.ValidateArg("amount", vars, cliutils.ValidatePositiveWeiAmount, &h.amountWei)
		},
	}, runAuctionCall[api.BidOnLotData])

	// Bid
	router.HandleFunc(fmt.Sprintf("/%s/bid-lot", route), func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		lotIndex, err := server.ValidateArg("index", vars, cliutils.ValidateUint)
		if err != nil {
			server.HandleInputError(w, err)
			return
		}
		amountWei, err := server.ValidateArg("amount", vars, cliutils.ValidatePositiveWeiAmount)
		if err != nil {
			server.HandleInputError(w, err)
			return
		}

		response, err := runAuctionCall[api.BidOnLotData](&auctionBidHandler{
			lotIndex:  lotIndex,
			amountWei: amountWei,
		})
		server.HandleResponse(w, response, err)
	})

	// Claim
	router.HandleFunc(fmt.Sprintf("/%s/claim-lot", route), func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		lotIndex, err := server.ValidateArg("index", vars, cliutils.ValidateUint)
		if err != nil {
			server.HandleInputError(w, err)
			return
		}

		response, err := runAuctionCall[api.ClaimFromLotData](&auctionClaimHandler{
			lotIndex: lotIndex,
		})
		server.HandleResponse(w, response, err)
	})

	// Create
	router.HandleFunc(fmt.Sprintf("/%s/create-lot", route), func(w http.ResponseWriter, r *http.Request) {
		response, err := runAuctionCall[api.CreateLotData](&auctionCreateHandler{})
		server.HandleResponse(w, response, err)
	})

	// Lots
	router.HandleFunc(fmt.Sprintf("/%s/lots", route), func(w http.ResponseWriter, r *http.Request) {
		response, err := runAuctionCall[api.AuctionLotsData](&auctionLotHandler{})
		server.HandleResponse(w, response, err)
	})

	// Recover Lot
	router.HandleFunc(fmt.Sprintf("/%s/recover-lot", route), func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		lotIndex, err := server.ValidateArg("index", vars, cliutils.ValidateUint)
		if err != nil {
			server.HandleInputError(w, err)
			return
		}

		response, err := runAuctionCall[api.RecoverRplFromLotData](&auctionRecoverHandler{
			lotIndex: lotIndex,
		})
		server.HandleResponse(w, response, err)
	})

	// Status
	router.HandleFunc(fmt.Sprintf("/%s/status", route), func(w http.ResponseWriter, r *http.Request) {
		response, err := runAuctionCall[api.AuctionStatusData](&auctionStatusHandler{})
		server.HandleResponse(w, response, err)
	})

}

// Create a scaffolded generic call handler, with caller-specific functionality where applicable
func runAuctionCall[dataType any, implType any](h handlers.ISingleStageCallHandler[dataType, callContext, implType]) (*api.ApiResponse[dataType], error) {
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
	context := &callContext{
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
		h.GetState(context, mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Supplemental function-specific response construction
	err = h.PrepareData(context, data)
	if err != nil {
		return nil, err
	}

	// Return
	return response, nil
}
