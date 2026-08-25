package node

import (
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canSetStakeRplForAllowed(c *cli.Command, caller common.Address, allowed bool) (*api.CanSetStakeRplForAllowedResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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
	response := api.CanSetStakeRplForAllowedResponse{}

	// Get gas estimates
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasLimits, err := node.EstimateSetStakeRPLForAllowedGas(rp, caller, allowed, opts)
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits

	// Update & return response
	response.CanSet = true
	return &response, nil

}

func setStakeRplForAllowed(c *cli.Command, caller common.Address, allowed bool, opts *bind.TransactOpts) (*api.SetStakeRplForAllowedResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SetStakeRplForAllowedResponse{}

	// Stake RPL
	hash, err := node.SetStakeRPLForAllowed(rp, caller, allowed, opts)
	if err != nil {
		return nil, err
	}

	response.SetTxHash = hash

	// Return response
	return &response, nil

}

func canSetStakeRplForAllowedHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := common.HexToAddress(r.URL.Query().Get("caller"))
		allowed := r.URL.Query().Get("allowed") == "true"
		resp, err := canSetStakeRplForAllowed(c, caller, allowed)
		response.WriteResponse(w, resp, err)
	}
}

func setStakeRplForAllowedHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller := common.HexToAddress(r.FormValue("caller"))
		allowed := r.FormValue("allowed") == "true"
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := setStakeRplForAllowed(c, caller, allowed, opts)
		response.WriteResponse(w, resp, err)
	}
}
