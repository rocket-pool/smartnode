package pdao

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

type protocolDaoCurrentVotingDelegateContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoCurrentVotingDelegateContextFactory) Create(args url.Values) (*protocolDaoCurrentVotingDelegateContext, error) {
	c := &protocolDaoCurrentVotingDelegateContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoCurrentVotingDelegateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoCurrentVotingDelegateContext, api.ProtocolDaoCurrentVotingDelegateData](
		router, "voting-delegate", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoCurrentVotingDelegateContext struct {
	handler *ProtocolDaoHandler
	rp      *rocketpool.RocketPool

	node *node.Node
}

func (c *protocolDaoCurrentVotingDelegateContext) Initialize() (types.ResponseStatus, error) {
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

func (c *protocolDaoCurrentVotingDelegateContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.CurrentVotingDelegate,
	)
}

func (c *protocolDaoCurrentVotingDelegateContext) PrepareData(data *api.ProtocolDaoCurrentVotingDelegateData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.AccountAddress = c.node.Address
	data.VotingDelegate = c.node.CurrentVotingDelegate.Get()
	return types.ResponseStatus_Success, nil
}
