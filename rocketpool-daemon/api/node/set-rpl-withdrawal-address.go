package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		router, "rpl-withdrawal-address/set", f, f.handler.serviceProvider,
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

func (c *nodeSetRplWithdrawalAddressContext) Initialize() error {
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

func (c *nodeSetRplWithdrawalAddressContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.node.IsRplWithdrawalAddressSet,
		c.node.RplWithdrawalAddress,
		c.node.PrimaryWithdrawalAddress,
	)
}

func (c *nodeSetRplWithdrawalAddressContext) PrepareData(data *api.NodeSetRplWithdrawalAddressData, opts *bind.TransactOpts) error {
	isRplWithdrawalAddressSet := c.node.IsRplWithdrawalAddressSet.Get()
	data.PrimaryAddressDiffers = (c.node.PrimaryWithdrawalAddress.Get() != c.nodeAddress || isRplWithdrawalAddressSet)
	data.RplAddressDiffers = (isRplWithdrawalAddressSet && c.node.RplWithdrawalAddress.Get() != c.nodeAddress)
	data.CanSet = !(data.PrimaryAddressDiffers || data.RplAddressDiffers)

	if data.CanSet {
		txInfo, err := c.node.SetRplWithdrawalAddress(c.address, c.confirm, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for SetRplWithdrawalAddress: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
