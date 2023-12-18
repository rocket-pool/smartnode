package node

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type nodeSetRplLockingAllowedContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSetRplLockingAllowedContextFactory) Create(vars map[string]string) (*nodeSetRplLockingAllowedContext, error) {
	c := &nodeSetRplLockingAllowedContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("allowed", vars, input.ValidateBool, &c.allowed),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSetRplLockingAllowedContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeSetRplLockingAllowedContext, api.NodeSetRplLockingAllowedData](
		router, "set-rpl-locking-allowed", f, f.handler.serviceProvider,
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

func (c *nodeSetRplLockingAllowedContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating node %s binding: %w", c.nodeAddress.Hex(), err)
	}
	return nil
}

func (c *nodeSetRplLockingAllowedContext) GetState(mc *batch.MultiCaller) {
	c.node.RplWithdrawalAddress.AddToQuery(mc)
}

func (c *nodeSetRplLockingAllowedContext) PrepareData(data *api.NodeSetRplLockingAllowedData, opts *bind.TransactOpts) error {
	data.DifferentRplAddress = (c.node.RplWithdrawalAddress.Get() != c.nodeAddress)
	data.CanSet = !(data.DifferentRplAddress)

	if data.CanSet {
		txInfo, err := c.node.SetRplLockingAllowed(c.allowed, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for SetRplLockingAllowed: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
