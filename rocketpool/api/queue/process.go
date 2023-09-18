package queue

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/rocketpool/common/wallet"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type queueProcessContextFactory struct {
	handler *QueueHandler
}

func (f *queueProcessContextFactory) Create(vars map[string]string) (*queueProcessContext, error) {
	c := &queueProcessContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *queueProcessContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*queueProcessContext, api.QueueProcessData](
		router, "process", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type queueProcessContext struct {
	handler *QueueHandler
	rp      *rocketpool.RocketPool
	w       *wallet.LocalWallet

	pSettings   *settings.ProtocolDaoSettings
	depositPool *deposit.DepositPool
}

func (c *queueProcessContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.w = sp.GetWallet()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return err
	}

	// Bindings
	c.pSettings, err = settings.NewProtocolDaoSettings(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO settings binding: %w", err)
	}
	c.depositPool, err = deposit.NewDepositPool(c.rp)
	if err != nil {
		return fmt.Errorf("error creating deposit pool binding: %w", err)
	}
	return nil
}

func (c *queueProcessContext) GetState(mc *batch.MultiCaller) {
	c.pSettings.GetAssignPoolDepositsEnabled(mc)
}

func (c *queueProcessContext) PrepareData(data *api.QueueProcessData, opts *bind.TransactOpts) error {
	data.AssignDepositsDisabled = !c.pSettings.Details.Deposit.AreDepositAssignmentsEnabled
	data.CanProcess = !data.AssignDepositsDisabled

	if data.CanProcess && opts != nil {
		txInfo, err := c.depositPool.AssignDeposits(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for AssignDeposits: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
