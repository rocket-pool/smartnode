package network

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type networkCurrentVotingDelegateContextFactory struct {
	handler *NetworkHandler
}

func (f *networkCurrentVotingDelegateContextFactory) Create(args url.Values) (*networkCurrentVotingDelegateContext, error) {
	c := &networkCurrentVotingDelegateContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *networkCurrentVotingDelegateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*networkCurrentVotingDelegateContext, api.NetworkCurrentVotingDelegateData](
		router, "voting-delegate", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type networkCurrentVotingDelegateContext struct {
	handler *NetworkHandler
	rp      *rocketpool.RocketPool

	node *node.Node
}

func (c *networkCurrentVotingDelegateContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	return nil
}

func (c *networkCurrentVotingDelegateContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.CurrentVotingDelegate,
	)
}

func (c *networkCurrentVotingDelegateContext) PrepareData(data *api.NetworkCurrentVotingDelegateData, opts *bind.TransactOpts) error {
	data.AccountAddress = c.node.Address
	data.VotingDelegate = c.node.CurrentVotingDelegate.Get()
	return nil
}
