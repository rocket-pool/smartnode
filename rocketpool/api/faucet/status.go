package faucet

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/shared/services/contracts"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type faucetStatusHandler struct {
	allowance *big.Int
}

func (h *faucetStatusHandler) CreateBindings(rp *rocketpool.RocketPool) error {
	return nil
}

func (h *faucetStatusHandler) GetState(f *contracts.RplFaucet, nodeAddress common.Address, mc *batch.MultiCaller) {
	f.GetBalance(mc)
	f.GetAllowanceFor(mc, &h.allowance, nodeAddress)
	f.GetWithdrawalFee(mc)
	f.GetWithdrawalPeriodStart(mc)
	f.GetWithdrawalPeriod(mc)
}

func (h *faucetStatusHandler) PrepareResponse(rp *rocketpool.RocketPool, f *contracts.RplFaucet, nodeAccount accounts.Account, opts *bind.TransactOpts, response *api.FaucetStatusResponse) error {
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
