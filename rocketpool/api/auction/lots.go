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

type auctionLotHandler struct {
	auctionMgr *auction.AuctionManager
}

func (h *auctionLotHandler) CreateBindings(rp *rocketpool.RocketPool) error {
	var err error
	h.auctionMgr, err = auction.NewAuctionManager(rp)
	if err != nil {
		return fmt.Errorf("error creating auction manager binding: %w", err)
	}
	return nil
}

func (h *auctionLotHandler) GetState(nodeAddress common.Address, mc *batch.MultiCaller) {
	h.auctionMgr.GetLotCount(mc)
}

func (h *auctionLotHandler) PrepareResponse(rp *rocketpool.RocketPool, nodeAccount accounts.Account, opts *bind.TransactOpts, response *api.AuctionLotsResponse) error {
	// Get lot details
	lotCount := h.auctionMgr.Details.LotCount.Formatted()
	lots := make([]*auction.AuctionLot, lotCount)
	details := make([]api.LotDetails, lotCount)

	// Load details
	err := rp.BatchQuery(int(lotCount), int(lotCountDetailsBatchSize), func(mc *batch.MultiCaller, i int) error {
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
	response.Lots = details
	return nil
}
