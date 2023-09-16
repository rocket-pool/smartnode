package auction

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type auctionLotContextFactory struct {
	h *AuctionHandler
}

func (f *auctionLotContextFactory) Create(vars map[string]string) (*auctionLotContext, error) {
	c := &auctionLotContext{
		h: f.h,
	}
	return c, nil
}

func (f *auctionLotContextFactory) Run(c *auctionLotContext) (*api.ApiResponse[api.AuctionLotsData], error) {
	return runAuctionCall[api.AuctionLotsData](c)
}

// ===============
// === Context ===
// ===============

type auctionLotContext struct {
	h           *AuctionHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address
	opts        *bind.TransactOpts

	auctionMgr *auction.AuctionManager
}

func (c *auctionLotContext) CreateBindings(ctx *callContext) error {
	var err error
	c.rp = ctx.rp
	c.nodeAddress = ctx.nodeAddress
	c.opts = ctx.opts

	c.auctionMgr, err = auction.NewAuctionManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating auction manager binding: %w", err)
	}
	return nil
}

func (c *auctionLotContext) GetState(mc *batch.MultiCaller) {
	c.auctionMgr.GetLotCount(mc)
}

func (c *auctionLotContext) PrepareData(data *api.AuctionLotsData) error {
	// Get lot details
	lotCount := c.auctionMgr.Details.LotCount.Formatted()
	lots := make([]*auction.AuctionLot, lotCount)
	details := make([]api.AuctionLotDetails, lotCount)

	// Load details
	err := c.rp.BatchQuery(int(lotCount), int(lotCountDetailsBatchSize), func(mc *batch.MultiCaller, i int) error {
		lot, err := auction.NewAuctionLot(c.rp, uint64(i))
		if err != nil {
			return fmt.Errorf("error creating lot %d binding: %w", i, err)
		}
		lots[i] = lot
		lot.GetAllDetails(mc)
		lot.GetLotAddressBidAmount(mc, &details[i].NodeBidAmount, c.nodeAddress)
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
		fullDetails.RplRecoveryAvailable = (fullDetails.Details.IsCleared && hasRemainingRpl && !fullDetails.Details.RplRecovered)
	}
	data.Lots = details
	return nil
}
