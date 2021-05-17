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
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.RplPriceResponse{}

    // Data
    var wg errgroup.Group
    var rplPrice *big.Int
    var minipoolUserAmount *big.Int
    var minPerMinipoolStake float64
    var maxPerMinipoolStake float64

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
        minipoolUserAmount, err = protocol.GetMinipoolHalfDepositUserAmount(rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        minPerMinipoolStake, err = protocol.GetMinimumPerMinipoolStake(rp, nil)
        return err
    })
    wg.Go(func() error {
        var err error
        maxPerMinipoolStake, err = protocol.GetMaximumPerMinipoolStake(rp, nil)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Calculate min & max per minipool stake amounts
    var tmp big.Int
    var minPerMinipoolRplStake big.Int
    var maxPerMinipoolRplStake big.Int
    tmp.Mul(minipoolUserAmount, eth.EthToWei(minPerMinipoolStake))
    minPerMinipoolRplStake.Quo(&tmp, rplPrice)
    minPerMinipoolRplStake.Add(&minPerMinipoolRplStake, big.NewInt(1))
    tmp.Mul(minipoolUserAmount, eth.EthToWei(maxPerMinipoolStake))
    maxPerMinipoolRplStake.Quo(&tmp, rplPrice)
    maxPerMinipoolRplStake.Add(&maxPerMinipoolRplStake, big.NewInt(1))

    // Update & return response
    response.RplPrice = rplPrice
    response.MinPerMinipoolRplStake = &minPerMinipoolRplStake
    response.MaxPerMinipoolRplStake = &maxPerMinipoolRplStake
    return &response, nil

}

