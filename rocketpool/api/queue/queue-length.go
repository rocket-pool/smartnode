package queue

import (
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

func getQueueDetails(c *cli.Context) (*api.GetQueueDetailsResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.GetQueueDetailsResponse{}

	// Get data
	var wg errgroup.Group
	// Get the express ticket count

	//
	wg.Go(func() error {
		response.ExpressLength, err = deposit.GetExpressQueueLength(rp, nil)
		return err
	})

	//
	wg.Go(func() error {
		response.StandardLength, err = deposit.GetStandardQueueLength(rp, nil)
		return err
	})

	wg.Go(func() error {
		response.TotalLength, err = deposit.GetTotalQueueLength(rp, nil)
		return err
	})

	wg.Go(func() error {
		response.ExpressRate, err = protocol.GetExpressQueueRate(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}
	return &response, nil

}
