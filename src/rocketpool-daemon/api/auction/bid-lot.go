package auction

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// ===============
// === Factory ===
// ===============

type auctionBidContextFactory struct {
	handler *AuctionHandler
}

func (f *auctionBidContextFactory) Create(args url.Values) (*auctionBidContext, error) {
	c := &auctionBidContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("index", args, input.ValidateUint, &c.lotIndex),
		server.ValidateArg("amount", args, input.ValidatePositiveWeiAmount, &c.amountWei),
	}
	return c, errors.Join(inputErrs...)
}

func (f *auctionBidContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*auctionBidContext, api.AuctionBidOnLotData](
		router, "lots/bid", f, f.handler.serviceProvider.ServiceProvider,
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
	pSettings *protocol.ProtocolDaoSettings
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
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	return nil
}

func (c *auctionBidContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.lot.Exists,
		c.lot.EndBlock,
		c.lot.RemainingRplAmount,
		c.pSettings.Auction.IsBidOnLotEnabled,
	)
}

func (c *auctionBidContext) PrepareData(data *api.AuctionBidOnLotData, opts *bind.TransactOpts) error {
	// Get the current block
	currentBlock, err := c.rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Check for validity
	data.DoesNotExist = !c.lot.Exists.Get()
	data.BiddingEnded = (currentBlock >= c.lot.EndBlock.Formatted())
	data.RplExhausted = (c.lot.RemainingRplAmount.Get().Cmp(big.NewInt(0)) == 0)
	data.BidOnLotDisabled = !c.pSettings.Auction.IsBidOnLotEnabled.Get()
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
