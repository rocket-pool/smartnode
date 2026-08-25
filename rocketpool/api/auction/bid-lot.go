package auction

import (
	"math/big"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/bindings/auction"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	"github.com/rocket-pool/smartnode/rocketpool/api/response"
	"github.com/rocket-pool/smartnode/rocketpool/api/snroute"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

func canBidOnLot(c *cli.Command, lotIndex uint64, amountWei *big.Int) (*api.CanBidOnLotResponse, error) {

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	w, err := services.GetWallet(c)
	if err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.CanBidOnLotResponse{}

	// Sync
	var wg errgroup.Group

	// Check if lot exists
	wg.Go(func() error {
		lotExists, err := auction.GetLotExists(rp, lotIndex, nil)
		if err == nil {
			response.DoesNotExist = !lotExists
		}
		return err
	})

	// Check if lot bidding has ended
	wg.Go(func() error {
		biddingEnded, err := getLotBiddingEnded(rp, lotIndex)
		if err == nil {
			response.BiddingEnded = biddingEnded
		}
		return err
	})

	// Check lot remaining RPL amount
	wg.Go(func() error {
		remainingRpl, err := auction.GetLotRemainingRPLAmount(rp, lotIndex, nil)
		if err == nil {
			response.RPLExhausted = (remainingRpl.Cmp(big.NewInt(0)) == 0)
		}
		return err
	})

	// Check if lot bidding is enabled
	wg.Go(func() error {
		bidOnLotEnabled, err := protocol.GetBidOnLotEnabled(rp, nil)
		if err == nil {
			response.BidOnLotDisabled = !bidOnLotEnabled
		}
		return err
	})

	// Get gas estimate
	wg.Go(func() error {
		opts, err := w.GetNodeAccountTransactor()
		if err != nil {
			return err
		}
		opts.Value = amountWei
		gasLimits, err := auction.EstimatePlaceBidGas(rp, lotIndex, opts)
		if err == nil {
			response.GasLimits = gasLimits
		}
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	// Update & return response
	response.CanBid = !response.DoesNotExist && !response.BiddingEnded && !response.RPLExhausted && !response.BidOnLotDisabled
	return &response, nil

}

func bidOnLot(c *cli.Command, lotIndex uint64, amountWei *big.Int, t *snroute.TransactOpts) (*api.BidOnLotResponse, error) {
	opts := t.Opts()

	// Get services
	if err := services.RequireNodeWallet(c); err != nil {
		return nil, err
	}
	if err := services.RequireRocketStorage(c); err != nil {
		return nil, err
	}
	rp, err := services.GetRocketPool(c)
	if err != nil {
		return nil, err
	}

	// Response
	response := api.BidOnLotResponse{}

	opts.Value = amountWei

	// Bid on lot
	hash, err := auction.PlaceBid(rp, lotIndex, opts)
	if err != nil {
		return nil, err
	}
	response.TxHash = hash

	// Return response
	return &response, nil

}

func canBidLotHandler(ctx snroute.Context) {
	lotIndex, amountWei, err := parseLotIndexAndAmount(ctx.Request)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := canBidOnLot(ctx.Command(), lotIndex, amountWei)
	response.WriteResponse(ctx.Writer, resp, err)
}

func bidLotHandler(ctx snroute.WriteContext) {
	lotIndex, amountWei, err := parseLotIndexAndAmount(ctx.Request)
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	opts, err := ctx.Transactor()
	if err != nil {
		response.WriteErrorResponse(ctx.Writer, err)
		return
	}
	resp, err := bidOnLot(ctx.Command(), lotIndex, amountWei, opts)
	response.WriteResponse(ctx.Writer, resp, err)
}
