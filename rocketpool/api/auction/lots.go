package auction

import (
	"fmt"
	"math/big"

	"github.com/urfave/cli"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Settings
const (
	lotDetailsBatchSize uint64 = 200
)

func getLots(c *cli.Context) (*api.AuctionLotsResponse, error) {
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
	nodeAccount, err := w.GetNodeAccount()
	if err != nil {
		return nil, err
	}

	// Response
	response := api.AuctionLotsResponse{}

	// Create the bindings
	auctionMgr, err := auction.NewAuctionManager(rp)
	if err != nil {
		return nil, fmt.Errorf("error creating auction manager binding: %w", err)
	}

	// Get contract state
	err = rp.Query(func(mc *batch.MultiCaller) error {
		auctionMgr.GetLotCount(mc)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting contract state: %w", err)
	}

	// Get lot details
	lotCount := auctionMgr.Details.LotCount.Formatted()
	lots := make([]*auction.AuctionLot, lotCount)
	details := make([]api.LotDetails, lotCount)

	// Load details
	err = rp.BatchQuery(int(lotCount), int(lotCountDetailsBatchSize), func(mc *batch.MultiCaller, i int) error {
		lot, err := auction.NewAuctionLot(rp, uint64(i))
		if err != nil {
			return fmt.Errorf("error creating lot %d binding: %w", i, err)
		}
		lots[i] = lot
		lot.GetAllDetails(mc)
		lot.GetLotAddressBidAmount(mc, &details[i].NodeBidAmount, nodeAccount.Address)
		return nil
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting lot details: %w", err)
	}

	// Process details
	for i := 0; i < int(lotCount); i++ {
		fullDetails := &details[i]
		fullDetails.Details = lots[i].Details

		// Check lot conditions
		addressHasBid := (fullDetails.NodeBidAmount.Cmp(big.NewInt(0)) > 0)
		hasRemainingRpl := (fullDetails.Details.RemainingRplAmount.Cmp(big.NewInt(0)) > 0)

		fullDetails.ClaimAvailable = (addressHasBid && fullDetails.Details.IsCleared)
		fullDetails.BiddingAvailable = (!fullDetails.Details.IsCleared && hasRemainingRpl)
		fullDetails.RPLRecoveryAvailable = (fullDetails.Details.IsCleared && hasRemainingRpl && !fullDetails.Details.RplRecovered)
	}
	response.Lots = details

	return &response, nil
}
