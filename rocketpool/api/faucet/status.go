package faucet

import (
	"context"
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type faucetStatusHandler struct {
	allowance *big.Int
}

func NewFaucetStatusHandler(vars map[string]string) (*faucetStatusHandler, error) {
	h := &faucetStatusHandler{}
	return h, nil
}

func (h *faucetStatusHandler) CreateBindings(ctx *callContext) error {
	return nil
}

func (h *faucetStatusHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	f := ctx.f
	address := ctx.nodeAddress

	f.GetBalance(mc)
	f.GetAllowanceFor(mc, &h.allowance, address)
	f.GetWithdrawalFee(mc)
	f.GetWithdrawalPeriodStart(mc)
	f.GetWithdrawalPeriod(mc)
}

func (h *faucetStatusHandler) PrepareData(ctx *callContext, data *api.FaucetStatusData) error {
	rp := ctx.rp
	f := ctx.f

	// Get the current block
	currentBlock, err := rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Populate the response
	data.Balance = f.Details.Balance
	data.Allowance = f.Details.Allowance
	data.WithdrawalFee = f.Details.WithdrawalFee
	currentPeriodStartBlock := f.Details.WithdrawalPeriodStart.Formatted()
	withdrawalPeriodBlocks := f.Details.WithdrawalPeriod.Formatted()

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
