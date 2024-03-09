package network

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/network"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkPriceContextFactory struct {
	handler *NetworkHandler
}

func (f *networkPriceContextFactory) Create(args url.Values) (*networkPriceContext, error) {
	c := &networkPriceContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkPriceContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*networkPriceContext, api.NetworkRplPriceData](
		router, "rpl-price", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkPriceContext struct {
	handler *NetworkHandler
	rp      *rocketpool.RocketPool

	pSettings  *protocol.ProtocolDaoSettings
	networkMgr *network.NetworkManager
}

func (c *networkPriceContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	err := sp.RequireEthClientSynced()
	if err != nil {
		return err
	}

	// Bindings
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	c.networkMgr, err = network.NewNetworkManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating network prices binding: %w", err)
	}
	return nil
}

func (c *networkPriceContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.networkMgr.PricesBlock,
		c.networkMgr.RplPrice,
		c.pSettings.Node.MinimumPerMinipoolStake,
		c.pSettings.Node.MaximumPerMinipoolStake,
	)
}

func (c *networkPriceContext) PrepareData(data *api.NetworkRplPriceData, opts *bind.TransactOpts) error {
	var rplPrice *big.Int
	_24Eth := eth.EthToWei(24)
	_16Eth := eth.EthToWei(16)
	_8Eth := eth.EthToWei(8)
	var minPerMinipoolStake *big.Int
	var maxPerMinipoolStake *big.Int

	data.RplPriceBlock = c.networkMgr.PricesBlock.Formatted()
	rplPrice = c.networkMgr.RplPrice.Raw()
	minPerMinipoolStake = c.pSettings.Node.MinimumPerMinipoolStake.Raw()
	maxPerMinipoolStake = c.pSettings.Node.MaximumPerMinipoolStake.Raw()

	// Min for LEB8s
	minPer8EthMinipoolRplStake := big.NewInt(0)
	minPer8EthMinipoolRplStake.Mul(_24Eth, minPerMinipoolStake) // Min is 10% of borrowed (24 ETH)
	minPer8EthMinipoolRplStake.Div(minPer8EthMinipoolRplStake, rplPrice)
	minPer8EthMinipoolRplStake.Add(minPer8EthMinipoolRplStake, big.NewInt(1))
	data.MinPer8EthMinipoolRplStake = minPer8EthMinipoolRplStake

	// Max for LEB8s
	maxPer8EthMinipoolRplStake := big.NewInt(0)
	maxPer8EthMinipoolRplStake.Mul(_8Eth, maxPerMinipoolStake) // Max is 150% of bonded (8 ETH)
	maxPer8EthMinipoolRplStake.Div(maxPer8EthMinipoolRplStake, rplPrice)
	maxPer8EthMinipoolRplStake.Add(maxPer8EthMinipoolRplStake, big.NewInt(1))
	data.MaxPer8EthMinipoolRplStake = maxPer8EthMinipoolRplStake

	// Min for 16s
	minPer16EthMinipoolRplStake := big.NewInt(0)
	minPer16EthMinipoolRplStake.Mul(_16Eth, minPerMinipoolStake) // Min is 10% of borrowed (16 ETH)
	minPer16EthMinipoolRplStake.Div(minPer16EthMinipoolRplStake, rplPrice)
	minPer16EthMinipoolRplStake.Add(minPer16EthMinipoolRplStake, big.NewInt(1))
	data.MinPer16EthMinipoolRplStake = minPer16EthMinipoolRplStake

	// Max for 16s
	maxPer16EthMinipoolRplStake := big.NewInt(0)
	maxPer16EthMinipoolRplStake.Mul(_16Eth, maxPerMinipoolStake) // Max is 150% of bonded (16 ETH)
	maxPer16EthMinipoolRplStake.Div(maxPer16EthMinipoolRplStake, rplPrice)
	maxPer16EthMinipoolRplStake.Add(maxPer16EthMinipoolRplStake, big.NewInt(1))
	data.MaxPer16EthMinipoolRplStake = maxPer16EthMinipoolRplStake

	// Update & return response
	data.RplPrice = rplPrice

	return nil
}
