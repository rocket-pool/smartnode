package auction

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/rocketpool-go/v2/auction"
	"github.com/rocket-pool/rocketpool-go/v2/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
		router, "lots/bid", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
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

func (c *auctionBidContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.lot, err = auction.NewAuctionLot(c.rp, c.lotIndex)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating lot %d binding: %w", c.lotIndex, err)
	}
	pMgr, err := protocol.NewProtocolDaoManager(c.rp)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error creating pDAO manager binding: %w", err)
	}
	c.pSettings = pMgr.Settings
	return types.ResponseStatus_Success, nil
}

func (c *auctionBidContext) GetState(mc *batch.MultiCaller) {
	eth.AddQueryablesToMulticall(mc,
		c.lot.Exists,
		c.lot.EndBlock,
		c.lot.RemainingRplAmount,
		c.pSettings.Auction.IsBidOnLotEnabled,
	)
}

func (c *auctionBidContext) PrepareData(data *api.AuctionBidOnLotData, opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get the current block
	ctx := c.handler.ctx
	currentBlock, err := c.rp.Client.BlockNumber(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting current EL block: %w", err)
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
			return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		data.TxInfo = txInfo
	}
	return types.ResponseStatus_Success, nil
}
