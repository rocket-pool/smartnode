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

func (h *faucetStatusHandler) PrepareResponse(ctx *callContext, response *api.FaucetStatusData) error {
	rp := ctx.rp
	f := ctx.f

	// Get the current block
	currentBlock, err := rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Populate the response
	response.Balance = f.Details.Balance
	response.Allowance = f.Details.Allowance
	response.WithdrawalFee = f.Details.WithdrawalFee
	currentPeriodStartBlock := f.Details.WithdrawalPeriodStart.Formatted()
	withdrawalPeriodBlocks := f.Details.WithdrawalPeriod.Formatted()

	// Get withdrawable amount
	if response.Balance.Cmp(response.Allowance) > 0 {
		response.WithdrawableAmount = response.Allowance
	} else {
		response.WithdrawableAmount = response.Balance
	}

	// Get reset block
	response.ResetsInBlocks = (currentPeriodStartBlock + withdrawalPeriodBlocks) - currentBlock
	return nil
}
