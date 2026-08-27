package rocketpool

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get RPL auction status
func (c *Client) AuctionStatus() (api.AuctionStatusResponse, error) {
	response, err := c.callAPI[api.AuctionStatusResponse]("GET", "/api/auction/status", nil, "Could not get auction status")
	if err != nil {
		return response, err
	}
	if response.TotalRPLBalance == nil {
		response.TotalRPLBalance = big.NewInt(0)
	}
	if response.AllottedRPLBalance == nil {
		response.AllottedRPLBalance = big.NewInt(0)
	}
	if response.RemainingRPLBalance == nil {
		response.RemainingRPLBalance = big.NewInt(0)
	}
	return response, nil
}

// Get RPL lots for auction
func (c *Client) AuctionLots() (api.AuctionLotsResponse, error) {
	response, err := c.callAPI[api.AuctionLotsResponse]("GET", "/api/auction/lots", nil, "Could not get auction lots")
	if err != nil {
		return response, err
	}
	for i := 0; i < len(response.Lots); i++ {
		details := &response.Lots[i].Details
		if details.StartPrice == nil {
			details.StartPrice = big.NewInt(0)
		}
		if details.ReservePrice == nil {
			details.ReservePrice = big.NewInt(0)
		}
		if details.PriceAtCurrentBlock == nil {
			details.PriceAtCurrentBlock = big.NewInt(0)
		}
		if details.PriceByTotalBids == nil {
			details.PriceByTotalBids = big.NewInt(0)
		}
		if details.CurrentPrice == nil {
			details.CurrentPrice = big.NewInt(0)
		}
		if details.TotalRPLAmount == nil {
			details.TotalRPLAmount = big.NewInt(0)
		}
		if details.ClaimedRPLAmount == nil {
			details.ClaimedRPLAmount = big.NewInt(0)
		}
		if details.RemainingRPLAmount == nil {
			details.RemainingRPLAmount = big.NewInt(0)
		}
		if details.TotalBidAmount == nil {
			details.TotalBidAmount = big.NewInt(0)
		}
		if details.AddressBidAmount == nil {
			details.AddressBidAmount = big.NewInt(0)
		}
	}
	return response, nil
}

// Check whether the node can create a new lot
func (c *Client) CanCreateLot() (api.CanCreateLotResponse, error) {
	return c.callAPI[api.CanCreateLotResponse]("GET", "/api/auction/can-create-lot", nil, "Could not get can create lot status")
}

// Create a new lot
func (c *Client) CreateLot() (api.CreateLotResponse, error) {
	return c.callAPI[api.CreateLotResponse]("POST", "/api/auction/create-lot", nil, "Could not create lot")
}

// Check whether the node can bid on a lot
func (c *Client) CanBidOnLot(lotIndex uint64, amountWei *big.Int) (api.CanBidOnLotResponse, error) {
	return c.callAPI[api.CanBidOnLotResponse]("GET", "/api/auction/can-bid-lot", url.Values{
		"lotIndex":  {fmt.Sprintf("%d", lotIndex)},
		"amountWei": {amountWei.String()},
	}, "Could not get can bid on lot status")
}

// Bid on a lot
func (c *Client) BidOnLot(lotIndex uint64, amountWei *big.Int) (api.BidOnLotResponse, error) {
	return c.callAPI[api.BidOnLotResponse]("POST", "/api/auction/bid-lot", url.Values{
		"lotIndex":  {fmt.Sprintf("%d", lotIndex)},
		"amountWei": {amountWei.String()},
	}, "Could not bid on lot")
}

// Check whether the node can claim RPL from a lot
func (c *Client) CanClaimFromLot(lotIndex uint64) (api.CanClaimFromLotResponse, error) {
	return c.callAPI[api.CanClaimFromLotResponse]("GET", "/api/auction/can-claim-lot", url.Values{
		"lotIndex": {fmt.Sprintf("%d", lotIndex)},
	}, "Could not get can claim RPL from lot status")
}

// Claim RPL from a lot
func (c *Client) ClaimFromLot(lotIndex uint64) (api.ClaimFromLotResponse, error) {
	return c.callAPI[api.ClaimFromLotResponse]("POST", "/api/auction/claim-lot", url.Values{
		"lotIndex": {fmt.Sprintf("%d", lotIndex)},
	}, "Could not claim RPL from lot")
}

// Check whether the node can recover unclaimed RPL from a lot
func (c *Client) CanRecoverUnclaimedRPLFromLot(lotIndex uint64) (api.CanRecoverRPLFromLotResponse, error) {
	return c.callAPI[api.CanRecoverRPLFromLotResponse]("GET", "/api/auction/can-recover-lot", url.Values{
		"lotIndex": {fmt.Sprintf("%d", lotIndex)},
	}, "Could not get can recover unclaimed RPL from lot status")
}

// Recover unclaimed RPL from a lot (returning it to the auction contract)
func (c *Client) RecoverUnclaimedRPLFromLot(lotIndex uint64) (api.RecoverRPLFromLotResponse, error) {
	return c.callAPI[api.RecoverRPLFromLotResponse]("POST", "/api/auction/recover-lot", url.Values{
		"lotIndex": {fmt.Sprintf("%d", lotIndex)},
	}, "Could not recover unclaimed RPL from lot")
}
