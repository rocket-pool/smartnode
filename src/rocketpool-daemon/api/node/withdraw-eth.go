package node

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
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

type nodeWithdrawEthContextFactory struct {
	handler *NodeHandler
}

func (f *nodeWithdrawEthContextFactory) Create(args url.Values) (*nodeWithdrawEthContext, error) {
	c := &nodeWithdrawEthContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("amount", args, input.ValidateBigInt, &c.amount),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeWithdrawEthContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeWithdrawEthContext, api.NodeWithdrawEthData](
		router, "withdraw-eth", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeWithdrawEthContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	ec          eth.IExecutionClient
	nodeAddress common.Address

	amount *big.Int
	node   *node.Node
}

func (c *nodeWithdrawEthContext) Initialize() (types.ResponseStatus, error) {
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
	return types.ResponseStatus_Success, nil
}

func (c *nodeWithdrawEthContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.node.DonatedEthBalance,
		c.node.PrimaryWithdrawalAddress,
	)
}

func (c *nodeWithdrawEthContext) PrepareData(data *api.NodeWithdrawEthData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	ethBalance := c.node.DonatedEthBalance.Get()
	hasDifferentPrimaryWithdrawalAddress := c.nodeAddress != c.node.PrimaryWithdrawalAddress.Get()

	data.InsufficientBalance = (c.amount.Cmp(ethBalance) > 0)
	data.HasDifferentPrimaryWithdrawalAddress = hasDifferentPrimaryWithdrawalAddress

	// Update & return response
	data.CanWithdraw = !(data.InsufficientBalance || data.HasDifferentPrimaryWithdrawalAddress)

	if data.CanWithdraw {
		txInfo, err := c.node.WithdrawDonatedEth(c.amount, opts)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for WithdrawDonatedEth: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
