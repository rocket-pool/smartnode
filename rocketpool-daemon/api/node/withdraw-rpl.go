package node

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
		router, "withdraw-rpl", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeWithdrawRplContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	ec          core.ExecutionClient
	nodeAddress common.Address

	amount    *big.Int
	node      *node.Node
	pSettings *protocol.ProtocolDaoSettings
}

func (c *nodeWithdrawRplContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.ec = sp.GetEthClient()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
	c.node, err = node.NewNode(c.rp, c.nodeAddress)
	if err != nil {
		return fmt.Errorf("error creating node %s binding: %w", c.nodeAddress.Hex(), err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	return nil
}

func (c *nodeWithdrawRplContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.node.RplStake,
		c.node.MinimumRplStake,
		c.node.RplStakedTime,
		c.node.IsRplWithdrawalAddressSet,
		c.node.RplWithdrawalAddress,
		c.pSettings.Rewards.IntervalTime,
	)
}

func (c *nodeWithdrawRplContext) PrepareData(data *api.NodeWithdrawRplData, opts *bind.TransactOpts) error {
	header, err := c.ec.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error getting latest block header: %w", err)
	}
	currentTime := time.Unix(int64(header.Time), 0)

	rplStake := c.node.RplStake.Get()
	minimumRplStake := c.node.MinimumRplStake.Get()
	remainingRplStake := big.NewInt(0).Sub(rplStake, c.amount)
	rplStakedTime := c.node.RplStakedTime.Formatted()
	withdrawalDelay := c.pSettings.Rewards.IntervalTime.Formatted()
	isRplWithdrawalAddressSet := c.node.IsRplWithdrawalAddressSet.Get()
	hasDifferentRplWithdrawalAddress := isRplWithdrawalAddressSet && c.nodeAddress != c.node.RplWithdrawalAddress.Get()

	data.InsufficientBalance = (c.amount.Cmp(rplStake) > 0)
	data.MinipoolsUndercollateralized = (remainingRplStake.Cmp(minimumRplStake) < 0)
	data.HasDifferentRplWithdrawalAddress = hasDifferentRplWithdrawalAddress
	data.WithdrawalDelayActive = (currentTime.Sub(rplStakedTime) < withdrawalDelay)

	// Update & return response
	data.CanWithdraw = !(data.InsufficientBalance || data.MinipoolsUndercollateralized || data.WithdrawalDelayActive || data.HasDifferentRplWithdrawalAddress)

	if data.CanWithdraw {
		txInfo, err := c.node.WithdrawRpl(c.amount, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for WithdrawRpl: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
