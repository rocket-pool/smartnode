package node

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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

type nodeDistributeContextFactory struct {
	handler *NodeHandler
}

func (f *nodeDistributeContextFactory) Create(args url.Values) (*nodeDistributeContext, error) {
	c := &nodeDistributeContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeDistributeContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeDistributeContext, api.NodeDistributeData](
		router, "distribute", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeDistributeContext struct {
	handler *NodeHandler
	rp      *rocketpool.RocketPool

	node *node.Node
}

func (c *nodeDistributeContext) Initialize() (types.ResponseStatus, error) {
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

func (c *nodeDistributeContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.IsFeeDistributorInitialized,
		c.node.DistributorAddress,
	)
}

func (c *nodeDistributeContext) PrepareData(data *api.NodeDistributeData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Make sure it's initialized
	data.IsInitialized = c.node.IsFeeDistributorInitialized.Get()

	// Create the distributor
	distributorAddress := c.node.DistributorAddress.Get()
	distributor, err := node.NewNodeDistributor(c.rp, c.node.Address, distributorAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node distributor binding: %w", err)
	}

	// Get its balance
	data.Balance, err = c.rp.Client.BalanceAt(context.Background(), distributorAddress, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting fee distributor balance: %w", err)
	}
	data.NoBalance = (data.Balance.Cmp(common.Big0) == 0)

	data.CanDistribute = data.IsInitialized && !data.NoBalance

	if data.CanDistribute {
		// Get the node share of the balance
		err = c.rp.Query(func(mc *batch.MultiCaller) error {
			eth.AddQueryablesToMulticall(mc,
				distributor.NodeShare,
			)
			return nil
		}, nil)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting node share for distributor %s: %w", distributorAddress.Hex(), err)
		}
		data.NodeShare = distributor.NodeShare.Get()

		// Get tx info
		txInfo, err := distributor.Distribute(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for Distribute: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
