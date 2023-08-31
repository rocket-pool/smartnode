package faucet

import (
	"context"
	"fmt"

	batch "github.com/rocket-pool/batch-query"
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
	response := api.FaucetStatusResponse{}

	// Data
	var wg errgroup.Group
	var currentPeriodStartBlock uint64
	var withdrawalPeriodBlocks uint64
	var currentBlock uint64

	// Get contract state
	wg.Go(func() error {
		err = rp.Query(func(mc *batch.MultiCaller) error {
			f.GetBalance(mc)
			f.GetAllowanceFor(mc, &response.Allowance, nodeAccount.Address)
			f.GetWithdrawalFee(mc)
			f.GetWithdrawalPeriodStart(mc)
			f.GetWithdrawalPeriod(mc)
			return nil
		}, nil)
		if err != nil {
			return fmt.Errorf("error getting contract state: %w", err)
		}
		return nil
	})

	// Get current block
	wg.Go(func() error {
		header, err := ec.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("error getting latest block header: %w", err)
		}
		currentBlock = header.Number.Uint64()
		return nil
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Populate the response
	response.Balance = f.Details.Balance
	response.Allowance = f.Details.Allowance
	response.WithdrawalFee = f.Details.WithdrawalFee
	currentPeriodStartBlock = f.Details.WithdrawalPeriodStart.Formatted()
	withdrawalPeriodBlocks = f.Details.WithdrawalPeriod.Formatted()

	// Get withdrawable amount
	if response.Balance.Cmp(response.Allowance) > 0 {
		response.WithdrawableAmount = response.Allowance
	} else {
		response.WithdrawableAmount = response.Balance
	}

	// Get reset block
	response.ResetsInBlocks = (currentPeriodStartBlock + withdrawalPeriodBlocks) - currentBlock

	// Return response
	return &response, nil

}
