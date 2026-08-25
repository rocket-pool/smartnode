package node

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canSetRplLockAllowed(c *cli.Command, allowed bool) (*api.CanSetRplLockingAllowedResponse, error) {

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
	// Get the node account
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanSetRplLockingAllowedResponse{}

	isAllowed, err := node.GetRPLLockedAllowed(rp, account.Address, nil)
	if err != nil {
		return nil, err
	}

	// Get gas estimates
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasLimits, err := node.EstimateSetStakeRPLForAllowedGas(rp, account.Address, allowed, opts)
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits

	// Update & return response
	response.CanSet = (!isAllowed && allowed) || (isAllowed && !allowed)
	if !response.CanSet {
		if allowed {
			response.Error = "RPL locking is already allowed"
		} else {
			response.Error = "RPL locking is already denied"
		}
	}
	return &response, nil

}

func setRplLockAllowed(c *cli.Command, allowed bool, t *snroute.TransactOpts) (*api.SetRplLockingAllowedResponse, error) {
	opts := t.Opts()

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
	// Get the node account
	account, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SetRplLockingAllowedResponse{}

	// Stake RPL
	hash, err := node.SetRPLLockingAllowed(rp, account.Address, allowed, opts)
	if err != nil {
		return nil, err
	}

	response.SetTxHash = hash

	// Return response
	return &response, nil

}

func canSetRplLockingAllowedHandler(ctx snroute.Context) {
	allowed := ctx.Request.URL.Query().Get("allowed") == "true"
	resp, err := canSetRplLockAllowed(ctx.Command(), allowed)
	response.WriteResponse(ctx.Writer, resp, err)
}

func setRplLockingAllowedHandler(ctx snroute.WriteContext) {
	allowed := ctx.Request.FormValue("allowed") == "true"
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := setRplLockAllowed(ctx.Command(), allowed, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
