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

type auctionRecoverHandler struct {
	lotIndex  uint64
	lot       *auction.AuctionLot
	pSettings *settings.ProtocolDaoSettings
}

func (h *auctionRecoverHandler) CreateBindings(rp *rocketpool.RocketPool) error {
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

func (h *auctionRecoverHandler) GetState(nodeAddress common.Address, mc *batch.MultiCaller) {
	h.lot.GetLotExists(mc)
	h.lot.GetLotEndBlock(mc)
	h.lot.GetLotRemainingRplAmount(mc)
	h.lot.GetLotRplRecovered(mc)
	h.pSettings.GetBidOnAuctionLotEnabled(mc)
}

func (h *auctionRecoverHandler) PrepareResponse(rp *rocketpool.RocketPool, nodeAccount accounts.Account, opts *bind.TransactOpts, response *api.RecoverRPLFromLotResponse) error {
	// Get the current block
	currentBlock, err := rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Check for validity
	response.DoesNotExist = !h.lot.Details.Exists
	response.BiddingNotEnded = !(currentBlock >= h.lot.Details.EndBlock.Formatted())
	response.NoUnclaimedRPL = (h.lot.Details.RemainingRplAmount.Cmp(big.NewInt(0)) == 0)
	response.RPLAlreadyRecovered = h.lot.Details.RplRecovered
	response.CanRecover = !(response.DoesNotExist || response.BiddingNotEnded || response.NoUnclaimedRPL || response.RPLAlreadyRecovered)

	// Get tx info
	if response.CanRecover {
		txInfo, err := h.lot.RecoverUnclaimedRpl(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for RecoverUnclaimedRpl: %w", err)
		}
		response.TxInfo = txInfo
	}
	return nil
}
