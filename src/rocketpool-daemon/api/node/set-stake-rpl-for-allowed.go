package node

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/v2/node"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
		router, "set-stake-rpl-for-allowed", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
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

func (c *nodeSetStakeRplForAllowedContext) PrepareData(data *api.NodeSetStakeRplForAllowedData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	rp := sp.GetRocketPool()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	node, err := node.NewNode(rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node binding: %w", err)
	}

	data.CanSet = true
	data.TxInfo, err = node.SetStakeRplForAllowed(c.caller, c.allowed, opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for SetStakeRplForAllowed: %w", err)
	}

	return types.ResponseStatus_Success, nil
}
