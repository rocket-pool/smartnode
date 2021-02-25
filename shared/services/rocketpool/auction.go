package rocketpool

import (
    "encoding/json"
    "fmt"
    "math/big"

    "github.com/rocket-pool/smartnode/shared/types/api"
)


// Get RPL auction status
func (c *Client) AuctionStatus() (api.AuctionStatusResponse, error) {
    responseBytes, err := c.callAPI("auction status")
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
    return response, nil
}


// Get RPL lots for auction
func (c *Client) AuctionLots() (api.AuctionLotsResponse, error) {
    responseBytes, err := c.callAPI("auction lots")
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
    return response, nil
}


// Check whether the node can create a new lot
func (c *Client) CanCreateLot() (api.CanCreateLotResponse, error) {
    responseBytes, err := c.callAPI("auction can-create-lot")
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
    responseBytes, err := c.callAPI("auction create-lot")
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
func (c *Client) CanBidOnLot() (api.CanBidOnLotResponse, error) {
    responseBytes, err := c.callAPI("auction can-bid-lot")
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
func (c *Client) BidOnLot() (api.BidOnLotResponse, error) {
    responseBytes, err := c.callAPI("auction bid-lot")
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
func (c *Client) CanClaimFromLot() (api.CanClaimFromLotResponse, error) {
    responseBytes, err := c.callAPI("auction can-claim-lot")
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
func (c *Client) ClaimFromLot() (api.ClaimFromLotResponse, error) {
    responseBytes, err := c.callAPI("auction claim-lot")
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
func (c *Client) CanRecoverUnclaimedRPLFromLot() (api.CanRecoverRPLFromLotResponse, error) {
    responseBytes, err := c.callAPI("auction can-recover-lot")
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
func (c *Client) RecoverUnclaimedRPLFromLot() (api.RecoverRPLFromLotResponse, error) {
    responseBytes, err := c.callAPI("auction recover-lot")
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

