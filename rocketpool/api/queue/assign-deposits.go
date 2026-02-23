package queue

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/deposit"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canAssignDeposits(c *cli.Context, max int64) (*api.CanAssignDepositsResponse, error) {

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
	response := api.CanAssignDepositsResponse{}

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
		gasInfo, err := deposit.EstimateAssignDepositsGas(rp, big.NewInt(max), opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	response.CanAssign = !response.AssignDepositsDisabled
	return &response, nil

}

func assignDeposits(c *cli.Context, max int64) (*api.AssignDepositsResponse, error) {

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
	response := api.AssignDepositsResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Assign deposits
	hash, err := deposit.AssignDeposits(rp, big.NewInt(max), opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
