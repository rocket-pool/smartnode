package node

import (
	"net/http"
	_ "time/tzdata" // Load the embedded tz data

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"

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

func setTimezoneLocation(c *cli.Command, timezoneLocation string, opts *bind.TransactOpts) (*api.SetNodeTimezoneResponse, error) {

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

func canSetTimezoneHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tz := r.URL.Query().Get("timezoneLocation")
		resp, err := canSetTimezoneLocation(c, tz)
		response.WriteResponse(w, resp, err)
	}
}

func setTimezoneHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tz := r.FormValue("timezoneLocation")
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := setTimezoneLocation(c, tz, opts)
		response.WriteResponse(w, resp, err)
	}
}
