package network

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/network"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
		router, "rpl-price", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
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

func (c *networkPriceContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	c.networkMgr, err = network.NewNetworkManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating network prices binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *networkPriceContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.networkMgr.PricesBlock,
		c.networkMgr.RplPrice,
		c.pSettings.Node.MinimumPerMinipoolStake,
	)
}

func (c *networkPriceContext) PrepareData(data *api.NetworkRplPriceData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	var rplPrice *big.Int
	_24Eth := eth.EthToWei(24)
	_16Eth := eth.EthToWei(16)
	var minPerMinipoolStake *big.Int

	data.RplPriceBlock = c.networkMgr.PricesBlock.Formatted()
	rplPrice = c.networkMgr.RplPrice.Raw()
	minPerMinipoolStake = c.pSettings.Node.MinimumPerMinipoolStake.Raw()

	// Min for LEB8s
	minPer8EthMinipoolRplStake := big.NewInt(0)
	minPer8EthMinipoolRplStake.Mul(_24Eth, minPerMinipoolStake) // Min is 10% of borrowed (24 ETH)
	minPer8EthMinipoolRplStake.Div(minPer8EthMinipoolRplStake, rplPrice)
	minPer8EthMinipoolRplStake.Add(minPer8EthMinipoolRplStake, big.NewInt(1))
	data.MinPer8EthMinipoolRplStake = minPer8EthMinipoolRplStake

	// Min for 16s
	minPer16EthMinipoolRplStake := big.NewInt(0)
	minPer16EthMinipoolRplStake.Mul(_16Eth, minPerMinipoolStake) // Min is 10% of borrowed (16 ETH)
	minPer16EthMinipoolRplStake.Div(minPer16EthMinipoolRplStake, rplPrice)
	minPer16EthMinipoolRplStake.Add(minPer16EthMinipoolRplStake, big.NewInt(1))
	data.MinPer16EthMinipoolRplStake = minPer16EthMinipoolRplStake

	// Update & return response
	data.RplPrice = rplPrice

	return types.ResponseStatus_Success, nil
}
