package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeSetRplLockingAllowedContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSetRplLockingAllowedContextFactory) Create(args url.Values) (*nodeSetRplLockingAllowedContext, error) {
	c := &nodeSetRplLockingAllowedContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("allowed", args, input.ValidateBool, &c.allowed),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSetRplLockingAllowedContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeSetRplLockingAllowedContext, api.NodeSetRplLockingAllowedData](
		router, "set-rpl-locking-allowed", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSetRplLockingAllowedContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	allowed bool
	node    *node.Node
}

func (c *nodeSetRplLockingAllowedContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", c.nodeAddress.Hex(), err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *nodeSetRplLockingAllowedContext) GetState(mc *batch.MultiCaller) {
	c.node.RplWithdrawalAddress.AddToQuery(mc)
}

func (c *nodeSetRplLockingAllowedContext) PrepareData(data *api.NodeSetRplLockingAllowedData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.DifferentRplAddress = (c.node.RplWithdrawalAddress.Get() != c.nodeAddress)
	data.CanSet = !(data.DifferentRplAddress)

	if data.CanSet {
		txInfo, err := c.node.SetRplLockingAllowed(c.allowed, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for SetRplLockingAllowed: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
