package pdao

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getVotePower(c *cli.Context) (*api.GetPDAOVotePowerResponse, error) {

	// Get services
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetPDAOVotePowerResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Get current block number
	blockNumber, err := ec.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Error getting block number: %w", err)
	}

	// Cast to uint32
	blockNumber32 := uint32(blockNumber)

	// Check voting power
	response.VotingPower, err = network.GetVotingPower(rp, nodeAccount.Address, blockNumber32, nil)
	if err != nil {
		return nil, err
	}

	// Update & return response
	return &response, nil
}
