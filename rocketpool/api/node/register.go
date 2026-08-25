package node

import (
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canRegisterNode(c *cli.Command, timezoneLocation string) (*api.CanRegisterNodeResponse, error) {

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
	response := api.CanRegisterNodeResponse{}

	// Sync
	var wg errgroup.Group

	// Check node is not already registered
	wg.Go(func() error {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return err
		}
		exists, err := node.GetNodeExists(rp, nodeAccount.Address, nil)
		if err != nil {
			return err
		}
		response.AlreadyRegistered = exists
		return nil
	})

	// Check node registrations are enabled
	wg.Go(func() error {
		registrationEnabled, err := protocol.GetNodeRegistrationEnabled(rp, nil)
		if err == nil {
			response.RegistrationDisabled = !registrationEnabled
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasLimits, err := node.EstimateRegisterNodeGas(rp, timezoneLocation, opts)
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
	response.CanRegister = !response.AlreadyRegistered && !response.RegistrationDisabled
	return &response, nil

}

func registerNode(c *cli.Command, timezoneLocation string, opts *bind.TransactOpts) (*api.RegisterNodeResponse, error) {

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
	response := api.RegisterNodeResponse{}

	// Register node
	hash, err := node.RegisterNode(rp, timezoneLocation, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canRegisterHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tz := r.URL.Query().Get("timezoneLocation")
		resp, err := canRegisterNode(c, tz)
		response.WriteResponse(w, resp, err)
	}
}

func registerHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tz := r.FormValue("timezoneLocation")
		opts, err := services.GetNodeAccountTransactorFromRequest(c, r)
		if err != nil {
			response.WriteErrorResponse(w, err)
			return
		}
		resp, err := registerNode(c, tz, opts)
		response.WriteResponse(w, resp, err)
	}
}
