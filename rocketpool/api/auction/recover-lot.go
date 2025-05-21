package auction

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/auction"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canRecoverRplFromLot(c *cli.Context, lotIndex uint64) (*api.CanRecoverRPLFromLotResponse, error) {

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
	response := api.CanRecoverRPLFromLotResponse{}

	// Sync
	var wg errgroup.Group

	// Check if lot exists
	wg.Go(func() error {
		lotExists, err := auction.GetLotExists(rp, lotIndex, nil)
		if err == nil {
			response.DoesNotExist = !lotExists
		}
		return err
	})

	// Check if lot bidding has ended
	wg.Go(func() error {
		biddingEnded, err := getLotBiddingEnded(rp, lotIndex)
		if err == nil {
			response.BiddingNotEnded = !biddingEnded
		}
		return err
	})

	// Check if lot contains unclaimed RPL
	wg.Go(func() error {
		remainingRpl, err := auction.GetLotRemainingRPLAmount(rp, lotIndex, nil)
		if err == nil {
			response.NoUnclaimedRPL = (remainingRpl.Cmp(big.NewInt(0)) == 0)
		}
		return err
	})

	// Check if unclaimed RPL has already been recovered
	wg.Go(func() error {
		rplRecovered, err := auction.GetLotRPLRecovered(rp, lotIndex, nil)
		if err == nil {
			response.RPLAlreadyRecovered = rplRecovered
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := auction.EstimateRecoverUnclaimedRPLGas(rp, lotIndex, opts)
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
	response.CanRecover = !(response.DoesNotExist || response.BiddingNotEnded || response.NoUnclaimedRPL || response.RPLAlreadyRecovered)
	return &response, nil

}

func recoverRplFromLot(c *cli.Context, lotIndex uint64) (*api.RecoverRPLFromLotResponse, error) {

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
	response := api.RecoverRPLFromLotResponse{}

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

	// Recover unclaimed RPL from lot
	hash, err := auction.RecoverUnclaimedRPL(rp, lotIndex, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
