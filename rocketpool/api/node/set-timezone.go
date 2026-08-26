package node

import (
	_ "time/tzdata" // Load the embedded tz data

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canSetTimezoneLocation(c *cli.Command, timezoneLocation string) (*api.CanSetNodeTimezoneResponse, error) {

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
	response := api.CanSetNodeTimezoneResponse{}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasLimits, err := node.EstimateSetTimezoneLocationGas(rp, timezoneLocation, opts)
	if err != nil {
		return nil, err
	}
	response.GasLimits = gasLimits
	response.CanSet = true
	return &response, nil

}

func setTimezoneLocation(c *cli.Command, timezoneLocation string, t *snroute.TransactOpts) (*api.SetNodeTimezoneResponse, error) {
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
	response := api.SetNodeTimezoneResponse{}

	// Set timezone location
	hash, err := node.SetTimezoneLocation(rp, timezoneLocation, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canSetTimezoneHandler(ctx snroute.Context) {
	tz := ctx.Request.URL.Query().Get("timezoneLocation")
	resp, err := canSetTimezoneLocation(ctx.Command(), tz)
	response.WriteResponse(ctx.Writer, resp, err)
}

func setTimezoneHandler(ctx snroute.WriteContext) {
	tz := ctx.Request.FormValue("timezoneLocation")
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := setTimezoneLocation(ctx.Command(), tz, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
