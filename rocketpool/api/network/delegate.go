package network

import (
	"github.com/urfave/cli/v3"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get the latest delegate contract address
func getLatestDelegate(c *cli.Command) (*api.GetLatestDelegateResponse, error) {

	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetLatestDelegateResponse{}

	// Get latest delegate address
	latestDelegateAddress, err := rp.GetAddress("rocketMinipoolDelegate", nil)
	if err != nil {
		return nil, err
	}
	response.Address = *latestDelegateAddress

	// Return response
	return &response, nil

}
