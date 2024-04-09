package node

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeWithdrawRplContextFactory struct {
	handler *NodeHandler
}

func (f *nodeWithdrawRplContextFactory) Create(args url.Values) (*nodeWithdrawRplContext, error) {
	c := &nodeWithdrawRplContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("amount", args, input.ValidateBigInt, &c.amount),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeWithdrawRplContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeWithdrawRplContext, api.NodeWithdrawRplData](
		router, "withdraw-rpl", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeWithdrawRplContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	ec          eth.IExecutionClient
	nodeAddress common.Address

	amount    *big.Int
	node      *node.Node
	pMgr      *protocol.ProtocolDaoManager
	pSettings *protocol.ProtocolDaoSettings
}

func (c *nodeWithdrawRplContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.ec = sp.GetEthClient()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, c.nodeAddress)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating node %s binding: %w", c.nodeAddress.Hex(), err)
	}
	c.pMgr, err = protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = c.pMgr.Settings
	return types.ResponseStatus_Success, nil
}

func (c *nodeWithdrawRplContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.RplStake,
		c.node.MinimumRplStake,
		c.node.MaximumRplStake,
		c.node.RplLocked,
		c.node.RplStakedTime,
		c.node.IsRplWithdrawalAddressSet,
		c.node.RplWithdrawalAddress,
		c.pMgr.IntervalTime,
	)
}

func (c *nodeWithdrawRplContext) PrepareData(data *api.NodeWithdrawRplData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ctx := c.handler.ctx
	header, err := c.ec.HeaderByNumber(ctx, nil)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(header.Time), 0)

	rplStake := c.node.RplStake.Get()
	minimumRplStake := c.node.MinimumRplStake.Get()
	maximumRplStake := c.node.MaximumRplStake.Get()
	nodeRplLocked := c.node.RplLocked.Get()
	remainingRplStake := big.NewInt(0).Sub(rplStake, c.amount)
	remainingRplStake.Sub(remainingRplStake, nodeRplLocked)
	rplStakedTime := c.node.RplStakedTime.Formatted()
	withdrawalDelay := c.pMgr.IntervalTime.Formatted()
	isRplWithdrawalAddressSet := c.node.IsRplWithdrawalAddressSet.Get()
	hasDifferentRplWithdrawalAddress := isRplWithdrawalAddressSet && c.nodeAddress != c.node.RplWithdrawalAddress.Get()

	data.InsufficientBalance = (c.amount.Cmp(rplStake) > 0)
	data.BelowMaxRplStake = (remainingRplStake.Cmp(maximumRplStake) < 0)
	data.MinipoolsUndercollateralized = (remainingRplStake.Cmp(minimumRplStake) < 0)
	data.HasDifferentRplWithdrawalAddress = hasDifferentRplWithdrawalAddress
	data.WithdrawalDelayActive = (currentTime.Sub(rplStakedTime) < withdrawalDelay)

	// Update & return response
	data.CanWithdraw = !(data.InsufficientBalance || data.MinipoolsUndercollateralized || data.WithdrawalDelayActive || data.HasDifferentRplWithdrawalAddress || data.BelowMaxRplStake)

	if data.CanWithdraw {
		txInfo, err := c.node.WithdrawRpl(c.amount, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for WithdrawRpl: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
