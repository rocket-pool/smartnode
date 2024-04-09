package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
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

type nodeSetRplWithdrawalAddressContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSetRplWithdrawalAddressContextFactory) Create(args url.Values) (*nodeSetRplWithdrawalAddressContext, error) {
	c := &nodeSetRplWithdrawalAddressContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("address", args, input.ValidateAddress, &c.address),
		server.ValidateArg("confirm", args, input.ValidateBool, &c.confirm),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSetRplWithdrawalAddressContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeSetRplWithdrawalAddressContext, api.NodeSetRplWithdrawalAddressData](
		router, "rpl-withdrawal-address/set", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSetRplWithdrawalAddressContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	address common.Address
	confirm bool
	node    *node.Node
}

func (c *nodeSetRplWithdrawalAddressContext) Initialize() (types.ResponseStatus, error) {
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

func (c *nodeSetRplWithdrawalAddressContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.IsRplWithdrawalAddressSet,
		c.node.RplWithdrawalAddress,
		c.node.PrimaryWithdrawalAddress,
	)
}

func (c *nodeSetRplWithdrawalAddressContext) PrepareData(data *api.NodeSetRplWithdrawalAddressData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	isRplWithdrawalAddressSet := c.node.IsRplWithdrawalAddressSet.Get()
	data.PrimaryAddressDiffers = (c.node.PrimaryWithdrawalAddress.Get() != c.nodeAddress || isRplWithdrawalAddressSet)
	data.RplAddressDiffers = (isRplWithdrawalAddressSet && c.node.RplWithdrawalAddress.Get() != c.nodeAddress)
	data.CanSet = !(data.PrimaryAddressDiffers || data.RplAddressDiffers)

	if data.CanSet {
		txInfo, err := c.node.SetRplWithdrawalAddress(c.address, c.confirm, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for SetRplWithdrawalAddress: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
