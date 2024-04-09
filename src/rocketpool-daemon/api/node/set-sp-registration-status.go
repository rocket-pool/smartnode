package node

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeSetSmoothingPoolRegistrationStatusContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSetSmoothingPoolRegistrationStatusContextFactory) Create(args url.Values) (*nodeSetSmoothingPoolRegistrationStatusContext, error) {
	c := &nodeSetSmoothingPoolRegistrationStatusContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("opt-in", args, input.ValidateBool, &c.state),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSetSmoothingPoolRegistrationStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeSetSmoothingPoolRegistrationStatusContext, api.NodeSetSmoothingPoolRegistrationStatusData](
		router, "set-smoothing-pool-registration-state", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSetSmoothingPoolRegistrationStatusContext struct {
	handler *NodeHandler
	rp      *rocketpool.RocketPool
	ec      eth.IExecutionClient

	state     bool
	node      *node.Node
	pMgr      *protocol.ProtocolDaoManager
	pSettings *protocol.ProtocolDaoSettings
}

func (c *nodeSetSmoothingPoolRegistrationStatusContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.ec = sp.GetEthClient()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	c.pMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = c.pMgr.Settings
	return types.ResponseStatus_Success, nil
}

func (c *nodeSetSmoothingPoolRegistrationStatusContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.SmoothingPoolRegistrationState,
		c.node.SmoothingPoolRegistrationChanged,
		c.pMgr.IntervalTime,
	)
}

func (c *nodeSetSmoothingPoolRegistrationStatusContext) PrepareData(data *api.NodeSetSmoothingPoolRegistrationStatusData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.NodeRegistered = c.node.SmoothingPoolRegistrationState.Get()

	// Get the time the user can next change their opt-in status
	ctx := c.handler.ctx
	latestBlockHeader, err := c.ec.HeaderByNumber(ctx, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block: %w", err)
	}
	latestBlockTime := time.Unix(int64(latestBlockHeader.Time), 0)

	regChangeTime := c.node.SmoothingPoolRegistrationChanged.Formatted()
	intervalTime := c.pMgr.IntervalTime.Formatted()
	changeAvailableTime := regChangeTime.Add(intervalTime)
	data.TimeLeftUntilChangeable = changeAvailableTime.Sub(latestBlockTime)

	data.CanChange = false
	if data.TimeLeftUntilChangeable < 0 {
		data.TimeLeftUntilChangeable = 0
		data.CanChange = true
	}

	// Ignore if the requested mode is already set
	if data.NodeRegistered == c.state {
		data.CanChange = false
	}

	if data.CanChange {
		txInfo, err := c.node.SetSmoothingPoolRegistrationState(c.state, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for SetSmoothingPoolRegistrationState: %w", err)
		}
		data.TxInfo = txInfo
	}

	return types.ResponseStatus_Success, nil
}
