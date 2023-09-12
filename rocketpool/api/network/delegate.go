package network

import (
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

// Get the latest delegate contract address
func getLatestDelegate(c *cli.Context) (*api.GetLatestDelegateResponse, error) {
	// Get services
	if err := services.RequireEthClientSynced(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetLatestDelegateResponse{}

	// Get latest delegate address
	delegateContract, err := rp.GetContract(rocketpool.ContractName_RocketMinipoolDelegate)
	if err != nil {
		return nil, err
	}
	response.Address = *delegateContract.Address

	// Return response
	return &response, nil
}
