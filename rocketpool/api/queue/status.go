package queue

import (
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getStatus(c *cli.Context) (*api.QueueStatusResponse, error) {

	// Get services
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.QueueStatusResponse{}

	// Sync
	var wg errgroup.Group

	// Get deposit pool balance
	wg.Go(func() error {
		depositPoolBalance, err := deposit.GetBalance(rp, nil)
		if err != nil {
			return err
		}
		response.DepositPoolBalance.Set(depositPoolBalance)
		return nil
	})

	// Get minipool queue length
	wg.Go(func() error {
		var err error
		response.MinipoolQueueLength, err = minipool.GetQueueTotalLength(rp, nil)
		return err
	})

	// Get minipool queue capacity
	wg.Go(func() error {
		minipoolQueueCapacity, err := minipool.GetQueueTotalCapacity(rp, nil)
		if err != nil {
			return err
		}
		response.MinipoolQueueCapacity.Set(minipoolQueueCapacity)
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}
