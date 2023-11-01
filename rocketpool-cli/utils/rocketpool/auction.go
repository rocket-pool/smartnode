package rocketpool

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type AuctionRequester struct {
	client *http.Client
	route  string
}

func NewAuctionRequester(client *http.Client) *AuctionRequester {
	return &AuctionRequester{
		client: client,
		route:  "auction",
	}
}

// Bid on a lot
func (r *AuctionRequester) BidOnLot(lotIndex uint64, amountWei *big.Int) (*api.ApiResponse[api.AuctionBidOnLotData], error) {
	method := "bid-lot"
	args := map[string]string{
		"index":  fmt.Sprint(lotIndex),
		"amount": amountWei.String(),
	}
	response, err := SendGetRequest[api.AuctionBidOnLotData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Auction BidOnLot request: %w", err)
	}
	return response, nil
}

// Claim RPL from a lot
func (r *AuctionRequester) ClaimFromLot(lotIndex uint64) (*api.ApiResponse[api.AuctionClaimFromLotData], error) {
	method := "claim-lot"
	args := map[string]string{
		"index": fmt.Sprint(lotIndex),
	}
	response, err := SendGetRequest[api.AuctionClaimFromLotData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Auction ClaimFromLot request: %w", err)
	}
	return response, nil
}

// Create a new lot
func (r *AuctionRequester) CreateLot() (*api.ApiResponse[api.AuctionCreateLotData], error) {
	method := "create-lot"
	args := map[string]string{}
	response, err := SendGetRequest[api.AuctionCreateLotData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Auction CreateLot request: %w", err)
	}
	return response, nil
}

// Get RPL lots for auction
func (r *AuctionRequester) Lots() (*api.ApiResponse[api.AuctionLotsData], error) {
	method := "lots"
	args := map[string]string{}
	response, err := SendGetRequest[api.AuctionLotsData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Auction Lots request: %w", err)
	}
	return response, nil
}

// Recover unclaimed RPL from a lot (returning it to the auction contract)
func (r *AuctionRequester) RecoverUnclaimedRplFromLot(lotIndex uint64) (*api.ApiResponse[api.AuctionRecoverRplFromLotData], error) {
	method := "recover-lot"
	args := map[string]string{
		"index": fmt.Sprint(lotIndex),
	}
	response, err := SendGetRequest[api.AuctionRecoverRplFromLotData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Auction RecoverUnclaimedRplFromLot request: %w", err)
	}
	return response, nil
}

// Get RPL auction status
func (r *AuctionRequester) Status() (*api.ApiResponse[api.AuctionStatusData], error) {
	method := "status"
	args := map[string]string{}
	response, err := SendGetRequest[api.AuctionStatusData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Auction Status request: %w", err)
	}
	return response, nil
}
