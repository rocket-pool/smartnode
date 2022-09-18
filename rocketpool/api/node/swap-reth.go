package node

import (
	"context"
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canSwapEth(c *cli.Context, amountWei *big.Int) (*api.CanStakeEthResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
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
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanStakeEthResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg1 errgroup.Group
	var maxPoolSize *big.Int
	var currentPoolSize *big.Int
	var depositFee *big.Int
	var amountWeiReth *big.Int

	// Check node balance
	wg1.Go(func() error {
		ethBalanceWei, err := ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
		if err == nil {
			response.InsufficientBalance = (amountWei.Cmp(ethBalanceWei) > 0)
		}
		return err
	})

	// Check deposits are enabled
	wg1.Go(func() error {
		depositEnabled, err := protocol.GetDepositEnabled(rp, nil)
		if err == nil {
			response.DepositDisabled = !depositEnabled
		}
		return err
	})

	// Check amount is above minimum
	wg1.Go(func() error {
		minDeposit, err := protocol.GetMinimumDeposit(rp, nil)
		if err == nil {
			response.BelowMinStakeAmount = amountWei.Cmp(minDeposit) < 0
		}
		return err
	})

	// Get max pool size
	wg1.Go(func() error {
		var err error
		maxPoolSize, err = protocol.GetMaximumDepositPoolSize(rp, nil)
		return err
	})

	// Get current pool size
	wg1.Go(func() error {
		var err error
		currentPoolSize, err = deposit.GetBalance(rp, nil)
		return err
	})

	// Get deposit fee
	wg1.Go(func() error {
		var err error
		depositFee, err = protocol.GetDepositFee(rp, nil)
		return err
	})

	// Get reth amount
	wg1.Go(func() error {
		var err error
		amountWeiReth, err = tokens.GetRETHValueOfETH(rp, amountWei, nil)
		return err
	})

	// Get gas estimates
	wg1.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		opts.Value = amountWei

		gasInfo, err := deposit.EstimateDepositGas(rp, opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return nil, err
	}

	// Note: There might be space after the minipool queue has been cleared, if possible
	response.DepositPoolFull = amountWei.Cmp(currentPoolSize.Sub(maxPoolSize, currentPoolSize)) > 0

	// Update & return response
	response.CanStake = !(response.InsufficientBalance || response.DepositDisabled || response.BelowMinStakeAmount || response.DepositPoolFull)
	if response.CanStake {
		var tmp big.Int
		var amountWeiRethWithFees big.Int
		tmp.Mul(amountWeiReth, depositFee)
		tmp.Quo(&tmp, eth.EthToWei(1))
		amountWeiRethWithFees.Sub(amountWeiReth, &tmp)
		response.RethAmount = &amountWeiRethWithFees
	}
	return &response, nil

}

func swapEth(c *cli.Context, amountWei *big.Int) (*api.StakeEthResponse, error) {

	// Get services
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.StakeEthResponse{}

	// Swap ETH for rETH
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	opts.Value = amountWei
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}
	if hash, err := deposit.Deposit(rp, opts); err != nil {
		return nil, err
	} else {
		response.StakeTxHash = hash
	}

	// Return response
	return &response, nil

}
