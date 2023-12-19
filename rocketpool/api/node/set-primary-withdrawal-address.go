package node

import (
	"errors"
	"fmt"
	"net/url"

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
		router, "primary-withdrawal-address/set", f, f.handler.serviceProvider,
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

func (c *nodeSetPrimaryWithdrawalAddressContext) Initialize() error {
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

func (c *nodeSetPrimaryWithdrawalAddressContext) GetState(mc *batch.MultiCaller) {
	c.node.PrimaryWithdrawalAddress.AddToQuery(mc)
}

func (c *nodeSetPrimaryWithdrawalAddressContext) PrepareData(data *api.NodeSetPrimaryWithdrawalAddressData, opts *bind.TransactOpts) error {
	data.AddressAlreadySet = (c.node.PrimaryWithdrawalAddress.Get() != c.nodeAddress)
	data.CanSet = !(data.AddressAlreadySet)

	if data.CanSet {
		txInfo, err := c.node.SetPrimaryWithdrawalAddress(c.address, c.confirm, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for SetPrimaryWithdrawalAddress: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
