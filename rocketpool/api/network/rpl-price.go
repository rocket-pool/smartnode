package network

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getRplPrice(c *cli.Context) (*api.RplPriceResponse, error) {

	// Get services
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.RplPriceResponse{}

	// Data
	var wg errgroup.Group
	var rplPrice *big.Int

	// Get RPL price set block
	wg.Go(func() error {
		pricesBlock, err := network.GetPricesBlock(rp, nil)
		if err == nil {
			response.RplPriceBlock = pricesBlock
		}
		return err
	})

	// Get data
	wg.Go(func() error {
		var err error
		rplPrice, err = network.GetRPLPrice(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.RplPrice = rplPrice
	return &response, nil

}
