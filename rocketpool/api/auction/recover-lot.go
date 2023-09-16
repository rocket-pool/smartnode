package auction

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	batch "github.com/rocket-pool/batch-query"
	"github.com/rocket-pool/rocketpool-go/auction"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/settings"

	"github.com/rocket-pool/smartnode/rocketpool/common/server"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// ===============
// === Factory ===
// ===============

type auctionRecoverContextFactory struct {
	h *AuctionHandler
}

func (f *auctionRecoverContextFactory) Create(vars map[string]string) (*auctionRecoverContext, error) {
	c := &auctionRecoverContext{
		h: f.h,
	}
	inputErrs := []error{
		server.ValidateArg("index", vars, cliutils.ValidateUint, &c.lotIndex),
	}
	return c, errors.Join(inputErrs...)
}

func (f *auctionRecoverContextFactory) Run(c *auctionRecoverContext) (*api.ApiResponse[api.RecoverRplFromLotData], error) {
	return runAuctionCall[api.RecoverRplFromLotData](c)
}

// ===============
// === Context ===
// ===============

type auctionRecoverContext struct {
	h    *AuctionHandler
	rp   *rocketpool.RocketPool
	opts *bind.TransactOpts

	lotIndex  uint64
	lot       *auction.AuctionLot
	pSettings *settings.ProtocolDaoSettings
}

func (c *auctionRecoverContext) CreateBindings(ctx *callContext) error {
	var err error
	c.rp = ctx.rp
	c.opts = ctx.opts

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

func (c *auctionRecoverContext) GetState(mc *batch.MultiCaller) {
	c.lot.GetLotExists(mc)
	c.lot.GetLotEndBlock(mc)
	c.lot.GetLotRemainingRplAmount(mc)
	c.lot.GetLotRplRecovered(mc)
	c.pSettings.GetBidOnAuctionLotEnabled(mc)
}

func (c *auctionRecoverContext) PrepareData(data *api.RecoverRplFromLotData) error {
	// Get the current block
	currentBlock, err := c.rp.Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("error getting current EL block: %w", err)
	}

	// Check for validity
	data.DoesNotExist = !c.lot.Details.Exists
	data.BiddingNotEnded = !(currentBlock >= c.lot.Details.EndBlock.Formatted())
	data.NoUnclaimedRpl = (c.lot.Details.RemainingRplAmount.Cmp(big.NewInt(0)) == 0)
	data.RplAlreadyRecovered = c.lot.Details.RplRecovered
	data.CanRecover = !(data.DoesNotExist || data.BiddingNotEnded || data.NoUnclaimedRpl || data.RplAlreadyRecovered)

	// Get tx info
	if data.CanRecover && c.opts != nil {
		txInfo, err := c.lot.RecoverUnclaimedRpl(c.opts)
		if err != nil {
			return fmt.Errorf("error getting TX info for RecoverUnclaimedRpl: %w", err)
		}
		data.TxInfo = txInfo
	}
	return nil
}
