package rocketpool

import (
	"fmt"
	"net/http"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type NetworkRequester struct {
	client *http.Client
	route  string
}

func NewNetworkRequester(client *http.Client) *NetworkRequester {
	return &NetworkRequester{
		client: client,
		route:  "network",
	}
}

// Get information about active Protocol DAO proposals on Snapshot
func (r *NetworkRequester) GetActiveDaoProposals() (*api.ApiResponse[api.NetworkDaoProposalsData], error) {
	method := "dao-proposals"
	args := map[string]string{}
	response, err := SendGetRequest[api.NetworkDaoProposalsData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Network GetActiveDaoProposals request: %w", err)
	}
	return response, nil
}

// Download a rewards info file from IPFS or Github for the given interval
func (r *NetworkRequester) DownloadRewardsFile(interval uint64) (*api.ApiResponse[api.SuccessData], error) {
	method := "download-rewards-file"
	args := map[string]string{
		"interval": fmt.Sprint(interval),
	}
	response, err := SendGetRequest[api.SuccessData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Network DownloadRewardsFile request: %w", err)
	}
	return response, nil
}

// Set a request marker for the watchtower to generate the rewards tree for the given interval
func (r *NetworkRequester) GenerateRewardsTree(index uint64) (*api.ApiResponse[api.SuccessData], error) {
	method := "generate-rewards-tree"
	args := map[string]string{
		"index": fmt.Sprint(index),
	}
	response, err := SendGetRequest[api.SuccessData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Network GenerateRewardsTree request: %w", err)
	}
	return response, nil
}

// Get the address of the latest minipool delegate contract
func (r *NetworkRequester) GetLatestDelegate() (*api.ApiResponse[api.NetworkLatestDelegateData], error) {
	method := "latest-delegate"
	args := map[string]string{}
	response, err := SendGetRequest[api.NetworkLatestDelegateData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Network GetLatestDelegate request: %w", err)
	}
	return response, nil
}

// Get network node fee
func (r *NetworkRequester) NodeFee() (*api.ApiResponse[api.NetworkNodeFeeData], error) {
	method := "node-fee"
	args := map[string]string{}
	response, err := SendGetRequest[api.NetworkNodeFeeData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Network NodeFee request: %w", err)
	}
	return response, nil
}

// Get network RPL price
func (r *NetworkRequester) RplPrice() (*api.ApiResponse[api.NetworkRplPriceData], error) {
	method := "rpl-price"
	args := map[string]string{}
	response, err := SendGetRequest[api.NetworkRplPriceData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Network RplPrice request: %w", err)
	}
	return response, nil
}

// Get network stats
func (r *NetworkRequester) Stats() (*api.ApiResponse[api.NetworkStatsData], error) {
	method := "stats"
	args := map[string]string{}
	response, err := SendGetRequest[api.NetworkStatsData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Network Stats request: %w", err)
	}
	return response, nil
}

// Get the timezone map
func (r *NetworkRequester) TimezoneMap() (*api.ApiResponse[api.NetworkTimezonesData], error) {
	method := "timezone-map"
	args := map[string]string{}
	response, err := SendGetRequest[api.NetworkTimezonesData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Network TimezoneMap request: %w", err)
	}
	return response, nil
}
