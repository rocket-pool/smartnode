package faucet

import (
	"context"

	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getStatus(c *cli.Context) (*api.FaucetStatusResponse, error) {

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
	ec, err := services.GetEthClient(c)
	if err != nil {
		return nil, err
	}
	f, err := services.GetRplFaucet(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.FaucetStatusResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg errgroup.Group
	var currentPeriodStartBlock uint64
	var withdrawalPeriodBlocks uint64
	var currentBlock uint64

	// Get faucet balance
	wg.Go(func() error {
		balance, err := f.GetBalance(nil)
		if err != nil {
			return err
		}
		response.Balance.Set(balance)
		return nil
	})

	// Get allowance
	wg.Go(func() error {
		allowance, err := f.GetAllowanceFor(nil, nodeAccount.Address)
		if err != nil {
			return err
		}
		response.Allowance.Set(allowance)
		return nil
	})

	// Get withdrawal fee
	wg.Go(func() error {
		withdrawalFee, err := f.WithdrawalFee(nil)
		if err != nil {
			return err
		}
		response.WithdrawalFee.Set(withdrawalFee)
		return nil
	})

	// Get current withdrawal period start block
	wg.Go(func() error {
		withdrawalPeriodStart, err := f.GetWithdrawalPeriodStart(nil)
		if err == nil {
			currentPeriodStartBlock = withdrawalPeriodStart.Uint64()
		}
		return err
	})

	// Get withdrawal period
	wg.Go(func() error {
		withdrawalPeriod, err := f.WithdrawalPeriod(nil)
		if err == nil {
			withdrawalPeriodBlocks = withdrawalPeriod.Uint64()
		}
		return err
	})

	// Get current block
	wg.Go(func() error {
		header, err := ec.HeaderByNumber(context.Background(), nil)
		if err == nil {
			currentBlock = header.Number.Uint64()
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Get withdrawable amount
	if response.Balance.Cmp(&response.Allowance) > 0 {
		response.WithdrawableAmount.Set(&response.Allowance)
	} else {
		response.WithdrawableAmount.Set(&response.Balance)
	}

	// Get reset block
	response.ResetsInBlocks = (currentPeriodStartBlock + withdrawalPeriodBlocks) - currentBlock

	// Return response
	return &response, nil

}
