package network

import (
	"math/big"

	"github.com/rocket-pool/rocketpool-go/network"
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

	// RPL stake amounts for 5,10,15% borrowed ETH per LEB8
	fivePercentBorrowedPerMinipool := new(big.Int)
	fivePercentBorrowedPerMinipool.SetString("50000000000000000", 10)

	fivePercentBorrowedRplStake := big.NewInt(0)
	fivePercentBorrowedRplStake.Mul(_24Eth, fivePercentBorrowedPerMinipool)
	fivePercentBorrowedRplStake.Div(fivePercentBorrowedRplStake, rplPrice)
	fivePercentBorrowedRplStake.Add(fivePercentBorrowedRplStake, big.NewInt(1))
	response.FivePercentBorrowedRplStake = fivePercentBorrowedRplStake
	response.TenPercentBorrowedRplStake = new(big.Int).Mul(fivePercentBorrowedRplStake, big.NewInt(2))
	response.FifteenPercentBorrowedRplStake = new(big.Int).Mul(fivePercentBorrowedRplStake, big.NewInt(3))

	// Update & return response
	response.RplPrice = rplPrice
	return &response, nil

}
