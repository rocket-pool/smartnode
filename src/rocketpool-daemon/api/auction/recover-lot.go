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
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool-daemon/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
)

const (
	recoverBatchSize int = 100
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
		server.ValidateArgBatch("indices", args, recoverBatchSize, input.ValidateUint, &c.indices),
	}
	return c, errors.Join(inputErrs...)
}

func (f *auctionRecoverContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*auctionRecoverContext, api.DataBatch[api.AuctionRecoverRplFromLotData]](
		router, "lots/recover", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type auctionRecoverContext struct {
	handler *AuctionHandler
	rp      *rocketpool.RocketPool

	indices []uint64
	lots    []*auction.AuctionLot
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
	c.lots = make([]*auction.AuctionLot, len(c.indices))
	for i, index := range c.indices {
		c.lots[i], err = auction.NewAuctionLot(c.rp, index)
		if err != nil {
			return fmt.Errorf("error creating lot %d binding: %w", index, err)
		}
	}
	return nil
}

func (c *auctionRecoverContext) GetState(mc *batch.MultiCaller) {
	for _, lot := range c.lots {
		eth.AddQueryablesToMulticall(mc,
			lot.Exists,
			lot.EndBlock,
			lot.RemainingRplAmount,
			lot.RplRecovered,
		)
	}
}

func (c *auctionRecoverContext) PrepareData(dataBatch *api.DataBatch[api.AuctionRecoverRplFromLotData], opts *bind.TransactOpts) error {
	// Get the current block
	currentBlock, err := c.rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	dataBatch.Batch = make([]api.AuctionRecoverRplFromLotData, len(c.indices))
	for i, lot := range c.lots {
		// Check for validity
		data := &dataBatch.Batch[i]
		data.DoesNotExist = !lot.Exists.Get()
		data.BiddingNotEnded = !(currentBlock >= lot.EndBlock.Formatted())
		data.NoUnclaimedRpl = (lot.RemainingRplAmount.Get().Cmp(big.NewInt(0)) == 0)
		data.RplAlreadyRecovered = lot.RplRecovered.Get()
		data.CanRecover = !(data.DoesNotExist || data.BiddingNotEnded || data.NoUnclaimedRpl || data.RplAlreadyRecovered)

		// Get tx info
		if data.CanRecover && opts != nil {
			txInfo, err := lot.RecoverUnclaimedRpl(opts)
			if err != nil {
				return fmt.Errorf("error getting TX info for RecoverUnclaimedRpl: %w", err)
			}
			data.TxInfo = txInfo
		}
	}
	return nil
}
