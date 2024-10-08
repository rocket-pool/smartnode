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

type protocolDaoIsVotingInitializedContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoIsVotingInitializedContextFactory) Create(args url.Values) (*protocolDaoIsVotingInitializedContext, error) {
	c := &protocolDaoIsVotingInitializedContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoIsVotingInitializedContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoIsVotingInitializedContext, api.ProtocolDaoIsVotingInitializedData](
		router, "is-voting-initialized", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoIsVotingInitializedContext struct {
	handler *ProtocolDaoHandler
	rp      *rocketpool.RocketPool

	node *node.Node
}

func (c *protocolDaoIsVotingInitializedContext) Initialize() (types.ResponseStatus, error) {
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

func (c *protocolDaoIsVotingInitializedContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.IsVotingInitialized,
	)
}

func (c *protocolDaoIsVotingInitializedContext) PrepareData(data *api.ProtocolDaoIsVotingInitializedData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.VotingInitialized = c.node.IsVotingInitialized.Get()

	return types.ResponseStatus_Success, nil
}
