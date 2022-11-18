package network

import (
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
	"github.com/urfave/cli"
)

func isAtlasDeployed(c *cli.Context) (*api.IsAtlasDeployedResponse, error) {

	// Get services
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.IsAtlasDeployedResponse{}

	isAtlasDeployed, err := rputils.IsAtlasDeployed(rp)
	if err != nil {
		return nil, err
	}
	response.IsAtlasDeployed = isAtlasDeployed

	// Return response
	return &response, nil

}
