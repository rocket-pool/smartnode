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
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/tokens"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
		router, "stake-rpl", f, f.handler.serviceProvider,
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

func (c *nodeStakeRplContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
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
	c.rpl, err = tokens.NewTokenRpl(c.rp)
	if err != nil {
		return fmt.Errorf("error creating RPL binding: %w", err)
	}
	rns, err := c.rp.GetContract(rocketpool.ContractName_RocketNodeStaking)
	if err != nil {
		return fmt.Errorf("error creating RocketNodeStaking binding: %w", err)
	}
	c.nsAddress = *rns.Address
	return nil
}

func (c *nodeStakeRplContext) GetState(mc *batch.MultiCaller) {
	c.rpl.BalanceOf(mc, &c.balance, c.nodeAddress)
	c.rpl.GetAllowance(mc, &c.allowance, c.nodeAddress, c.nsAddress)
}

func (c *nodeStakeRplContext) PrepareData(data *api.NodeStakeRplData, opts *bind.TransactOpts) error {
	data.InsufficientBalance = (c.amount.Cmp(c.balance) > 0)
	data.Allowance = c.allowance
	data.CanStake = !(data.InsufficientBalance)

	if data.CanStake {
		// Check allowance
		if c.amount.Cmp(c.allowance) > 0 {
			approvalAmount := big.NewInt(0).Sub(c.amount, c.allowance)
			txInfo, err := c.rpl.Approve(c.nsAddress, approvalAmount, opts)
			if err != nil {
				return fmt.Errorf("error getting TX info to approve increasing RPL's allowance: %w", err)
			}
			data.ApproveTxInfo = txInfo
		}

		txInfo, err := c.node.StakeRpl(c.amount, opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for StakeRpl: %w", err)
		}
		data.StakeTxInfo = txInfo
	}
	return nil
}
