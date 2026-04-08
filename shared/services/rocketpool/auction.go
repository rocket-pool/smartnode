package rocketpool

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get RPL auction status
func (c *Client) AuctionStatus() (api.AuctionStatusResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/auction/status", nil)
	if err != nil {
		return api.AuctionStatusResponse{}, fmt.Errorf("Could not get auction status: %w", err)
	}
	var response api.AuctionStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.AuctionStatusResponse{}, fmt.Errorf("Could not decode auction stats response: %w", err)
	}
	if response.Error != "" {
		return api.AuctionStatusResponse{}, fmt.Errorf("Could not get auction status: %s", response.Error)
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
	responseBytes, err := c.callHTTPAPI("GET", "/api/auction/lots", nil)
	if err != nil {
		return api.AuctionLotsResponse{}, fmt.Errorf("Could not get auction lots: %w", err)
	}
	var response api.AuctionLotsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.AuctionLotsResponse{}, fmt.Errorf("Could not decode auction lots response: %w", err)
	}
	if response.Error != "" {
		return api.AuctionLotsResponse{}, fmt.Errorf("Could not get auction lots: %s", response.Error)
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
	responseBytes, err := c.callHTTPAPI("GET", "/api/auction/can-create-lot", nil)
	if err != nil {
		return api.CanCreateLotResponse{}, fmt.Errorf("Could not get can create lot status: %w", err)
	}
	var response api.CanCreateLotResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanCreateLotResponse{}, fmt.Errorf("Could not decode can create lot response: %w", err)
	}
	if response.Error != "" {
		return api.CanCreateLotResponse{}, fmt.Errorf("Could not get can create lot status: %s", response.Error)
	}
	return response, nil
}

// Create a new lot
func (c *Client) CreateLot() (api.CreateLotResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/auction/create-lot", nil)
	if err != nil {
		return api.CreateLotResponse{}, fmt.Errorf("Could not create lot: %w", err)
	}
	var response api.CreateLotResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CreateLotResponse{}, fmt.Errorf("Could not decode create lot response: %w", err)
	}
	if response.Error != "" {
		return api.CreateLotResponse{}, fmt.Errorf("Could not create lot: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can bid on a lot
func (c *Client) CanBidOnLot(lotIndex uint64, amountWei *big.Int) (api.CanBidOnLotResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/auction/can-bid-lot", url.Values{
		"lotIndex":  {fmt.Sprintf("%d", lotIndex)},
		"amountWei": {amountWei.String()},
	})
	if err != nil {
		return api.CanBidOnLotResponse{}, fmt.Errorf("Could not get can bid on lot status: %w", err)
	}
	var response api.CanBidOnLotResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanBidOnLotResponse{}, fmt.Errorf("Could not decode can bid on lot response: %w", err)
	}
	if response.Error != "" {
		return api.CanBidOnLotResponse{}, fmt.Errorf("Could not get can bid on lot status: %s", response.Error)
	}
	return response, nil
}

// Bid on a lot
func (c *Client) BidOnLot(lotIndex uint64, amountWei *big.Int) (api.BidOnLotResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/auction/bid-lot", url.Values{
		"lotIndex":  {fmt.Sprintf("%d", lotIndex)},
		"amountWei": {amountWei.String()},
	})
	if err != nil {
		return api.BidOnLotResponse{}, fmt.Errorf("Could not bid on lot: %w", err)
	}
	var response api.BidOnLotResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.BidOnLotResponse{}, fmt.Errorf("Could not decode bid on lot response: %w", err)
	}
	if response.Error != "" {
		return api.BidOnLotResponse{}, fmt.Errorf("Could not bid on lot: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can claim RPL from a lot
func (c *Client) CanClaimFromLot(lotIndex uint64) (api.CanClaimFromLotResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/auction/can-claim-lot", url.Values{
		"lotIndex": {fmt.Sprintf("%d", lotIndex)},
	})
	if err != nil {
		return api.CanClaimFromLotResponse{}, fmt.Errorf("Could not get can claim RPL from lot status: %w", err)
	}
	var response api.CanClaimFromLotResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanClaimFromLotResponse{}, fmt.Errorf("Could not decode can claim RPL from lot response: %w", err)
	}
	if response.Error != "" {
		return api.CanClaimFromLotResponse{}, fmt.Errorf("Could not get can claim RPL from lot status: %s", response.Error)
	}
	return response, nil
}

// Claim RPL from a lot
func (c *Client) ClaimFromLot(lotIndex uint64) (api.ClaimFromLotResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/auction/claim-lot", url.Values{
		"lotIndex": {fmt.Sprintf("%d", lotIndex)},
	})
	if err != nil {
		return api.ClaimFromLotResponse{}, fmt.Errorf("Could not claim RPL from lot: %w", err)
	}
	var response api.ClaimFromLotResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ClaimFromLotResponse{}, fmt.Errorf("Could not decode claim RPL from lot response: %w", err)
	}
	if response.Error != "" {
		return api.ClaimFromLotResponse{}, fmt.Errorf("Could not claim RPL from lot: %s", response.Error)
	}
	return response, nil
}

// Check whether the node can recover unclaimed RPL from a lot
func (c *Client) CanRecoverUnclaimedRPLFromLot(lotIndex uint64) (api.CanRecoverRPLFromLotResponse, error) {
	responseBytes, err := c.callHTTPAPI("GET", "/api/auction/can-recover-lot", url.Values{
		"lotIndex": {fmt.Sprintf("%d", lotIndex)},
	})
	if err != nil {
		return api.CanRecoverRPLFromLotResponse{}, fmt.Errorf("Could not get can recover unclaimed RPL from lot status: %w", err)
	}
	var response api.CanRecoverRPLFromLotResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanRecoverRPLFromLotResponse{}, fmt.Errorf("Could not decode can recover unclaimed RPL from lot response: %w", err)
	}
	if response.Error != "" {
		return api.CanRecoverRPLFromLotResponse{}, fmt.Errorf("Could not get can recover unclaimed RPL from lot status: %s", response.Error)
	}
	return response, nil
}

// Recover unclaimed RPL from a lot (returning it to the auction contract)
func (c *Client) RecoverUnclaimedRPLFromLot(lotIndex uint64) (api.RecoverRPLFromLotResponse, error) {
	responseBytes, err := c.callHTTPAPI("POST", "/api/auction/recover-lot", url.Values{
		"lotIndex": {fmt.Sprintf("%d", lotIndex)},
	})
	if err != nil {
		return api.RecoverRPLFromLotResponse{}, fmt.Errorf("Could not recover unclaimed RPL from lot: %w", err)
	}
	var response api.RecoverRPLFromLotResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.RecoverRPLFromLotResponse{}, fmt.Errorf("Could not decode recover unclaimed RPL from lot response: %w", err)
	}
	if response.Error != "" {
		return api.RecoverRPLFromLotResponse{}, fmt.Errorf("Could not recover unclaimed RPL from lot: %s", response.Error)
	}
	return response, nil
}
