package pdao

import (
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type protocolDaoInitializeVotingWithDelegateContextFactory struct {
	handler *ProtocolDaoHandler
}

func (f *protocolDaoInitializeVotingWithDelegateContextFactory) Create(args url.Values) (*protocolDaoInitializeVotingWithDelegateContext, error) {
	c := &protocolDaoInitializeVotingWithDelegateContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("delegate", args, input.ValidateAddress, &c.delegate),
	}
	return c, errors.Join(inputErrs...)
}

func (f *protocolDaoInitializeVotingWithDelegateContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*protocolDaoInitializeVotingWithDelegateContext, api.ProtocolDaoInitializeVotingData](
		router, "initialize-voting-with-delegate", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type protocolDaoInitializeVotingWithDelegateContext struct {
	handler *ProtocolDaoHandler
	rp      *rocketpool.RocketPool

	delegate common.Address
	node     *node.Node
}

func (c *protocolDaoInitializeVotingWithDelegateContext) Initialize() (types.ResponseStatus, error) {
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

func (c *protocolDaoInitializeVotingWithDelegateContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.IsVotingInitialized,
	)
}

func (c *protocolDaoInitializeVotingWithDelegateContext) PrepareData(data *api.ProtocolDaoInitializeVotingData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.VotingInitialized = c.node.IsVotingInitialized.Get()
	data.CanInitialize = !(data.VotingInitialized)

	// Get TX info
	if data.CanInitialize {
		txInfo, err := c.node.InitializeVotingWithDelegate(c.delegate, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for InitializeVotingWithDelegate: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
