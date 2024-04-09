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

type nodeSetPrimaryWithdrawalAddressContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSetPrimaryWithdrawalAddressContextFactory) Create(args url.Values) (*nodeSetPrimaryWithdrawalAddressContext, error) {
	c := &nodeSetPrimaryWithdrawalAddressContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
		server.ValidateArg("confirm", args, input.ValidateBool, &c.confirm),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSetPrimaryWithdrawalAddressContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeSetPrimaryWithdrawalAddressContext, api.NodeSetPrimaryWithdrawalAddressData](
		router, "primary-withdrawal-address/set", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSetPrimaryWithdrawalAddressContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	address common.Address
	confirm bool
	node    *node.Node
}

func (c *nodeSetPrimaryWithdrawalAddressContext) Initialize() (types.ResponseStatus, error) {
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

func (c *nodeSetPrimaryWithdrawalAddressContext) GetState(mc *batch.MultiCaller) {
	c.node.PrimaryWithdrawalAddress.AddToQuery(mc)
}

func (c *nodeSetPrimaryWithdrawalAddressContext) PrepareData(data *api.NodeSetPrimaryWithdrawalAddressData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.AddressAlreadySet = (c.node.PrimaryWithdrawalAddress.Get() != c.nodeAddress)
	data.CanSet = !(data.AddressAlreadySet)

	if data.CanSet {
		txInfo, err := c.node.SetPrimaryWithdrawalAddress(c.address, c.confirm, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for SetPrimaryWithdrawalAddress: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
