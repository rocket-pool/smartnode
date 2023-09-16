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

type faucetWithdrawContextFactory struct {
	h *FaucetHandler
}

func (f *faucetWithdrawContextFactory) Create(vars map[string]string) (*faucetWithdrawContext, error) {
	c := &faucetWithdrawContext{
		h: f.h,
	}
	return c, nil
}

func (f *faucetWithdrawContextFactory) Run(c *faucetWithdrawContext) (*api.ApiResponse[api.FaucetWithdrawRplData], error) {
	return runFaucetCall[api.FaucetWithdrawRplData](c)
}

// ===============
// === Context ===
// ===============

type faucetWithdrawContext struct {
	h         *FaucetHandler
	allowance *big.Int
	*commonContext
}

func (c *faucetWithdrawContext) CreateBindings(ctx *commonContext) error {
	c.commonContext = ctx
	return nil
}

func (c *faucetWithdrawContext) GetState(mc *batch.MultiCaller) {
	c.f.GetBalance(mc)
	c.f.GetAllowanceFor(mc, &c.allowance, c.nodeAddress)
	c.f.GetWithdrawalFee(mc)
}

func (c *faucetWithdrawContext) PrepareData(data *api.FaucetWithdrawRplData) error {
	// Get node account balance
	nodeAccountBalance, err := c.rp.Client.BalanceAt(context.Background(), c.nodeAddress, nil)
	if err != nil {
		return fmt.Errorf("error getting node account balance: %w", err)
	}

	// Populate the response
	data.InsufficientFaucetBalance = (c.f.Details.Balance.Cmp(big.NewInt(0)) == 0)
	data.InsufficientAllowance = (c.allowance.Cmp(big.NewInt(0)) == 0)
	data.InsufficientNodeBalance = (nodeAccountBalance.Cmp(c.f.Details.WithdrawalFee) < 0)
	data.CanWithdraw = !(data.InsufficientFaucetBalance || data.InsufficientAllowance || data.InsufficientNodeBalance)

	if data.CanWithdraw && c.opts != nil {
		c.opts.Value = c.f.Details.WithdrawalFee

		// Get withdrawal amount
		var amount *big.Int
		balance := c.f.Details.Balance
		if balance.Cmp(c.allowance) > 0 {
			amount = c.allowance
		} else {
			amount = balance
		}

		txInfo, err := c.f.Withdraw(c.opts, amount)
		if err != nil {
			return fmt.Errorf("error getting TX info for Withdraw: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
