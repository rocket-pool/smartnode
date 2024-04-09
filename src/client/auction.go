package client

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type AuctionRequester struct {
	context *client.RequesterContext
}

func NewAuctionRequester(context *client.RequesterContext) *AuctionRequester {
	return &AuctionRequester{
		context: context,
	}
}

func (r *AuctionRequester) GetName() string {
	return "Auction"
}
func (r *AuctionRequester) GetRoute() string {
	return "auction"
}
func (r *AuctionRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Bid on a lot
func (r *AuctionRequester) BidOnLot(lotIndex uint64, amountWei *big.Int) (*types.ApiResponse[api.AuctionBidOnLotData], error) {
	args := map[string]string{
		"index":  fmt.Sprint(lotIndex),
		"amount": amountWei.String(),
	}
	return client.SendGetRequest[api.AuctionBidOnLotData](r, "lots/bid", "BidOnLot", args)
}

// Claim RPL from lots
func (r *AuctionRequester) ClaimFromLots(indices []uint64) (*types.ApiResponse[types.DataBatch[api.AuctionClaimFromLotData]], error) {
	args := map[string]string{
		"indices": client.MakeBatchArg(indices),
	}
	return client.SendGetRequest[types.DataBatch[api.AuctionClaimFromLotData]](r, "lots/claim", "ClaimFromLots", args)
}

// Create a new lot
func (r *AuctionRequester) CreateLot() (*types.ApiResponse[api.AuctionCreateLotData], error) {
	return client.SendGetRequest[api.AuctionCreateLotData](r, "lots/create", "CreateLot", nil)
}

// Get RPL lots for auction
func (r *AuctionRequester) Lots() (*types.ApiResponse[api.AuctionLotsData], error) {
	return client.SendGetRequest[api.AuctionLotsData](r, "lots", "Lots", nil)
}

// Recover unclaimed RPL from lots (returning it to the auction contract)
func (r *AuctionRequester) RecoverUnclaimedRplFromLots(indices []uint64) (*types.ApiResponse[types.DataBatch[api.AuctionRecoverRplFromLotData]], error) {
	args := map[string]string{
		"indices": client.MakeBatchArg(indices),
	}
	return client.SendGetRequest[types.DataBatch[api.AuctionRecoverRplFromLotData]](r, "lots/recover", "RecoverUnclaimedRplFromLots", args)
}

// Get RPL auction status
func (r *AuctionRequester) Status() (*types.ApiResponse[api.AuctionStatusData], error) {
	return client.SendGetRequest[api.AuctionStatusData](r, "status", "Status", nil)
}
