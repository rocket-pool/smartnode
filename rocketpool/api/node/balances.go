package node

import (
	"context"
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/deposit"
	"github.com/rocket-pool/smartnode/bindings/node"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

func getNodeEthBalance(c *cli.Context) (*api.NodeEthBalanceResponse, error) {
	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
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
	response := api.NodeEthBalanceResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	response.Balance, err = ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting ETH balance of node %s: %w", nodeAccount.Address.Hex(), err)
	}

	var wg1 errgroup.Group

	// Check credit balance
	wg1.Go(func() error {
		ethBalanceWei, err := node.GetNodeCreditAndBalance(rp, nodeAccount.Address, nil)
		if err == nil {
			response.CreditBalance = ethBalanceWei
		}
		return err
	})

	// Check node balance
	wg1.Go(func() error {
		ethBalanceWei, err := ec.BalanceAt(context.Background(), nodeAccount.Address, nil)
		if err == nil {
			response.Balance = ethBalanceWei
		}
		return err
	})

	// Get deposit pool balance
	wg1.Go(func() error {
		var err error
		depositPoolBalance, err := deposit.GetBalance(rp, nil)
		if err == nil {
			response.DepositPoolBalance = depositPoolBalance
		}
		return err
	})

	// Wait for data
	if err := wg1.Wait(); err != nil {
		return nil, err
	}

	return &response, nil
}
