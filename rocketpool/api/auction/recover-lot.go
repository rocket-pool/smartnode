package auction

import (
	"context"
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/settings"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type auctionRecoverHandler struct {
	lotIndex  uint64
	lot       *auction.AuctionLot
	pSettings *settings.ProtocolDaoSettings
}

func (h *auctionRecoverHandler) CreateBindings(ctx *callContext) error {
	var err error
	rp := ctx.rp

	h.lot, err = auction.NewAuctionLot(rp, h.lotIndex)
	if err != nil {
		return fmt.Errorf("error creating lot %d binding: %w", h.lotIndex, err)
	}
	h.pSettings, err = settings.NewProtocolDaoSettings(rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO settings binding: %w", err)
	}
	return nil
}

func (h *auctionRecoverHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	h.lot.GetLotExists(mc)
	h.lot.GetLotEndBlock(mc)
	h.lot.GetLotRemainingRplAmount(mc)
	h.lot.GetLotRplRecovered(mc)
	h.pSettings.GetBidOnAuctionLotEnabled(mc)
}

func (h *auctionRecoverHandler) PrepareData(ctx *callContext, data *api.RecoverRplFromLotData) error {
	rp := ctx.rp
	opts := ctx.opts

	// Get the current block
	currentBlock, err := rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Check for validity
	data.DoesNotExist = !h.lot.Details.Exists
	data.BiddingNotEnded = !(currentBlock >= h.lot.Details.EndBlock.Formatted())
	data.NoUnclaimedRPL = (h.lot.Details.RemainingRplAmount.Cmp(big.NewInt(0)) == 0)
	data.RPLAlreadyRecovered = h.lot.Details.RplRecovered
	data.CanRecover = !(data.DoesNotExist || data.BiddingNotEnded || data.NoUnclaimedRPL || data.RPLAlreadyRecovered)

	// Get tx info
	if data.CanRecover && opts != nil {
		txInfo, err := h.lot.RecoverUnclaimedRpl(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for RecoverUnclaimedRpl: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
