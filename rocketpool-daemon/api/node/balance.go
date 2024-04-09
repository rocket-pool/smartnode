package node

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
		router, "balance", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeBalanceContext struct {
	handler *NodeHandler
}

func (c *nodeBalanceContext) PrepareData(data *api.NodeBalanceData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	ec := sp.GetEthClient()
	ctx := c.handler.ctx
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}
	err = sp.RequireEthClientSynced(ctx)
	if err != nil {
		return types.ResponseStatus_ClientsNotSynced, err
	}

	data.Balance, err = ec.BalanceAt(ctx, nodeAddress, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting ETH balance of node %s: %w", nodeAddress.Hex(), err)
	}
	return types.ResponseStatus_Success, nil
}
