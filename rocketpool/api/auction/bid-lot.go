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

// ===============
// === Factory ===
// ===============

type auctionBidContextFactory struct {
	h *AuctionHandler
}

func (f *auctionBidContextFactory) Create(vars map[string]string) (*auctionBidContext, error) {
	c := &auctionBidContext{
		h: f.h,
	}
	inputErrs := []error{
		server.ValidateArg("index", vars, cliutils.ValidateUint, &c.lotIndex),
		server.ValidateArg("amount", vars, cliutils.ValidatePositiveWeiAmount, &c.amountWei),
	}
	return c, errors.Join(inputErrs...)
}

func (f *auctionBidContextFactory) Run(c *auctionBidContext) (*api.ApiResponse[api.BidOnLotData], error) {
	return runAuctionCall[api.BidOnLotData](c)
}

// ===============
// === Context ===
// ===============

type auctionBidContext struct {
	h         *AuctionHandler
	lotIndex  uint64
	amountWei *big.Int
	lot       *auction.AuctionLot
	pSettings *settings.ProtocolDaoSettings
	*commonContext
}

func (c *auctionBidContext) CreateBindings(ctx *commonContext) error {
	var err error
	c.commonContext = ctx

	c.lot, err = auction.NewAuctionLot(c.rp, c.lotIndex)
	if err != nil {
		return fmt.Errorf("error creating lot %d binding: %w", c.lotIndex, err)
	}
	c.pSettings, err = settings.NewProtocolDaoSettings(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO settings binding: %w", err)
	}
	return nil
}

func (c *auctionBidContext) GetState(mc *batch.MultiCaller) {
	c.lot.GetLotExists(mc)
	c.lot.GetLotEndBlock(mc)
	c.lot.GetLotRemainingRplAmount(mc)
	c.pSettings.GetBidOnAuctionLotEnabled(mc)
}

func (c *auctionBidContext) PrepareData(Data *api.BidOnLotData) error {
	// Get the current block
	currentBlock, err := c.rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Check for validity
	Data.DoesNotExist = !c.lot.Details.Exists
	Data.BiddingEnded = (currentBlock >= c.lot.Details.EndBlock.Formatted())
	Data.RplExhausted = (c.lot.Details.RemainingRplAmount.Cmp(big.NewInt(0)) == 0)
	Data.BidOnLotDisabled = !c.pSettings.Details.Auction.IsBidOnLotEnabled
	Data.CanBid = !(Data.DoesNotExist || Data.BiddingEnded || Data.RplExhausted || Data.BidOnLotDisabled)

	// Get tx info
	if Data.CanBid && c.opts != nil {
		txInfo, err := c.lot.PlaceBid(c.opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		Data.TxInfo = txInfo
	}
	return nil
}
