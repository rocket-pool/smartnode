package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type nodeBalanceContextFactory struct {
	handler *NodeHandler
}

func (f *nodeBalanceContextFactory) Create(vars map[string]string) (*nodeBalanceContext, error) {
	c := &nodeBalanceContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("addresses", vars, input.ValidateAddresses, &c.minipoolAddresses),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeBalanceContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterQuerylessRoute[*nodeBalanceContext, api.NodeBalanceData](
		router, "dissolve", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeBalanceContext struct {
	handler           *NodeHandler
	minipoolAddresses []common.Address
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
