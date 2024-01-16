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
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/core"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/input"
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
	server.RegisterSingleStageRoute[*auctionClaimContext, api.DataBatch[api.AuctionClaimFromLotData]](
		router, "lots/claim", f, f.handler.serviceProvider,
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

func (c *auctionClaimContext) Initialize() error {
	sp := c.handler.serviceProvider
	c.rp = sp.GetRocketPool()
	c.nodeAddress, _ = sp.GetWallet().GetAddress()

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

func (c *auctionClaimContext) GetState(mc *batch.MultiCaller) {
	for i, lot := range c.lots {
		core.AddQueryablesToMulticall(mc,
			lot.Exists,
			lot.IsCleared,
		)
		lot.GetLotAddressBidAmount(mc, &c.addressBidAmounts[i], c.nodeAddress)
	}
}

func (c *auctionClaimContext) PrepareData(dataBatch *api.DataBatch[api.AuctionClaimFromLotData], opts *bind.TransactOpts) error {
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
				return fmt.Errorf("error getting TX info for PlaceBid: %w", err)
			}
			data.TxInfo = txInfo
		}
	}
	return nil
}
