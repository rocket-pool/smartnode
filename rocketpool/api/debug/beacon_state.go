package debug

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getBeaconStateForSlot(c *cli.Context, slot uint64) (*api.BeaconStateResponse, error) {
	// Create a new response
	response := api.BeaconStateResponse{}

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}

	// Get beacon state
	beaconState, err := bc.GetBeaconState(slot)
	if err != nil {
		return nil, err
	}

	fmt.Println("%w", beaconState)

	// Return response
	return &response, nil
}
