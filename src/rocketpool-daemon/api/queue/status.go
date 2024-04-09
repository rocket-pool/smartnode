package queue

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/deposit"
	"github.com/rocket-pool/rocketpool-go/v2/minipool"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type queueStatusContextFactory struct {
	handler *QueueHandler
}

func (f *queueStatusContextFactory) Create(args url.Values) (*queueStatusContext, error) {
	c := &queueStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *queueStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*queueStatusContext, api.QueueStatusData](
		router, "status", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type queueStatusContext struct {
	handler *QueueHandler
	rp      *rocketpool.RocketPool

	depositPool *deposit.DepositPoolManager
	mpMgr       *minipool.MinipoolManager
}

func (c *queueStatusContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.depositPool, err = deposit.NewDepositPoolManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating deposit pool manager binding: %w", err)
	}
	c.mpMgr, err = minipool.NewMinipoolManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating minipool queue binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *queueStatusContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.depositPool.Balance,
		c.mpMgr.TotalQueueCapacity,
		c.mpMgr.TotalQueueLength,
	)
}

func (c *queueStatusContext) PrepareData(data *api.QueueStatusData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.DepositPoolBalance = c.depositPool.Balance.Get()
	data.MinipoolQueueCapacity = c.mpMgr.TotalQueueCapacity.Get()
	data.MinipoolQueueLength = c.mpMgr.TotalQueueLength.Formatted()
	return types.ResponseStatus_Success, nil
}
