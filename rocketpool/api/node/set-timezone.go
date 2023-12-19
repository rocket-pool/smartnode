package node

import (
	"errors"
	"fmt"
	"net/url"
	_ "time/tzdata"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/rocketpool-go/node"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
	server.RegisterQuerylessGet[*nodeSetTimezoneContext, api.NodeSetTimezoneData](
		router, "set-timezone", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSetTimezoneContext struct {
	handler *NodeHandler

	timezoneLocation string
}

func (c *nodeSetTimezoneContext) PrepareData(data *api.NodeSetTimezoneData, opts *bind.TransactOpts) error {
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
	data.TxInfo, err = node.SetTimezoneLocation(c.timezoneLocation, opts)
	if err != nil {
		return fmt.Errorf("error getting TX info for SetTimezoneLocation: %w", err)
	}

	return nil
}
