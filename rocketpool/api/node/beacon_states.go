package node

import (
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getBeaconStates(c *cli.Context, slot int64) (*api.GetBeaconStatesResponse, error) {
	// Create a new response
	response := api.GetBeaconStatesResponse{}

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Get beacon states
	beaconStates, err := bc.GetBeaconStates(slot)
	if err != nil {
		return nil, err
	}

	// Set response
	response.BeaconStates = beaconStates

	// Return response
	return &response, nil
}
