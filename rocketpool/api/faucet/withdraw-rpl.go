package faucet

import (
	"context"
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type faucetWithdrawHandler struct {
	allowance *big.Int
}

func (h *faucetWithdrawHandler) CreateBindings(ctx *callContext) error {
	return nil
}

func (h *faucetWithdrawHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	f := ctx.f
	nodeAddress := ctx.nodeAddress

	f.GetBalance(mc)
	f.GetAllowanceFor(mc, &h.allowance, nodeAddress)
	f.GetWithdrawalFee(mc)
}

func (h *faucetWithdrawHandler) PrepareResponse(ctx *callContext, response *api.FaucetWithdrawRplData) error {
	rp := ctx.rp
	f := ctx.f
	address := ctx.nodeAddress
	opts := ctx.opts

	// Get node account balance
	nodeAccountBalance, err := rp.Client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return fmt.Errorf("error getting node account balance: %w", err)
	}

	// Populate the response
	response.InsufficientFaucetBalance = (f.Details.Balance.Cmp(big.NewInt(0)) == 0)
	response.InsufficientAllowance = (h.allowance.Cmp(big.NewInt(0)) == 0)
	response.InsufficientNodeBalance = (nodeAccountBalance.Cmp(f.Details.WithdrawalFee) < 0)
	response.CanWithdraw = !(response.InsufficientFaucetBalance || response.InsufficientAllowance || response.InsufficientNodeBalance)

	if response.CanWithdraw && opts != nil {
		opts.Value = f.Details.WithdrawalFee

		// Get withdrawal amount
		var amount *big.Int
		balance := f.Details.Balance
		if balance.Cmp(h.allowance) > 0 {
			amount = h.allowance
		} else {
			amount = balance
		}

		txInfo, err := f.Withdraw(opts, amount)
		if err != nil {
			return fmt.Errorf("error getting TX info for Withdraw: %w", err)
		}
		response.TxInfo = txInfo
	}
	return nil
}
