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

type nodeConfirmPrimaryWithdrawalAddressContextFactory struct {
	handler *NodeHandler
}

func (f *nodeConfirmPrimaryWithdrawalAddressContextFactory) Create(args url.Values) (*nodeConfirmPrimaryWithdrawalAddressContext, error) {
	c := &nodeConfirmPrimaryWithdrawalAddressContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeConfirmPrimaryWithdrawalAddressContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeConfirmPrimaryWithdrawalAddressContext, api.NodeConfirmPrimaryWithdrawalAddressData](
		router, "primary-withdrawal-address/confirm", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeConfirmPrimaryWithdrawalAddressContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	node *node.Node
}

func (c *nodeConfirmPrimaryWithdrawalAddressContext) Initialize() error {
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

func (c *nodeConfirmPrimaryWithdrawalAddressContext) GetState(mc *batch.MultiCaller) {
	c.node.PendingPrimaryWithdrawalAddress.AddToQuery(mc)
}

func (c *nodeConfirmPrimaryWithdrawalAddressContext) PrepareData(data *api.NodeConfirmPrimaryWithdrawalAddressData, opts *bind.TransactOpts) error {
	data.IncorrectPendingAddress = (c.node.PendingPrimaryWithdrawalAddress.Get() != c.nodeAddress)
	data.CanConfirm = !(data.IncorrectPendingAddress)

	if data.CanConfirm {
		txInfo, err := c.node.ConfirmPrimaryWithdrawalAddress(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for ConfirmPrimaryWithdrawalAddress: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
