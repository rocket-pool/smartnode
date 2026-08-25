package minipool

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canDissolveMinipool(c *cli.Command, minipoolAddress common.Address) (*api.CanDissolveMinipoolResponse, error) {

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
	response := api.CanDissolveMinipoolResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Validate minipool owner
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	if err := validateMinipoolOwner(mp, nodeAccount.Address); err != nil {
		return nil, err
	}

	// Check minipool status
	status, err := mp.GetStatus(nil)
	if err != nil {
		return nil, err
	}
	response.InvalidStatus = status != types.Initialized && status != types.Prelaunch

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasLimits, err := mp.EstimateDissolveGas(opts)
	if err == nil {
		response.GasLimits = gasLimits
	}

	// Update & return response
	response.CanDissolve = !response.InvalidStatus
	return &response, nil

}

func dissolveMinipool(c *cli.Command, minipoolAddress common.Address, t *snroute.TransactOpts) (*api.DissolveMinipoolResponse, error) {
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
	response := api.DissolveMinipoolResponse{}

	// Create minipool
	mp, err := minipool.NewMinipool(rp, minipoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Dissolve
	hash, err := mp.Dissolve(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canDissolveHandler(ctx snroute.Context) {
	addr, err := parseAddress(ctx.Request, "address")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canDissolveMinipool(ctx.Command(), addr)
	response.WriteResponse(ctx.Writer, resp, err)
}

func dissolveHandler(ctx snroute.WriteContext) {
	addr, err := parseAddress(ctx.Request, "address")
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := dissolveMinipool(ctx.Command(), addr, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
