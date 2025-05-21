package auction

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/auction"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canCreateLot(c *cli.Context) (*api.CanCreateLotResponse, error) {

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
	response := api.CanCreateLotResponse{}

	// Sync
	var wg errgroup.Group

	// Check if sufficient remaining RPL is available to create a lot
	wg.Go(func() error {
		sufficientRemainingRplForLot, err := getSufficientRemainingRPLForLot(rp)
		if err == nil {
			response.InsufficientBalance = !sufficientRemainingRplForLot
		}
		return err
	})

	// Check if lot creation is enabled
	wg.Go(func() error {
		createLotEnabled, err := protocol.GetCreateLotEnabled(rp, nil)
		if err == nil {
			response.CreateLotDisabled = !createLotEnabled
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := auction.EstimateCreateLotGas(rp, opts)
		if err == nil {
			response.GasInfo = gasInfo
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.CanCreate = !(response.InsufficientBalance || response.CreateLotDisabled)
	return &response, nil

}

func createLot(c *cli.Context) (*api.CreateLotResponse, error) {

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
	response := api.CreateLotResponse{}

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

	// Create lot
	lotIndex, hash, err := auction.CreateLot(rp, opts)
	if err != nil {
		return nil, err
	}
	response.LotId = lotIndex
	response.TxHash = hash

	// Return response
	return &response, nil

}
