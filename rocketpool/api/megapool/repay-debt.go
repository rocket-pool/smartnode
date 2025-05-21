package megapool

import (
	"context"
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/megapool"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canRepayDebt(c *cli.Context, amount *big.Int) (*api.CanRepayDebtResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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

	// Get the node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanRepayDebtResponse{}

	// Check if the megapool is deployed
	megapoolDeployed, err := megapool.GetMegapoolDeployed(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}
	if !megapoolDeployed {
		response.CanRepay = false
		return &response, nil
	}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get the megapool debt
	debt, err := mp.GetDebt(nil)
	if err != nil {
		return nil, err
	}

	// Check if amount is greater than debt
	if amount.Cmp(debt) > 0 {
		response.CanRepay = false
		response.NotEnoughDebt = true
		return &response, nil
	}

	// Get call options block number
	var blockNumber *big.Int

	// Check if node has enough balance to repay debt
	ethBalance, err := rp.Client.BalanceAt(context.Background(), nodeAccount.Address, blockNumber)
	if err != nil {
		return nil, err
	}
	if ethBalance.Cmp(amount) < 0 {
		response.CanRepay = false
		response.NotEnoughBalance = true
		return &response, nil
	}

	// Get gas estimate
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	gasInfo, err := mp.EstimateRepayDebtGas(opts)
	if err != nil {
		return nil, err
	}
	response.GasInfo = gasInfo
	response.CanRepay = true

	return &response, nil

}

func repayDebt(c *cli.Context, amount *big.Int) (*api.RepayDebtResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
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

	// Get the node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.RepayDebtResponse{}

	// Get the megapool address
	megapoolAddress, err := megapool.GetMegapoolExpectedAddress(rp, nodeAccount.Address, nil)
	if err != nil {
		return nil, err
	}

	// Load the megapool
	mp, err := megapool.NewMegaPoolV1(rp, megapoolAddress, nil)
	if err != nil {
		return nil, err
	}

	// Get the megapool debt
	debt, err := mp.GetDebt(nil)
	if err != nil {
		return nil, err
	}

	if debt.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("no debt to repay")
	}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	opts.Value = amount

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Repay debt
	hash, err := mp.RepayDebt(opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
