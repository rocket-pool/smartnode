package node

import (
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/v2/node"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
)

// ===============
// === Factory ===
// ===============

type nodeSetTimezoneContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSetTimezoneContextFactory) Create(args url.Values) (*nodeSetTimezoneContext, error) {
	c := &nodeSetTimezoneContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("timezone", args, input.ValidateTimezoneLocation, &c.timezoneLocation),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSetTimezoneContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeSetTimezoneContext, types.TxInfoData](
		router, "set-timezone", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSetTimezoneContext struct {
	handler *NodeHandler

	timezoneLocation string
}

func (c *nodeSetTimezoneContext) PrepareData(data *types.TxInfoData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
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

	data.TxInfo, err = node.SetTimezoneLocation(c.timezoneLocation, opts)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for SetTimezoneLocation: %w", err)
	}

	return types.ResponseStatus_Success, nil
}
