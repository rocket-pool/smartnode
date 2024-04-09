package queue

import (
	"fmt"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/deposit"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/node/wallet"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type queueProcessContextFactory struct {
	handler *QueueHandler
}

func (f *queueProcessContextFactory) Create(args url.Values) (*queueProcessContext, error) {
	c := &queueProcessContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *queueProcessContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*queueProcessContext, api.QueueProcessData](
		router, "process", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type queueProcessContext struct {
	handler *QueueHandler
	rp      *rocketpool.RocketPool
	w       *wallet.Wallet

	pSettings   *protocol.ProtocolDaoSettings
	depositPool *deposit.DepositPoolManager
}

func (c *queueProcessContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.w = sp.GetWallet()

	// Requirements
	err := sp.RequireNodeAddress()
	if err != nil {
		return types.ResponseStatus_AddressNotPresent, err
	}
	status, err := sp.RequireRocketPoolContracts(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	c.depositPool, err = deposit.NewDepositPoolManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating deposit pool binding: %w", err)
	}
	return types.ResponseStatus_Success, nil
}

func (c *queueProcessContext) GetState(mc *batch.MultiCaller) {
	c.pSettings.Deposit.AreDepositAssignmentsEnabled.AddToQuery(mc)
}

func (c *queueProcessContext) PrepareData(data *api.QueueProcessData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.AssignDepositsDisabled = !c.pSettings.Deposit.AreDepositAssignmentsEnabled.Get()
	data.CanProcess = !data.AssignDepositsDisabled

	if data.CanProcess {
		txInfo, err := c.depositPool.AssignDeposits(opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for AssignDeposits: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
