package rocketpool

import (
	"fmt"
	"math/big"
	"net/url"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get network node fee
func (c *Client) NodeFee() (api.NodeFeeResponse, error) {
	return c.callAPI[api.NodeFeeResponse]("GET", "/api/network/node-fee", nil, "Could not get network node fee")
}

// Get network RPL price
func (c *Client) RplPrice() (api.RplPriceResponse, error) {
	response, err := c.callAPI[api.RplPriceResponse]("GET", "/api/network/rpl-price", nil, "Could not get network RPL price")
	if err != nil {
		return response, err
	}
	if response.RplPrice == nil {
		response.RplPrice = big.NewInt(0)
	}
	return response, nil
}

// Get network stats
func (c *Client) NetworkStats() (api.NetworkStatsResponse, error) {
	return c.callAPI[api.NetworkStatsResponse]("GET", "/api/network/stats", nil, "Could not get network stats")
}

// Get the timezone map
func (c *Client) TimezoneMap() (api.NetworkTimezonesResponse, error) {
	return c.callAPI[api.NetworkTimezonesResponse]("GET", "/api/network/timezone-map", nil, "Could not get network timezone map")
}

// Check if the rewards tree for the provided interval can be generated
func (c *Client) CanGenerateRewardsTree(index uint64) (api.CanNetworkGenerateRewardsTreeResponse, error) {
	return c.callAPI[api.CanNetworkGenerateRewardsTreeResponse]("GET", "/api/network/can-generate-rewards-tree", url.Values{"index": {fmt.Sprintf("%d", index)}}, "Could not check rewards tree generation status")
}

// Set a request marker for the watchtower to generate the rewards tree for the given interval
func (c *Client) GenerateRewardsTree(index uint64) (api.NetworkGenerateRewardsTreeResponse, error) {
	return c.callAPI[api.NetworkGenerateRewardsTreeResponse]("POST", "/api/network/generate-rewards-tree", url.Values{"index": {fmt.Sprintf("%d", index)}}, "Could not initialize rewards tree generation")
}

// GetActiveDAOProposals fetches information about active DAO proposals
func (c *Client) GetActiveDAOProposals() (api.NetworkDAOProposalsResponse, error) {
	return c.callAPI[api.NetworkDAOProposalsResponse]("GET", "/api/network/dao-proposals", nil, "could not request active DAO proposals")
}

// Download a rewards info file from IPFS for the given interval
func (c *Client) DownloadRewardsFile(interval uint64) (api.DownloadRewardsFileResponse, error) {
	return c.callAPI[api.DownloadRewardsFileResponse]("POST", "/api/network/download-rewards-file", url.Values{"interval": {fmt.Sprintf("%d", interval)}}, "could not download rewards file")
}

// Get the address of the latest minipool delegate contract
func (c *Client) GetLatestDelegate() (api.GetLatestDelegateResponse, error) {
	return c.callAPI[api.GetLatestDelegateResponse]("GET", "/api/network/latest-delegate", nil, "could not get latest delegate")
}
