package auction

import (
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
	claimBatchSize int = 100
)

// ===============
// === Factory ===
// ===============

type auctionClaimContextFactory struct {
	handler *AuctionHandler
}

func (f *auctionClaimContextFactory) Create(args url.Values) (*auctionClaimContext, error) {
	c := &auctionClaimContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArgBatch("indices", args, claimBatchSize, input.ValidateUint, &c.indices),
	}
	return c, errors.Join(inputErrs...)
}

func (f *auctionClaimContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*auctionClaimContext, types.DataBatch[api.AuctionClaimFromLotData]](
		router, "lots/claim", f, f.handler.logger.Logger, f.handler.serviceProvider.ServiceProvider,
	)
}

// ===============
// === Context ===
// ===============

type auctionClaimContext struct {
	handler     *AuctionHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	indices           []uint64
	addressBidAmounts []*big.Int
	lots              []*auction.AuctionLot
}

func (c *auctionClaimContext) Initialize() (types.ResponseStatus, error) {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

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

func (c *auctionClaimContext) GetState(mc *batch.MultiCaller) {
	for i, lot := range c.lots {
		eth.AddQueryablesToMulticall(mc,
			lot.Exists,
			lot.IsCleared,
		)
		lot.GetLotAddressBidAmount(mc, &c.addressBidAmounts[i], c.nodeAddress)
	}
}

func (c *auctionClaimContext) PrepareData(dataBatch *types.DataBatch[api.AuctionClaimFromLotData], opts *bind.TransactOpts) (types.ResponseStatus, error) {
	dataBatch.Batch = make([]api.AuctionClaimFromLotData, len(c.indices))
	for i, lot := range c.lots {
		addressBidAmount := c.addressBidAmounts[i]

		// Check for validity
		data := &dataBatch.Batch[i]
		data.DoesNotExist = !lot.Exists.Get()
		data.NoBidFromAddress = (addressBidAmount.Cmp(big.NewInt(0)) == 0)
		data.NotCleared = !lot.IsCleared.Get()
		data.CanClaim = !(data.DoesNotExist || data.NoBidFromAddress || data.NotCleared)

		// Get tx info
		if data.CanClaim && opts != nil {
			txInfo, err := lot.ClaimBid(opts)
			if err != nil {
				return types.ResponseStatus_Error, fmt.Errorf("error getting TX info for PlaceBid: %w", err)
			}
			data.TxInfo = txInfo
		}
	}
	return types.ResponseStatus_Success, nil
}
