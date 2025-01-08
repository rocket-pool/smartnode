package megapool

import (
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func getStatus(c *cli.Context) (*api.MegapoolStatusResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
		return nil, err
	}

	// Response
	response := api.MegapoolStatusResponse{}

	response.Test = true

	return &response, nil
}
