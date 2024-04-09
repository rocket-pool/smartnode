package pdao

import (
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type protocolDaoInitializeVotingContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoInitializeVotingContextFactory) Create(args url.Values) (*protocolDaoInitializeVotingContext, error) {
	c := &protocolDaoInitializeVotingContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *protocolDaoInitializeVotingContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoInitializeVotingContext, api.ProtocolDaoInitializeVotingData](
		router, "initialize-voting", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoInitializeVotingContext struct {
	handler *ProtocolDaoHandler
	rp      *rocketpool.RocketPool

	node *node.Node
}

func (c *protocolDaoInitializeVotingContext) Initialize() (types.ResponseStatus, error) {
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

func (c *protocolDaoInitializeVotingContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.IsVotingInitialized,
	)
}

func (c *protocolDaoInitializeVotingContext) PrepareData(data *api.ProtocolDaoInitializeVotingData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
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
