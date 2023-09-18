package auction

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
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
	handler *AuctionHandler
}

func (f *auctionClaimContextFactory) Create(vars map[string]string) (*auctionClaimContext, error) {
	c := &auctionClaimContext{
		handler: f.handler,
	}
	inputErrs := []error{
		server.ValidateArg("index", vars, cliutils.ValidateUint, &c.lotIndex),
	}
	return c, errors.Join(inputErrs...)
}

func (f *auctionClaimContextFactory) RegisterRoute(router *mux.Router) {
	server.RegisterSingleStageRoute[*auctionClaimContext, api.AuctionClaimFromLotData](
		router, "claim-lot", f, f.handler.serviceProvider,
	)
}

// ===============
// === Context ===
// ===============

type auctionClaimContext struct {
	handler     *AuctionHandler
	rp          *rocketpool.RocketPool
	nodeAddress common.Address

	lotIndex         uint64
	addressBidAmount *big.Int
	lot              *auction.AuctionLot
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

func (c *auctionClaimContext) PrepareData(data *api.AuctionClaimFromLotData, opts *bind.TransactOpts) error {
	// Check for validity
	data.DoesNotExist = !c.lot.Details.Exists
	data.NoBidFromAddress = (c.addressBidAmount.Cmp(big.NewInt(0)) == 0)
	data.NotCleared = !c.lot.Details.IsCleared
	data.CanClaim = !(data.DoesNotExist || data.NoBidFromAddress || data.NotCleared)

	// Get tx info
	if data.CanClaim && opts != nil {
		txInfo, err := c.lot.ClaimBid(opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for PlaceBid: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
