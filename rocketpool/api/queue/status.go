package queue

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type queueStatusContextFactory struct {
	handler *QueueHandler
}

func (f *queueStatusContextFactory) Create(vars map[string]string) (*queueStatusContext, error) {
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

	depositPool *deposit.DepositPool
	queue       *minipool.MinipoolQueue
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
	c.depositPool, err = deposit.NewDepositPool(c.rp)
	if err != nil {
		return fmt.Errorf("error creating deposit pool binding: %w", err)
	}
	c.queue, err = minipool.NewMinipoolQueue(c.rp)
	if err != nil {
		return fmt.Errorf("error creating minipool queue binding: %w", err)
	}
	return nil
}

func (c *queueStatusContext) GetState(mc *batch.MultiCaller) {
	c.depositPool.GetBalance(mc)
	c.queue.GetTotalCapacity(mc)
	c.queue.GetTotalLength(mc)
}

func (c *queueStatusContext) PrepareData(data *api.QueueStatusData, opts *bind.TransactOpts) error {
	data.DepositPoolBalance = c.depositPool.Details.Balance
	data.MinipoolQueueCapacity = c.queue.Details.TotalCapacity
	data.MinipoolQueueLength = c.queue.Details.TotalLength.Formatted()
	return nil
}
