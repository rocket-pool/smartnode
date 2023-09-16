package faucet

import (
	"context"
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type faucetStatusContextFactory struct {
	h *FaucetHandler
}

func (f *faucetStatusContextFactory) Create(vars map[string]string) (*faucetStatusContext, error) {
	c := &faucetStatusContext{
		h: f.h,
	}
	return c, nil
}

func (f *faucetStatusContextFactory) Run(c *faucetStatusContext) (*api.ApiResponse[api.FaucetStatusData], error) {
	return runFaucetCall[api.FaucetStatusData](c)
}

// ===============
// === Context ===
// ===============

type faucetStatusContext struct {
	h         *FaucetHandler
	allowance *big.Int
	*commonContext
}

func (c *faucetStatusContext) CreateBindings(ctx *commonContext) error {
	c.commonContext = ctx
	return nil
}

func (c *faucetStatusContext) GetState(mc *batch.MultiCaller) {
	c.f.GetBalance(mc)
	c.f.GetAllowanceFor(mc, &c.allowance, c.nodeAddress)
	c.f.GetWithdrawalFee(mc)
	c.f.GetWithdrawalPeriodStart(mc)
	c.f.GetWithdrawalPeriod(mc)
}

func (c *faucetStatusContext) PrepareData(data *api.FaucetStatusData) error {
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
