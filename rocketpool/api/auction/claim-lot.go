package auction

import (
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type auctionClaimHandler struct {
	lotIndex         uint64
	addressBidAmount *big.Int
	lot              *auction.AuctionLot
}

func (h *auctionClaimHandler) CreateBindings(ctx *callContext) error {
	var err error
	rp := ctx.rp

	h.lot, err = auction.NewAuctionLot(rp, h.lotIndex)
	if err != nil {
		return fmt.Errorf("error creating lot %d binding: %w", h.lotIndex, err)
	}
	return nil
}

func (h *auctionClaimHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	nodeAddress := ctx.nodeAddress

	h.lot.GetLotExists(mc)
	h.lot.GetLotAddressBidAmount(mc, &h.addressBidAmount, nodeAddress)
	h.lot.GetLotIsCleared(mc)
}

func (h *auctionClaimHandler) PrepareData(ctx *callContext, data *api.ClaimFromLotData) error {
	opts := ctx.opts

	// Check for validity
	data.DoesNotExist = !h.lot.Details.Exists
	data.NoBidFromAddress = (h.addressBidAmount.Cmp(big.NewInt(0)) == 0)
	data.NotCleared = !h.lot.Details.IsCleared
	data.CanClaim = !(data.DoesNotExist || data.NoBidFromAddress || data.NotCleared)

	// Get tx info
	if data.CanClaim && opts != nil {
		txInfo, err := h.lot.ClaimBid(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
