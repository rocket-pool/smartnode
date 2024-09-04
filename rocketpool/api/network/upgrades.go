package network

import (
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func isHoustonHotfixDeployed(c *cli.Context) (*api.IsHoustonHotfixDeployedResponse, error) {

	// Get services
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.IsHoustonHotfixDeployedResponse{}

	isHoustonHotfixDeployed, err := state.IsHoustonHotfixDeployed(rp, nil)
	if err != nil {
		return nil, err
	}
	response.IsHoustonHotfixDeployed = isHoustonHotfixDeployed

	// Return response
	return &response, nil

}
