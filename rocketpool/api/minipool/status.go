package minipool

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	rputils "github.com/rocket-pool/smartnode/shared/utils/rp"
)

func getStatus(c *cli.Context) (*api.MinipoolStatusResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	if err := services.RequireBeaconClientSynced(c); err != nil {
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
	bc, err := services.GetBeaconClient(c)
	if err != nil {
		return nil, err
	}
	cfg, err := services.GetConfig(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.MinipoolStatusResponse{}

	// Check if Atlas is deployed
	response.IsAtlasDeployed, err = rputils.IsAtlasDeployed(rp)
	if err != nil {
		return nil, fmt.Errorf("error checking if Atlas has been deployed: %w", err)
	}

	// Get the legacy MinipoolQueue contract address
	legacyMinipoolQueueAddress := cfg.Smartnode.GetV110MinipoolQueueAddress()

	// Get minipool details
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}
	details, err := getNodeMinipoolDetails(rp, bc, nodeAccount.Address, response.IsAtlasDeployed, &legacyMinipoolQueueAddress)
	if err != nil {
		return nil, err
	}
	response.Minipools = details

	delegate, err := rp.GetContract("rocketMinipoolDelegate", nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting latest minipool delegate contract: %w", err)
	}

	response.LatestDelegate = *delegate.Address

	// Return response
	return &response, nil

}
