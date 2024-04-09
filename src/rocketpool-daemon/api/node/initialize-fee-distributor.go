package node

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeInitializeFeeDistributorContextFactory struct {
	handler *NodeHandler
}

func (f *nodeInitializeFeeDistributorContextFactory) Create(args url.Values) (*nodeInitializeFeeDistributorContext, error) {
	c := &nodeInitializeFeeDistributorContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeInitializeFeeDistributorContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeInitializeFeeDistributorContext, api.NodeInitializeFeeDistributorData](
		router, "initialize-fee-distributor", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeInitializeFeeDistributorContext struct {
	handler *NodeHandler
	rp      *rocketpool.RocketPool

	node *node.Node
}

func (c *nodeInitializeFeeDistributorContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *nodeInitializeFeeDistributorContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.IsFeeDistributorInitialized,
		c.node.DistributorAddress,
	)
}

func (c *nodeInitializeFeeDistributorContext) PrepareData(data *api.NodeInitializeFeeDistributorData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.Distributor = c.node.DistributorAddress.Get()
	data.IsInitialized = c.node.IsFeeDistributorInitialized.Get()
	data.CanInitialize = !(data.IsInitialized)

	// Get tx info
	if data.CanInitialize {
		txInfo, err := c.node.InitializeFeeDistributor(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for InitializeFeeDistributor: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
