package faucet

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/contracts"
	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type faucetStatusContextFactory struct {
	handler *FaucetHandler
}

func (f *faucetStatusContextFactory) Create(vars map[string]string) (*faucetStatusContext, error) {
	c := &faucetStatusContext{
		handler: f.handler,
	}
	return c, nil
}

func (f *faucetStatusContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*faucetStatusContext, api.FaucetStatusData](
		router, "status", f, f.handler.serviceProvider,
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

func (c *faucetStatusContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.f = sp.GetRplFaucet()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

	// Requirements
	return errors.Join(
		sp.RequireNodeRegistered(),
		sp.RequireRplFaucet(),
	)
}

func (c *faucetStatusContext) GetState(mc *batch.MultiCaller) {
	c.f.GetBalance(mc)
	c.f.GetAllowanceFor(mc, &c.allowance, c.nodeAddress)
	c.f.GetWithdrawalFee(mc)
	c.f.GetWithdrawalPeriodStart(mc)
	c.f.GetWithdrawalPeriod(mc)
}

func (c *faucetStatusContext) PrepareData(data *api.FaucetStatusData, opts *bind.TransactOpts) error {
	// Get the current block
	currentBlock, err := c.rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Populate the response
	data.Balance = c.f.Details.Balance
	data.Allowance = c.f.Details.Allowance
	data.WithdrawalFee = c.f.Details.WithdrawalFee
	currentPeriodStartBlock := c.f.Details.WithdrawalPeriodStart.Formatted()
	withdrawalPeriodBlocks := c.f.Details.WithdrawalPeriod.Formatted()

	// Get withdrawable amount
	if data.Balance.Cmp(data.Allowance) > 0 {
		data.WithdrawableAmount = data.Allowance
	} else {
		data.WithdrawableAmount = data.Balance
	}

	// Get reset block
	data.ResetsInBlocks = (currentPeriodStartBlock + withdrawalPeriodBlocks) - currentBlock
	return nil
}
