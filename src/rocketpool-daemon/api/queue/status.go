package queue

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
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
		router, "status", f, f.handler.serviceProvider,
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

func (c *queueStatusContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	err := sp.RequireEthClientSynced()
	if err != nil {
		return err
	}

	// Bindings
	c.depositPool, err = deposit.NewDepositPoolManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating deposit pool manager binding: %w", err)
	}
	c.mpMgr, err = minipool.NewMinipoolManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool queue binding: %w", err)
	}
	return nil
}

func (c *queueStatusContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.depositPool.Balance,
		c.mpMgr.TotalQueueCapacity,
		c.mpMgr.TotalQueueLength,
	)
}

func (c *queueStatusContext) PrepareData(data *api.QueueStatusData, opts *bind.TransactOpts) error {
	data.DepositPoolBalance = c.depositPool.Balance.Get()
	data.MinipoolQueueCapacity = c.mpMgr.TotalQueueCapacity.Get()
	data.MinipoolQueueLength = c.mpMgr.TotalQueueLength.Formatted()
	return nil
}
