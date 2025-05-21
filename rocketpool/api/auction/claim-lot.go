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

func canClaimFromLot(c *cli.Context, lotIndex uint64) (*api.CanClaimFromLotResponse, error) {

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
	response := api.CanClaimFromLotResponse{}

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

	// Check if address has bid
	wg.Go(func() error {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return err
		}
		addressBidAmount, err := auction.GetLotAddressBidAmount(rp, lotIndex, nodeAccount.Address, nil)
		if err == nil {
			response.NoBidFromAddress = (addressBidAmount.Cmp(big.NewInt(0)) == 0)
		}
		return err
	})

	// Check if lot has cleared
	wg.Go(func() error {
		isCleared, err := auction.GetLotIsCleared(rp, lotIndex, nil)
		if err == nil {
			response.NotCleared = !isCleared
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		gasInfo, err := auction.EstimateClaimBidGas(rp, lotIndex, opts)
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
	response.CanClaim = !(response.DoesNotExist || response.NoBidFromAddress || response.NotCleared)
	return &response, nil

}

func claimFromLot(c *cli.Context, lotIndex uint64) (*api.ClaimFromLotResponse, error) {

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
	response := api.ClaimFromLotResponse{}

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

	// Claim from lot
	hash, err := auction.ClaimBid(rp, lotIndex, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
