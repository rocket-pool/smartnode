package faucet

import (
	"context"
	"fmt"
	"math/big"

	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func withdrawRpl(c *cli.Context) (*api.FaucetWithdrawRplResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRplFaucet(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	f, err := services.GetRplFaucet(c)
	if err != nil {
		return nil, err
	}
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, fmt.Errorf("error getting node account: %w", err)
	}

	// Response
	response := api.FaucetWithdrawRplResponse{}

	// Data
	var wg errgroup.Group
	var nodeAccountBalance *big.Int
	var balance *big.Int
	var allowance *big.Int

	// Get contract state
	wg.Go(func() error {
		err = rp.Query(func(mc *batch.MultiCaller) error {
			f.GetBalance(mc)
			f.GetAllowanceFor(mc, &allowance, nodeAccount.Address)
			f.GetWithdrawalFee(mc)
			return nil
		}, nil)
		if err != nil {
			return fmt.Errorf("error getting contract state: %w", err)
		}
		return nil
	})

	// Get node account balance
	wg.Go(func() error {
		var err error
		nodeAccountBalance, err = ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
		if err != nil {
			return fmt.Errorf("error getting node account balance: %w", err)
		}
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Populate the response
	response.InsufficientFaucetBalance = (f.Details.Balance.Cmp(big.NewInt(0)) == 0)
	response.InsufficientAllowance = (allowance.Cmp(big.NewInt(0)) == 0)
	response.InsufficientNodeBalance = (nodeAccountBalance.Cmp(f.Details.WithdrawalFee) < 0)
	response.CanWithdraw = !(response.InsufficientFaucetBalance || response.InsufficientAllowance || response.InsufficientNodeBalance)

	if response.CanWithdraw {
		// Get the gas estimate
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return nil, err
		}
		opts.Value = f.Details.WithdrawalFee

		// Get withdrawal amount
		var amount *big.Int
		if balance.Cmp(allowance) > 0 {
			amount = allowance
		} else {
			amount = balance
		}

		txInfo, err := f.Withdraw(opts, amount)
		if err != nil {
			return nil, err
		}
		response.TxInfo = txInfo
	}

	return &response, nil
}
