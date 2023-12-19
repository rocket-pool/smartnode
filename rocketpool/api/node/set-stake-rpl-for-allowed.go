package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/node"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type nodeSetStakeRplForAllowedContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSetStakeRplForAllowedContextFactory) Create(args url.Values) (*nodeSetStakeRplForAllowedContext, error) {
	c := &nodeSetStakeRplForAllowedContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("caller", args, input.ValidateAddress, &c.caller),
		server.ValidateArg("allowed", args, input.ValidateBool, &c.allowed),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSetStakeRplForAllowedContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeSetStakeRplForAllowedContext, api.NodeSetStakeRplForAllowedData](
		router, "set-stake-rpl-for-allowed", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSetStakeRplForAllowedContext struct {
	handler *NodeHandler

	caller  common.Address
	allowed bool
}

func (c *nodeSetStakeRplForAllowedContext) PrepareData(data *api.NodeSetStakeRplForAllowedData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	node, err := node.NewNode(rp, nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating node binding: %w", err)
	}

	data.CanSet = true
	data.TxInfo, err = node.SetStakeRplForAllowed(c.caller, c.allowed, opts)
	if err != nil {
		return fmt.Errorf("error getting TX info for SetStakeRplForAllowed: %w", err)
	}

	return nil
}
