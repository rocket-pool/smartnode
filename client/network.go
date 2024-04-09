package client

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type NetworkRequester struct {
	context *client.RequesterContext
}

func NewNetworkRequester(context *client.RequesterContext) *NetworkRequester {
	return &NetworkRequester{
		context: context,
	}
}

func (r *NetworkRequester) GetName() string {
	return "Network"
}
func (r *NetworkRequester) GetRoute() string {
	return "network"
}
func (r *NetworkRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Get information about active Protocol DAO proposals on Snapshot
func (r *NetworkRequester) GetActiveDaoProposals() (*types.ApiResponse[api.NetworkDaoProposalsData], error) {
	return client.SendGetRequest[api.NetworkDaoProposalsData](r, "dao-proposals", "GetActiveDaoProposals", nil)
}

// Get the deposit contract info for Rocket Pool and the Beacon Client
func (r *NetworkRequester) GetDepositContractInfo() (*types.ApiResponse[api.NetworkDepositContractInfoData], error) {
	return client.SendGetRequest[api.NetworkDepositContractInfoData](r, "deposit-contract-info", "GetDepositContractInfo", nil)
}

// Download a rewards info file from IPFS or Github for the given interval
func (r *NetworkRequester) DownloadRewardsFile(interval uint64) (*types.ApiResponse[types.SuccessData], error) {
	args := map[string]string{
		"interval": fmt.Sprint(interval),
	}
	return client.SendGetRequest[types.SuccessData](r, "download-rewards-file", "DownloadRewardsFile", args)
}

// Set a request marker for the watchtower to generate the rewards tree for the given interval
func (r *NetworkRequester) GenerateRewardsTree(index uint64) (*types.ApiResponse[types.SuccessData], error) {
	args := map[string]string{
		"index": fmt.Sprint(index),
	}
	return client.SendGetRequest[types.SuccessData](r, "generate-rewards-tree", "GenerateRewardsTree", args)
}

// Get the address of the latest minipool delegate contract
func (r *NetworkRequester) GetLatestDelegate() (*types.ApiResponse[api.NetworkLatestDelegateData], error) {
	return client.SendGetRequest[api.NetworkLatestDelegateData](r, "latest-delegate", "GetLatestDelegate", nil)
}

// Get network node fee
func (r *NetworkRequester) NodeFee() (*types.ApiResponse[api.NetworkNodeFeeData], error) {
	return client.SendGetRequest[api.NetworkNodeFeeData](r, "node-fee", "NodeFee", nil)
}

// Get information about whether or not a rewards file can be regenerated, and whether or not one already exists
func (r *NetworkRequester) RewardsFileInfo(index uint64) (*types.ApiResponse[api.NetworkRewardsFileData], error) {
	args := map[string]string{
		"index": fmt.Sprint(index),
	}
	return client.SendGetRequest[api.NetworkRewardsFileData](r, "rewards-file-info", "RewardsFileInfo", args)
}

// Get network RPL price
func (r *NetworkRequester) RplPrice() (*types.ApiResponse[api.NetworkRplPriceData], error) {
	return client.SendGetRequest[api.NetworkRplPriceData](r, "rpl-price", "RplPrice", nil)
}

// Get network stats
func (r *NetworkRequester) Stats() (*types.ApiResponse[api.NetworkStatsData], error) {
	return client.SendGetRequest[api.NetworkStatsData](r, "stats", "Stats", nil)
}

// Get the timezone map
func (r *NetworkRequester) TimezoneMap() (*types.ApiResponse[api.NetworkTimezonesData], error) {
	return client.SendGetRequest[api.NetworkTimezonesData](r, "timezone-map", "TimezoneMap", nil)
}
