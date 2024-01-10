package rocketpool

import (
	"fmt"
	"net/http"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type NetworkRequester struct {
	client *http.Client
}

func NewNetworkRequester(client *http.Client) *NetworkRequester {
	return &NetworkRequester{
		client: client,
	}
}

func (r *NetworkRequester) GetName() string {
	return "Network"
}
func (r *NetworkRequester) GetRoute() string {
	return "network"
}
func (r *NetworkRequester) GetClient() *http.Client {
	return r.client
}

// Get information about active Protocol DAO proposals on Snapshot
func (r *NetworkRequester) GetActiveDaoProposals() (*api.ApiResponse[api.NetworkDaoProposalsData], error) {
	return sendGetRequest[api.NetworkDaoProposalsData](r, "dao-proposals", "GetActiveDaoProposals", nil)
}

// Get the deposit contract info for Rocket Pool and the Beacon Client
func (r *NetworkRequester) GetDepositContractInfo() (*api.ApiResponse[api.NetworkDepositContractInfoData], error) {
	return sendGetRequest[api.NetworkDepositContractInfoData](r, "deposit-contract-info", "GetDepositContractInfo", nil)
}

// Download a rewards info file from IPFS or Github for the given interval
func (r *NetworkRequester) DownloadRewardsFile(interval uint64) (*api.ApiResponse[api.SuccessData], error) {
	args := map[string]string{
		"interval": fmt.Sprint(interval),
	}
	return sendGetRequest[api.SuccessData](r, "download-rewards-file", "DownloadRewardsFile", args)
}

// Set a request marker for the watchtower to generate the rewards tree for the given interval
func (r *NetworkRequester) GenerateRewardsTree(index uint64) (*api.ApiResponse[api.SuccessData], error) {
	args := map[string]string{
		"index": fmt.Sprint(index),
	}
	return sendGetRequest[api.SuccessData](r, "generate-rewards-tree", "GenerateRewardsTree", args)
}

// Get the address of the latest minipool delegate contract
func (r *NetworkRequester) GetLatestDelegate() (*api.ApiResponse[api.NetworkLatestDelegateData], error) {
	return sendGetRequest[api.NetworkLatestDelegateData](r, "latest-delegate", "GetLatestDelegate", nil)
}

// Get network node fee
func (r *NetworkRequester) NodeFee() (*api.ApiResponse[api.NetworkNodeFeeData], error) {
	return sendGetRequest[api.NetworkNodeFeeData](r, "node-fee", "NodeFee", nil)
}

// Get information about whether or not a rewards file can be regenerated, and whether or not one already exists
func (r *NetworkRequester) RewardsFileInfo(index uint64) (*api.ApiResponse[api.NetworkRewardsFileData], error) {
	args := map[string]string{
		"index": fmt.Sprint(index),
	}
	return sendGetRequest[api.NetworkRewardsFileData](r, "rewards-file-info", "RewardsFileInfo", args)
}

// Get network RPL price
func (r *NetworkRequester) RplPrice() (*api.ApiResponse[api.NetworkRplPriceData], error) {
	return sendGetRequest[api.NetworkRplPriceData](r, "rpl-price", "RplPrice", nil)
}

// Get network stats
func (r *NetworkRequester) Stats() (*api.ApiResponse[api.NetworkStatsData], error) {
	return sendGetRequest[api.NetworkStatsData](r, "stats", "Stats", nil)
}

// Get the timezone map
func (r *NetworkRequester) TimezoneMap() (*api.ApiResponse[api.NetworkTimezonesData], error) {
	return sendGetRequest[api.NetworkTimezonesData](r, "timezone-map", "TimezoneMap", nil)
}
