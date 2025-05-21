package auction

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/smartnode/bindings/auction"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/eth1"
)

func canBidOnLot(c *cli.Context, lotIndex uint64, amountWei *big.Int) (*api.CanBidOnLotResponse, error) {

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
	response := api.CanBidOnLotResponse{}

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
			response.BiddingEnded = biddingEnded
		}
		return err
	})

	// Check lot remaining RPL amount
	wg.Go(func() error {
		remainingRpl, err := auction.GetLotRemainingRPLAmount(rp, lotIndex, nil)
		if err == nil {
			response.RPLExhausted = (remainingRpl.Cmp(big.NewInt(0)) == 0)
		}
		return err
	})

	// Check if lot bidding is enabled
	wg.Go(func() error {
		bidOnLotEnabled, err := protocol.GetBidOnLotEnabled(rp, nil)
		if err == nil {
			response.BidOnLotDisabled = !bidOnLotEnabled
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		opts.Value = amountWei
		gasInfo, err := auction.EstimatePlaceBidGas(rp, lotIndex, opts)
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
	response.CanBid = !(response.DoesNotExist || response.BiddingEnded || response.RPLExhausted || response.BidOnLotDisabled)
	return &response, nil

}

func bidOnLot(c *cli.Context, lotIndex uint64, amountWei *big.Int) (*api.BidOnLotResponse, error) {

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
	response := api.BidOnLotResponse{}

	// Get transactor
	opts, err := w.GetNodeAccountTransactor()
	if err != nil {
		return nil, err
	}
	opts.Value = amountWei

	// Override the provided pending TX if requested
	err = eth1.CheckForNonceOverride(c, opts)
	if err != nil {
		return nil, fmt.Errorf("Error checking for nonce override: %w", err)
	}

	// Bid on lot
	hash, err := auction.PlaceBid(rp, lotIndex, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}
