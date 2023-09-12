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

type faucetWithdrawHandler struct {
	allowance *big.Int
}

func (h *faucetWithdrawHandler) CreateBindings(rp *rocketpool.RocketPool) error {
	return nil
}

func (h *faucetWithdrawHandler) GetState(f *contracts.RplFaucet, nodeAddress common.Address, mc *batch.MultiCaller) {
	f.GetBalance(mc)
	f.GetAllowanceFor(mc, &h.allowance, nodeAddress)
	f.GetWithdrawalFee(mc)
}

func (h *faucetWithdrawHandler) PrepareResponse(rp *rocketpool.RocketPool, f *contracts.RplFaucet, nodeAccount accounts.Account, opts *bind.TransactOpts, response *api.FaucetWithdrawRplResponse) error {
	// Get node account balance
	nodeAccountBalance, err := rp.Client.BalanceAt(context.Background(), nodeAccount.Address, nil)
	if err != nil {
		return fmt.Errorf("error getting node account balance: %w", err)
	}

	// Populate the response
	response.InsufficientFaucetBalance = (f.Details.Balance.Cmp(big.NewInt(0)) == 0)
	response.InsufficientAllowance = (h.allowance.Cmp(big.NewInt(0)) == 0)
	response.InsufficientNodeBalance = (nodeAccountBalance.Cmp(f.Details.WithdrawalFee) < 0)
	response.CanWithdraw = !(response.InsufficientFaucetBalance || response.InsufficientAllowance || response.InsufficientNodeBalance)

	if response.CanWithdraw {
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
