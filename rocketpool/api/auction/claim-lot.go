package auction

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/rocketpool"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// ===============
// === Factory ===
// ===============

type auctionClaimContextFactory struct {
	h *AuctionHandler
}

func (f *auctionClaimContextFactory) Create(vars map[string]string) (*auctionClaimContext, error) {
	c := &auctionClaimContext{
		h: f.h,
	}
	inputErrs := []error{
		server.ValidateArg("index", vars, cliutils.ValidateUint, &c.lotIndex),
	}
	return c, errors.Join(inputErrs...)
}

func (f *auctionClaimContextFactory) Run(c *auctionClaimContext) (*api.ApiResponse[api.ClaimFromLotData], error) {
	return runAuctionCall[api.ClaimFromLotData](c)
}

// ===============
// === Context ===
// ===============

type auctionClaimContext struct {
	h           *AuctionHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address
	opts        *bind.TransactOpts

	lotIndex         uint64
	addressBidAmount *big.Int
	lot              *auction.AuctionLot
}

func (c *auctionClaimContext) CreateBindings(ctx *callContext) error {
	var err error
	c.rp = ctx.rp
	c.nodeAddress = ctx.nodeAddress
	c.opts = ctx.opts

	c.lot, err = auction.NewAuctionLot(c.rp, c.lotIndex)
	if err != nil {
		return fmt.Errorf("error creating lot %d binding: %w", c.lotIndex, err)
	}
	return nil
}

func (c *auctionClaimContext) GetState(mc *batch.MultiCaller) {
	c.lot.GetLotExists(mc)
	c.lot.GetLotAddressBidAmount(mc, &c.addressBidAmount, c.nodeAddress)
	c.lot.GetLotIsCleared(mc)
}

func (c *auctionClaimContext) PrepareData(data *api.ClaimFromLotData) error {
	// Check for validity
	data.DoesNotExist = !c.lot.Details.Exists
	data.NoBidFromAddress = (c.addressBidAmount.Cmp(big.NewInt(0)) == 0)
	data.NotCleared = !c.lot.Details.IsCleared
	data.CanClaim = !(data.DoesNotExist || data.NoBidFromAddress || data.NotCleared)

	// Get tx info
	if data.CanClaim && c.opts != nil {
		txInfo, err := c.lot.ClaimBid(c.opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
