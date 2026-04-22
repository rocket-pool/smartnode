package queue

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/bindings/deposit"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canProcessQueue(c *cli.Command, m int64) (*api.CanProcessQueueResponse, error) {

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
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanProcessQueueResponse{}

	// Data
	var wg errgroup.Group

	// Check deposit assignments are enabled
	wg.Go(func() error {
		assignDepositsEnabled, err := protocol.GetAssignDepositsEnabled(rp, nil)
		if err == nil {
			response.AssignDepositsDisabled = !assignDepositsEnabled
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := deposit.EstimateAssignDepositsGas(rp, big.NewInt(m), opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Check next minipool capacity & deposit pool balance
	response.CanProcess = !response.AssignDepositsDisabled
	return &response, nil

}

func processQueue(c *cli.Command, m int64, opts *bind.TransactOpts) (*api.ProcessQueueResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.ProcessQueueResponse{}

	// Process queue
	hash, err := deposit.AssignDeposits(rp, big.NewInt(m), opts)

	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
