package node

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/node"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canNodeWithdrawCredit(c *cli.Command, amountWei *big.Int) (*api.CanNodeWithdrawCreditResponse, error) {

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

	// Response
	response := api.CanNodeWithdrawCreditResponse{}

	// Get node account
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Data
	var wg errgroup.Group
	var credit *big.Int

	wg.Go(func() error {
		var err error
		credit, err = node.GetNodeDepositCredit(rp, nodeAccount.Address, nil)
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := node.EstimateWithdrawCreditGas(rp, amountWei, opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check data
	response.InsufficientBalance = (amountWei.Cmp(credit) > 0)

	// Update & return response
	response.CanWithdraw = !(response.InsufficientBalance)
	return &response, nil

}

func nodeWithdrawCredit(c *cli.Command, amountWei *big.Int, opts *bind.TransactOpts) (*api.NodeWithdrawCreditResponse, error) {

	// Get services
	if err := services.RequireNodeRegistered(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.NodeWithdrawCreditResponse{}

	// Withdraw credit
	tx, err := node.WithdrawCredit(rp, amountWei, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = tx.Hash()

	// Return response
	return &response, nil

}
