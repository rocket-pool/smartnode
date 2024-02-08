package network

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
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
	_24Eth := eth.EthToWei(24)
	_16Eth := eth.EthToWei(16)
	var minPerMinipoolStake *big.Int

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
	wg.Go(func() error {
		var err error
		minPerMinipoolStake, err = protocol.GetMinimumPerMinipoolStakeRaw(rp, nil)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Min for LEB8s
	minPer8EthMinipoolRplStake := big.NewInt(0)
	minPer8EthMinipoolRplStake.Mul(_24Eth, minPerMinipoolStake) // Min is 10% of borrowed (24 ETH)
	minPer8EthMinipoolRplStake.Div(minPer8EthMinipoolRplStake, rplPrice)
	minPer8EthMinipoolRplStake.Add(minPer8EthMinipoolRplStake, big.NewInt(1))
	response.MinPer8EthMinipoolRplStake = minPer8EthMinipoolRplStake

	// Min for 16s
	minPer16EthMinipoolRplStake := big.NewInt(0)
	minPer16EthMinipoolRplStake.Mul(_16Eth, minPerMinipoolStake) // Min is 10% of borrowed (16 ETH)
	minPer16EthMinipoolRplStake.Div(minPer16EthMinipoolRplStake, rplPrice)
	minPer16EthMinipoolRplStake.Add(minPer16EthMinipoolRplStake, big.NewInt(1))
	response.MinPer16EthMinipoolRplStake = minPer16EthMinipoolRplStake

	// Update & return response
	response.RplPrice = rplPrice
	return &response, nil

}
