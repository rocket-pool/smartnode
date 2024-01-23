package node

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		router, "set-smoothing-pool-registration-state", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSetSmoothingPoolRegistrationStatusContext struct {
	handler *NodeHandler
	rp      *rocketpool.RocketPool
	ec      core.ExecutionClient

	state     bool
	node      *node.Node
	pSettings *protocol.ProtocolDaoSettings
}

func (c *nodeSetSmoothingPoolRegistrationStatusContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.ec = sp.GetEthClient()
	nodeAddress, _ := sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating node %s binding: %w", nodeAddress.Hex(), err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	return nil
}

func (c *nodeSetSmoothingPoolRegistrationStatusContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.node.SmoothingPoolRegistrationState,
		c.node.SmoothingPoolRegistrationChanged,
		c.pSettings.Rewards.IntervalTime,
	)
}

func (c *nodeSetSmoothingPoolRegistrationStatusContext) PrepareData(data *api.NodeSetSmoothingPoolRegistrationStatusData, opts *bind.TransactOpts) error {
	data.NodeRegistered = c.node.SmoothingPoolRegistrationState.Get()

	// Get the time the user can next change their opt-in status
	latestBlockHeader, err := c.ec.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
	latestBlockTime := time.Unix(int64(latestBlockHeader.Time), 0)

	regChangeTime := c.node.SmoothingPoolRegistrationChanged.Formatted()
	intervalTime := c.pSettings.Rewards.IntervalTime.Formatted()
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
			return fmt.Errorf("error getting TX info for SetSmoothingPoolRegistrationState: %w", err)
		}
		data.TxInfo = txInfo
	}

	return nil
}
