package auction

import (
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type auctionLotHandler struct {
	auctionMgr *auction.AuctionManager
}

func NewAuctionLotHandler(vars map[string]string) (*auctionLotHandler, error) {
	h := &auctionLotHandler{}
	return h, nil
}

func (h *auctionLotHandler) CreateBindings(ctx *callContext) error {
	var err error
	rp := ctx.rp

	h.auctionMgr, err = auction.NewAuctionManager(rp)
	if err != nil {
		return fmt.Errorf("error creating auction manager binding: %w", err)
	}
	return nil
}

func (h *auctionLotHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	h.auctionMgr.GetLotCount(mc)
}

func (h *auctionLotHandler) PrepareData(ctx *callContext, data *api.AuctionLotsData) error {
	rp := ctx.rp
	nodeAddress := ctx.nodeAddress

	// Get lot details
	lotCount := h.auctionMgr.Details.LotCount.Formatted()
	lots := make([]*auction.AuctionLot, lotCount)
	details := make([]api.AuctionLotDetails, lotCount)

	// Load details
	err := rp.BatchQuery(int(lotCount), int(lotCountDetailsBatchSize), func(mc *batch.MultiCaller, i int) error {
		lot, err := auction.NewAuctionLot(rp, uint64(i))
		if err != nil {
			return fmt.Errorf("error creating lot %d binding: %w", i, err)
		}
		lots[i] = lot
		lot.GetAllDetails(mc)
		lot.GetLotAddressBidAmount(mc, &details[i].NodeBidAmount, nodeAddress)
		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("error getting lot details: %w", err)
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
	data.Lots = details
	return nil
}
