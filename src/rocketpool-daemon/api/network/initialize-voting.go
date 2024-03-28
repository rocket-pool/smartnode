package network

import (
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkInitializeVotingContextFactory struct {
	handler *NetworkHandler
}

func (f *networkInitializeVotingContextFactory) Create(args url.Values) (*networkInitializeVotingContext, error) {
	c := &networkInitializeVotingContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkInitializeVotingContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*networkInitializeVotingContext, api.NetworkInitializeVotingData](
		router, "initialize-voting", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkInitializeVotingContext struct {
	handler *NetworkHandler
	rp      *rocketpool.RocketPool

	node *node.Node
}

func (c *networkInitializeVotingContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered()
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

func (c *networkInitializeVotingContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.IsVotingInitialized,
	)
}

func (c *networkInitializeVotingContext) PrepareData(data *api.NetworkInitializeVotingData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.VotingInitialized = c.node.IsVotingInitialized.Get()
	data.CanInitialize = !(data.VotingInitialized)

	// Get TX info
	if data.CanInitialize {
		txInfo, err := c.node.InitializeVoting(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for InitializeVoting: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
