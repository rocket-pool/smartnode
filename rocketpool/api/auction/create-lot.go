package auction

import (
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/auction"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canCreateLot(c *cli.Command) (*api.CanCreateLotResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanCreateLotResponse{}

	// Sync
	var wg errgroup.Group

	// Check if sufficient remaining RPL is available to create a lot
	wg.Go(func() error {
		sufficientRemainingRplForLot, err := getSufficientRemainingRPLForLot(rp)
		if err == nil {
			response.InsufficientBalance = !sufficientRemainingRplForLot
		}
		return err
	})

	// Check if lot creation is enabled
	wg.Go(func() error {
		createLotEnabled, err := protocol.GetCreateLotEnabled(rp, nil)
		if err == nil {
			response.CreateLotDisabled = !createLotEnabled
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasLimits, err := auction.EstimateCreateLotGas(rp, opts)
		if err == nil {
			response.GasLimits = gasLimits
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.CanCreate = !response.InsufficientBalance && !response.CreateLotDisabled
	return &response, nil

}

func createLot(c *cli.Command, opts *bind.TransactOpts) (*api.CreateLotResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CreateLotResponse{}

	// Create lot
	lotIndex, hash, err := auction.CreateLot(rp, opts)
	if err != nil {
		return nil, err
	}
	response.LotId = lotIndex
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canCreateLotHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := canCreateLot(c)
		response.WriteResponse(w, resp, err)
	}
}

func createLotHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := createLot(c, opts)
		response.WriteResponse(w, resp, err)
	}
}
