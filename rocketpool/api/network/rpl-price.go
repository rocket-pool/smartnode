package network

import (
	"math/big"
	"net/http"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/network"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getRplPrice(c *cli.Command) (*api.RplPriceResponse, error) {

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

func rplPriceHandler(c *cli.Command) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := getRplPrice(c)
		response.WriteResponse(w, resp, err)
	}
}
