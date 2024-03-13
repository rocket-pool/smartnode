package node

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeConfirmRplWithdrawalAddressContextFactory struct {
	handler *NodeHandler
}

func (f *nodeConfirmRplWithdrawalAddressContextFactory) Create(args url.Values) (*nodeConfirmRplWithdrawalAddressContext, error) {
	c := &nodeConfirmRplWithdrawalAddressContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeConfirmRplWithdrawalAddressContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeConfirmRplWithdrawalAddressContext, api.NodeConfirmRplWithdrawalAddressData](
		router, "rpl-withdrawal-address/confirm", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeConfirmRplWithdrawalAddressContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	node *node.Node
}

func (c *nodeConfirmRplWithdrawalAddressContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered(c.handler.context)
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

func (c *nodeConfirmRplWithdrawalAddressContext) GetState(mc *batch.MultiCaller) {
	c.node.PendingRplWithdrawalAddress.AddToQuery(mc)
}

func (c *nodeConfirmRplWithdrawalAddressContext) PrepareData(data *api.NodeConfirmRplWithdrawalAddressData, opts *bind.TransactOpts) error {
	data.IncorrectPendingAddress = (c.node.PendingRplWithdrawalAddress.Get() != c.nodeAddress)
	data.CanConfirm = !(data.IncorrectPendingAddress)

	if data.CanConfirm {
		txInfo, err := c.node.ConfirmRplWithdrawalAddress(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ConfirmRplWithdrawalAddress: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
