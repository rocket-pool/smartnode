package node

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeBalanceContextFactory struct {
	handler *NodeHandler
}

func (f *nodeBalanceContextFactory) Create(args url.Values) (*nodeBalanceContext, error) {
	c := &nodeBalanceContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *nodeBalanceContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessGet[*nodeBalanceContext, api.NodeBalanceData](
		router, "balance", f, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeBalanceContext struct {
	handler *NodeHandler
}

func (c *nodeBalanceContext) PrepareData(data *api.NodeBalanceData, opts *bind.TransactOpts) error {
	sp := c.handler.serviceProvider
	ec := sp.GetEthClient()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return err
	}

	data.Balance, err = ec.BalanceAt(context.Background(), nodeAddress, nil)
	if err != nil {
		return fmt.Errorf("error getting ETH balance of node %s: %w", nodeAddress.Hex(), err)
	}
	return nil
}
