package node

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
	"github.com/urfave/cli"
)

func getSmoothingPoolRegistrationStatus(c *cli.Context) (*api.GetSmoothingPoolRegistrationStatusResponse, error) {

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
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetSmoothingPoolRegistrationStatusResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Check registration status
	response.NodeRegistered, err = node.GetSmoothingPoolRegistrationState(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}

func canSetSmoothingPoolStatus(c *cli.Context, status bool) (*api.CanSetSmoothingPoolRegistrationStatusResponse, error) {

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
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanSetSmoothingPoolRegistrationStatusResponse{}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := node.EstimateSetSmoothingPoolRegistrationStateGas(rp, status, opts)
	if err == nil {
		response.GasInfo = gasInfo
	}

	return &response, nil

}

func setSmoothingPoolStatus(c *cli.Context, status bool) (*api.SetSmoothingPoolRegistrationStatusResponse, error) {

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
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.SetSmoothingPoolRegistrationStatusResponse{}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Set the registration status
	hash, err := node.SetSmoothingPoolRegistrationState(rp, status, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
