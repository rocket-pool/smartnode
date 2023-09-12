package auction

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type auctionClaimHandler struct {
	lotIndex         uint64
	addressBidAmount *big.Int
	lot              *auction.AuctionLot
}

func (h *auctionClaimHandler) CreateBindings(rp *rocketpool.RocketPool) error {
	var err error
	h.lot, err = auction.NewAuctionLot(rp, h.lotIndex)
	if err != nil {
		return fmt.Errorf("error creating lot %d binding: %w", h.lotIndex, err)
	}
	return nil
}

func (h *auctionClaimHandler) GetState(nodeAddress common.Address, mc *batch.MultiCaller) {
	h.lot.GetLotExists(mc)
	h.lot.GetLotAddressBidAmount(mc, &h.addressBidAmount, nodeAddress)
	h.lot.GetLotIsCleared(mc)
}

func (h *auctionClaimHandler) PrepareResponse(rp *rocketpool.RocketPool, nodeAccount accounts.Account, opts *bind.TransactOpts, response *api.ClaimFromLotResponse) error {
	// Check for validity
	response.DoesNotExist = !h.lot.Details.Exists
	response.NoBidFromAddress = (h.addressBidAmount.Cmp(big.NewInt(0)) == 0)
	response.NotCleared = !h.lot.Details.IsCleared
	response.CanClaim = !(response.DoesNotExist || response.NoBidFromAddress || response.NotCleared)

	// Get tx info
	if response.CanClaim {
		txInfo, err := h.lot.ClaimBid(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		response.TxInfo = txInfo
	}
	return nil
}
