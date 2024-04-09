package faucet

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/contracts"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type faucetStatusContextFactory struct {
	handler *FaucetHandler
}

func (f *faucetStatusContextFactory) Create(args url.Values) (*faucetStatusContext, error) {
	c := &faucetStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *faucetStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*faucetStatusContext, api.FaucetStatusData](
		router, "status", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type faucetStatusContext struct {
	handler     *FaucetHandler
	rp          *rocketpool.RocketPool
	f           *contracts.RplFaucet
	nodeAddress common.Address

	allowance *big.Int
}

func (c *faucetStatusContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.f = sp.GetRplFaucet()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}
	err = sp.RequireRplFaucet()
	if err != nil {
		return types.ResponseStatus_InvalidChainState, err
	}

	return types.ResponseStatus_Success, nil
}

func (c *faucetStatusContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.f.Balance,
		c.f.WithdrawalFee,
		c.f.WithdrawalPeriodStart,
		c.f.WithdrawalPeriod,
	)
	c.f.GetAllowanceFor(mc, &c.allowance, c.nodeAddress)
}

func (c *faucetStatusContext) PrepareData(data *api.FaucetStatusData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get the current block
	ctx := c.handler.ctx
	currentBlock, err := c.rp.Client.BlockNumber(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting current EL block: %w", err)
	}

	// Populate the response
	data.Balance = c.f.Balance.Get()
	data.Allowance = c.f.Allowance.Get()
	data.WithdrawalFee = c.f.WithdrawalFee.Get()
	currentPeriodStartBlock := c.f.WithdrawalPeriodStart.Formatted()
	withdrawalPeriodBlocks := c.f.WithdrawalPeriod.Formatted()

	// Get withdrawable amount
	if data.Balance.Cmp(data.Allowance) > 0 {
		data.WithdrawableAmount = data.Allowance
	} else {
		data.WithdrawableAmount = data.Balance
	}

	// Get reset block
	data.ResetsInBlocks = (currentPeriodStartBlock + withdrawalPeriodBlocks) - currentBlock
	return types.ResponseStatus_Success, nil
}
