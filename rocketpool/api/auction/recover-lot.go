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
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

// ===============
// === Factory ===
// ===============

type auctionRecoverContextFactory struct {
	handler *AuctionHandler
}

func (f *auctionRecoverContextFactory) Create(args url.Values) (*auctionRecoverContext, error) {
	c := &auctionRecoverContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("index", args, input.ValidateUint, &c.lotIndex),
	}
	return c, errors.Join(inputErrs...)
}

func (f *auctionRecoverContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*auctionRecoverContext, api.AuctionRecoverRplFromLotData](
		router, "recover-lot", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type auctionRecoverContext struct {
	handler *AuctionHandler
	rp      *rocketpool.RocketPool

	lotIndex uint64
	lot      *auction.AuctionLot
}

func (c *auctionRecoverContext) Initialize() error {
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
	return nil
}

func (c *auctionRecoverContext) GetState(mc *batch.MultiCaller) {
	core.AddQueryablesToMulticall(mc,
		c.lot.Exists,
		c.lot.EndBlock,
		c.lot.RemainingRplAmount,
		c.lot.RplRecovered,
	)
}

func (c *auctionRecoverContext) PrepareData(data *api.AuctionRecoverRplFromLotData, opts *bind.TransactOpts) error {
	// Get the current block
	currentBlock, err := c.rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Check for validity
	data.DoesNotExist = !c.lot.Exists.Get()
	data.BiddingNotEnded = !(currentBlock >= c.lot.EndBlock.Formatted())
	data.NoUnclaimedRpl = (c.lot.RemainingRplAmount.Get().Cmp(big.NewInt(0)) == 0)
	data.RplAlreadyRecovered = c.lot.RplRecovered.Get()
	data.CanRecover = !(data.DoesNotExist || data.BiddingNotEnded || data.NoUnclaimedRpl || data.RplAlreadyRecovered)

	// Get tx info
	if data.CanRecover && opts != nil {
		txInfo, err := c.lot.RecoverUnclaimedRpl(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for RecoverUnclaimedRpl: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
