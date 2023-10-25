package network

import (
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func isHoustonDeployed(c *cli.Context) (*api.IsHoustonDeployedResponse, error) {

	// Get services
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.IsHoustonDeployedResponse{}

	isHoustonDeployed, err := state.IsHoustonDeployed(rp, nil)
	if err != nil {
		return nil, err
	}
	response.IsHoustonDeployed = isHoustonDeployed

	// Return response
	return &response, nil

}
