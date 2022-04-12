package network

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

func mergeUpdateStatus(c *cli.Context) (*api.NetworkMergeUpdateStatusResponse, error) {
	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NetworkMergeUpdateStatusResponse{}

	// Get merge update deployment status
	isMergeUpdateDeployed, err := rputils.IsMergeUpdateDeployed(rp)
	if err != nil {
		return nil, fmt.Errorf("error determining if merge update contracts have been deployed: %w", err)
	}
	response.IsUpdateDeployed = isMergeUpdateDeployed

	return &response, nil
}
