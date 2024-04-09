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
	"github.com/rocket-pool/rocketpool-go/v2/node"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"
	"github.com/rocket-pool/rocketpool-go/v2/tokens"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type nodeStakeRplContextFactory struct {
	handler *NodeHandler
}

func (f *nodeStakeRplContextFactory) Create(args url.Values) (*nodeStakeRplContext, error) {
	c := &nodeStakeRplContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("amount", args, input.ValidateBigInt, &c.amount),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeStakeRplContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeStakeRplContext, api.NodeStakeRplData](
		router, "stake-rpl", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeStakeRplContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	amount    *big.Int
	rpl       *tokens.TokenRpl
	nsAddress common.Address
	node      *node.Node
	balance   *big.Int
	allowance *big.Int
}

func (c *nodeStakeRplContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
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
	c.rpl, err = tokens.NewTokenRpl(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating RPL binding: %w", err)
	}
	rns, err := c.rp.GetContract(rocketpool.ContractName_RocketNodeStaking)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating RocketNodeStaking binding: %w", err)
	}
	c.nsAddress = rns.Address
	return types.ResponseStatus_Success, nil
}

func (c *nodeStakeRplContext) GetState(mc *batch.MultiCaller) {
	c.rpl.BalanceOf(mc, &c.balance, c.nodeAddress)
	c.rpl.GetAllowance(mc, &c.allowance, c.nodeAddress, c.nsAddress)
}

func (c *nodeStakeRplContext) PrepareData(data *api.NodeStakeRplData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.InsufficientBalance = (c.amount.Cmp(c.balance) > 0)
	data.Allowance = c.allowance
	data.CanStake = !(data.InsufficientBalance)

	if data.CanStake {
		if c.allowance.Cmp(c.amount) < 0 {
			// Do the approve TX if needed
			approvalAmount := getMaxApproval()
			txInfo, err := c.rpl.Approve(c.nsAddress, approvalAmount, opts)
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error getting TX info to approve increasing RPL's allowance: %w", err)
			}
			data.ApproveTxInfo = txInfo
		} else {
			// Just do the stake
			txInfo, err := c.node.StakeRpl(c.amount, opts)
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for StakeRpl: %w", err)
			}
			data.StakeTxInfo = txInfo
		}
	}
	return types.ResponseStatus_Success, nil
}
