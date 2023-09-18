package auction

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type auctionBidContextFactory struct {
	handler *AuctionHandler
}

func (f *auctionBidContextFactory) Create(vars map[string]string) (*auctionBidContext, error) {
	c := &auctionBidContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("index", vars, input.ValidateUint, &c.lotIndex),
		server.ValidateArg("amount", vars, input.ValidatePositiveWeiAmount, &c.amountWei),
	}
	return c, errors.Join(inputErrs...)
}

func (f *auctionBidContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*auctionBidContext, api.AuctionBidOnLotData](
		router, "bid-lot", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type auctionBidContext struct {
	handler *AuctionHandler
	rp      *rocketpool.RocketPool

	lotIndex  uint64
	amountWei *big.Int
	lot       *auction.AuctionLot
	pSettings *settings.ProtocolDaoSettings
}

func (c *auctionBidContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	err := sp.RequireNodeRegistered()
	if err != nil {
		return err
	}

	// Bindings
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

func (c *auctionBidContext) PrepareData(data *api.AuctionBidOnLotData, opts *bind.TransactOpts) error {
	// Get the current block
	currentBlock, err := c.rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Check for validity
	data.DoesNotExist = !c.lot.Details.Exists
	data.BiddingEnded = (currentBlock >= c.lot.Details.EndBlock.Formatted())
	data.RplExhausted = (c.lot.Details.RemainingRplAmount.Cmp(big.NewInt(0)) == 0)
	data.BidOnLotDisabled = !c.pSettings.Details.Auction.IsBidOnLotEnabled
	data.CanBid = !(data.DoesNotExist || data.BiddingEnded || data.RplExhausted || data.BidOnLotDisabled)

	// Get tx info
	if data.CanBid && opts != nil {
		txInfo, err := c.lot.PlaceBid(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
