package auction

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/gorilla/mux"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/auction"
	"github.com/rocket-pool/rocketpool-go/v2/rocketpool"

	"github.com/rocket-pool/node-manager-core/api/server"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
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
	server.RegisterSingleStageRoute[*auctionRecoverContext, types.DataBatch[api.AuctionRecoverRplFromLotData]](
		router, "lots/recover", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
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

func (c *auctionRecoverContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()

	// Requirements
	status, err := sp.RequireNodeRegistered(c.handler.ctx)
	if err != nil {
		return status, err
	}

	// Bindings
	c.lots = make([]*auction.AuctionLot, len(c.indices))
	for i, index := range c.indices {
		c.lots[i], err = auction.NewAuctionLot(c.rp, index)
		if err != nil {
			return types.ResponseStatus_Error, fmt.Errorf("error creating lot %d binding: %w", index, err)
		}
	}
	return types.ResponseStatus_Success, nil
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

func (c *auctionRecoverContext) PrepareData(dataBatch *types.DataBatch[api.AuctionRecoverRplFromLotData], opts *bind.TransactOpts) (types.ResponseStatus, error) {
	// Get the current block
	ctx := c.handler.ctx
	currentBlock, err := c.rp.Client.BlockNumber(ctx)
	if err != nil {
		return types.ResponseStatus_Error, fmt.Errorf("error getting current EL block: %w", err)
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
				return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for RecoverUnclaimedRpl: %w", err)
			}
			data.TxInfo = txInfo
		}
	}
	return types.ResponseStatus_Success, nil
}
