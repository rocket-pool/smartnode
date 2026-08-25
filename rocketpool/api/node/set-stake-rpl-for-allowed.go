package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

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

func setStakeRplForAllowed(c *cli.Command, caller common.Address, allowed bool, t *snroute.TransactOpts) (*api.SetStakeRplForAllowedResponse, error) {
	opts := t.Opts()

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

func canSetStakeRplForAllowedHandler(ctx snroute.Context) {
	caller := common.HexToAddress(ctx.Request.URL.Query().Get("caller"))
	allowed := ctx.Request.URL.Query().Get("allowed") == "true"
	resp, err := canSetStakeRplForAllowed(ctx.Command(), caller, allowed)
	response.WriteResponse(ctx.Writer, resp, err)
}

func setStakeRplForAllowedHandler(ctx snroute.WriteContext) {
	caller := common.HexToAddress(ctx.Request.FormValue("caller"))
	allowed := ctx.Request.FormValue("allowed") == "true"
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := setStakeRplForAllowed(ctx.Command(), caller, allowed, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
