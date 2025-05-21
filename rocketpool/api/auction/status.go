package auction

import (
	"github.com/rocket-pool/smartnode/bindings/auction"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func getStatus(c *cli.Context) (*api.AuctionStatusResponse, error) {

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
	response := api.AuctionStatusResponse{}

	// Sync
	var wg errgroup.Group

	// Get auction contract RPL balances
	wg.Go(func() error {
		totalRplBalance, err := auction.GetTotalRPLBalance(rp, nil)
		if err == nil {
			response.TotalRPLBalance = totalRplBalance
		}
		return err
	})
	wg.Go(func() error {
		allottedRplBalance, err := auction.GetAllottedRPLBalance(rp, nil)
		if err == nil {
			response.AllottedRPLBalance = allottedRplBalance
		}
		return err
	})
	wg.Go(func() error {
		remainingRplBalance, err := auction.GetRemainingRPLBalance(rp, nil)
		if err == nil {
			response.RemainingRPLBalance = remainingRplBalance
		}
		return err
	})

	// Check if lot can be created
	wg.Go(func() error {
		sufficientRemainingRplForLot, err := getSufficientRemainingRPLForLot(rp)
		if err == nil {
			response.CanCreateLot = sufficientRemainingRplForLot
		}
		return err
	})

	// Get lot counts
	wg.Go(func() error {
		nodeAccount, err := w.GetNodeAccount()
		if err != nil {
			return err
		}
		lotCountDetails, err := getAllLotCountDetails(rp, nodeAccount.Address)
		if err == nil {
			for _, details := range lotCountDetails {
				if details.AddressHasBid && details.Cleared {
					response.LotCounts.ClaimAvailable++
				}
				if !details.Cleared && details.HasRemainingRpl {
					response.LotCounts.BiddingAvailable++
				}
				if details.Cleared && details.HasRemainingRpl && !details.RplRecovered {
					response.LotCounts.RPLRecoveryAvailable++
				}
			}
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Return response
	return &response, nil

}
