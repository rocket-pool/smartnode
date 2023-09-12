package auction

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type auctionBidHandler struct {
	lotIndex  uint64
	amountWei *big.Int
	lot       *auction.AuctionLot
	pSettings *settings.ProtocolDaoSettings
}

func (h *auctionBidHandler) CreateBindings(rp *rocketpool.RocketPool) error {
	var err error
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

func (h *auctionBidHandler) GetState(nodeAddress common.Address, mc *batch.MultiCaller) {
	h.lot.GetLotExists(mc)
	h.lot.GetLotEndBlock(mc)
	h.lot.GetLotRemainingRplAmount(mc)
	h.pSettings.GetBidOnAuctionLotEnabled(mc)
}

func (h *auctionBidHandler) PrepareResponse(rp *rocketpool.RocketPool, nodeAccount accounts.Account, opts *bind.TransactOpts, response *api.BidOnLotResponse) error {
	// Get the current block
	currentBlock, err := rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Check for validity
	response.DoesNotExist = !h.lot.Details.Exists
	response.BiddingEnded = (currentBlock >= h.lot.Details.EndBlock.Formatted())
	response.RPLExhausted = (h.lot.Details.RemainingRplAmount.Cmp(big.NewInt(0)) == 0)
	response.BidOnLotDisabled = !h.pSettings.Details.Auction.IsBidOnLotEnabled
	response.CanBid = !(response.DoesNotExist || response.BiddingEnded || response.RPLExhausted || response.BidOnLotDisabled)

	// Get tx info
	if response.CanBid {
		txInfo, err := h.lot.PlaceBid(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		response.TxInfo = txInfo
	}
	return nil
}
