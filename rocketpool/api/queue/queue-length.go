package queue

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
)

func getTotalQueueLength(c *cli.Context) (*api.GetQueueLengthResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetQueueLengthResponse{}

	// Get data
	totalLength, err := deposit.GetTotalQueueLength(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting total queue length: %w", err)
	}

	//Return response
	response.Length = totalLength
	return &response, nil

}

func getExpressQueueLength(c *cli.Context) (*api.GetQueueLengthResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetQueueLengthResponse{}

	// Get data
	totalLength, err := deposit.GetExpressQueueLength(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting express queue length: %w", err)
	}

	//Return response
	response.Length = totalLength
	return &response, nil

}

func getStandardQueueLength(c *cli.Context) (*api.GetQueueLengthResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetQueueLengthResponse{}

	// Get data
	totalLength, err := deposit.GetStandardQueueLength(rp, nil)
	if err != nil {
		return nil, fmt.Errorf("Error getting standard queue length: %w", err)
	}

	//Return response
	response.Length = totalLength
	return &response, nil

}
