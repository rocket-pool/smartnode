package node

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
		router, "primary-withdrawal-address/confirm", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
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

func (c *nodeConfirmPrimaryWithdrawalAddressContext) Initialize() (types.ResponseStatus, error) {
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

func (c *nodeConfirmPrimaryWithdrawalAddressContext) GetState(mc *batch.MultiCaller) {
	c.node.PendingPrimaryWithdrawalAddress.AddToQuery(mc)
}

func (c *nodeConfirmPrimaryWithdrawalAddressContext) PrepareData(data *api.NodeConfirmPrimaryWithdrawalAddressData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.IncorrectPendingAddress = (c.node.PendingPrimaryWithdrawalAddress.Get() != c.nodeAddress)
	data.CanConfirm = !(data.IncorrectPendingAddress)

	if data.CanConfirm {
		txInfo, err := c.node.ConfirmPrimaryWithdrawalAddress(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for ConfirmPrimaryWithdrawalAddress: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
