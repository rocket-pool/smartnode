package network

import (
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/settings"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

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
	var rplPrice *big.Int
	_24Eth := eth.EthToWei(24)
	_16Eth := eth.EthToWei(16)
	_8Eth := eth.EthToWei(8)
	var minPerMinipoolStake *big.Int
	var maxPerMinipoolStake *big.Int

	// Create bindings
	network, err := network.NewNetworkPrices(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating network prices binding: %w", err)
	}
	pSettings, err := settings.NewProtocolDaoSettings(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating protocol DAO settings binding: %w", err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		network.GetPricesBlock(mc)
		network.GetRplPrice(mc)
		pSettings.GetMinimumPerMinipoolStake(mc)
		pSettings.GetMaximumPerMinipoolStake(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Get the details
	response.RplPriceBlock = network.Details.PricesBlock.Formatted()
	rplPrice = network.Details.RplPrice.RawValue
	minPerMinipoolStake = pSettings.Details.Node.MinimumPerMinipoolStake.RawValue
	maxPerMinipoolStake = pSettings.Details.Node.MaximumPerMinipoolStake.RawValue

	// Min for LEB8s
	minPer8EthMinipoolRplStake := big.NewInt(0)
	minPer8EthMinipoolRplStake.Mul(_24Eth, minPerMinipoolStake) // Min is 10% of borrowed (24 ETH)
	minPer8EthMinipoolRplStake.Div(minPer8EthMinipoolRplStake, rplPrice)
	minPer8EthMinipoolRplStake.Add(minPer8EthMinipoolRplStake, big.NewInt(1))
	response.MinPer8EthMinipoolRplStake = minPer8EthMinipoolRplStake

	// Max for LEB8s
	maxPer8EthMinipoolRplStake := big.NewInt(0)
	maxPer8EthMinipoolRplStake.Mul(_8Eth, maxPerMinipoolStake) // Max is 150% of bonded (8 ETH)
	maxPer8EthMinipoolRplStake.Div(maxPer8EthMinipoolRplStake, rplPrice)
	maxPer8EthMinipoolRplStake.Add(maxPer8EthMinipoolRplStake, big.NewInt(1))
	response.MaxPer8EthMinipoolRplStake = maxPer8EthMinipoolRplStake

	// Min for 16s
	minPer16EthMinipoolRplStake := big.NewInt(0)
	minPer16EthMinipoolRplStake.Mul(_16Eth, minPerMinipoolStake) // Min is 10% of borrowed (16 ETH)
	minPer16EthMinipoolRplStake.Div(minPer16EthMinipoolRplStake, rplPrice)
	minPer16EthMinipoolRplStake.Add(minPer16EthMinipoolRplStake, big.NewInt(1))
	response.MinPer16EthMinipoolRplStake = minPer16EthMinipoolRplStake

	// Max for 16s
	maxPer16EthMinipoolRplStake := big.NewInt(0)
	maxPer16EthMinipoolRplStake.Mul(_16Eth, maxPerMinipoolStake) // Max is 150% of bonded (16 ETH)
	maxPer16EthMinipoolRplStake.Div(maxPer16EthMinipoolRplStake, rplPrice)
	maxPer16EthMinipoolRplStake.Add(maxPer16EthMinipoolRplStake, big.NewInt(1))
	response.MaxPer16EthMinipoolRplStake = maxPer16EthMinipoolRplStake

	// Update & return response
	response.RplPrice = rplPrice
	return &response, nil
}
