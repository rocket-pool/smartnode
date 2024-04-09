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

type nodeSwapRplContextFactory struct {
	handler *NodeHandler
}

func (f *nodeSwapRplContextFactory) Create(args url.Values) (*nodeSwapRplContext, error) {
	c := &nodeSwapRplContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("amount", args, input.ValidateBigInt, &c.amount),
	}
	return c, errors.Join(inputErrs...)
}

func (f *nodeSwapRplContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*nodeSwapRplContext, api.NodeSwapRplData](
		router, "swap-rpl", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type nodeSwapRplContext struct {
	handler     *NodeHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	amount     *big.Int
	fsrpl      *tokens.TokenRplFixedSupply
	rpl        *tokens.TokenRpl
	rplAddress common.Address
	balance    *big.Int
	allowance  *big.Int
}

func (c *nodeSwapRplContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.fsrpl, err = tokens.NewTokenRplFixedSupply(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating legacy RPL binding: %w", err)
	}
	c.rpl, err = tokens.NewTokenRpl(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating RPL binding: %w", err)
	}
	rplContract, err := c.rp.GetContract(rocketpool.ContractName_RocketTokenRPL)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating RPL contract: %w", err)
	}
	c.rplAddress = rplContract.Address
	return types.ResponseStatus_Success, nil
}

func (c *nodeSwapRplContext) GetState(mc *batch.MultiCaller) {
	c.fsrpl.BalanceOf(mc, &c.balance, c.nodeAddress)
	c.fsrpl.GetAllowance(mc, &c.allowance, c.nodeAddress, c.rplAddress)
}

func (c *nodeSwapRplContext) PrepareData(data *api.NodeSwapRplData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	data.InsufficientBalance = (c.amount.Cmp(c.balance) > 0)
	data.Allowance = c.allowance
	data.CanSwap = !(data.InsufficientBalance)

	if data.CanSwap {
		if c.allowance.Cmp(c.amount) < 0 {
			// Do the approve TX if needed
			approvalAmount := getMaxApproval()
			txInfo, err := c.fsrpl.Approve(c.rplAddress, approvalAmount, opts)
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error getting TX info to approve increasing legacy RPL's allowance: %w", err)
			}
			data.ApproveTxInfo = txInfo
		} else {
			// Just do the swap
			txInfo, err := c.rpl.SwapFixedSupplyRplForRpl(c.amount, opts)
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for SwapFixedSupplyRPLForRPL: %w", err)
			}
			data.SwapTxInfo = txInfo
		}
	}
	return types.ResponseStatus_Success, nil
}
