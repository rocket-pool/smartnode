package auction

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/settings"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

type auctionBidHandler struct {
	lotIndex  uint64
	amountWei *big.Int
	lot       *auction.AuctionLot
	pSettings *settings.ProtocolDaoSettings
}

func NewAuctionBidHandler(vars map[string]string) (*auctionBidHandler, error) {
	h := &auctionBidHandler{}
	inputErrs := []error{
		server.ValidateArg("index", vars, cliutils.ValidateUint, &h.lotIndex),
		server.ValidateArg("amount", vars, cliutils.ValidatePositiveWeiAmount, &h.amountWei),
	}
	return h, errors.Join(inputErrs...)
}

func (h *auctionBidHandler) CreateBindings(ctx *callContext) error {
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

func (h *auctionBidHandler) GetState(ctx *callContext, mc *batch.MultiCaller) {
	h.lot.GetLotExists(mc)
	h.lot.GetLotEndBlock(mc)
	h.lot.GetLotRemainingRplAmount(mc)
	h.pSettings.GetBidOnAuctionLotEnabled(mc)
}

func (h *auctionBidHandler) PrepareData(ctx *callContext, Data *api.BidOnLotData) error {
	rp := ctx.rp
	opts := ctx.opts

	// Get the current block
	currentBlock, err := rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Check for validity
	Data.DoesNotExist = !h.lot.Details.Exists
	Data.BiddingEnded = (currentBlock >= h.lot.Details.EndBlock.Formatted())
	Data.RPLExhausted = (h.lot.Details.RemainingRplAmount.Cmp(big.NewInt(0)) == 0)
	Data.BidOnLotDisabled = !h.pSettings.Details.Auction.IsBidOnLotEnabled
	Data.CanBid = !(Data.DoesNotExist || Data.BiddingEnded || Data.RPLExhausted || Data.BidOnLotDisabled)

	// Get tx info
	if Data.CanBid && opts != nil {
		txInfo, err := h.lot.PlaceBid(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		Data.TxInfo = txInfo
	}
	return nil
}
