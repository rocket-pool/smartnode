package rocketpool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get network node fee
func (c *Client) NodeFee() (api.NodeFeeResponse, error) {
	responseBytes, err := c.callAPI("network node-fee")
	if err != nil {
		return api.NodeFeeResponse{}, fmt.Errorf("Could not get network node fee: %w", err)
	}
	var response api.NodeFeeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeFeeResponse{}, fmt.Errorf("Could not decode network node fee response: %w", err)
	}
	if response.Error != "" {
		return api.NodeFeeResponse{}, fmt.Errorf("Could not get network node fee: %s", response.Error)
	}
	return response, nil
}

// Get network RPL price
func (c *Client) RplPrice() (api.RplPriceResponse, error) {
	responseBytes, err := c.callAPI("network rpl-price")
	if err != nil {
		return api.RplPriceResponse{}, fmt.Errorf("Could not get network RPL price: %w", err)
	}
	var response api.RplPriceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.RplPriceResponse{}, fmt.Errorf("Could not decode network RPL price response: %w", err)
	}
	if response.Error != "" {
		return api.RplPriceResponse{}, fmt.Errorf("Could not get network RPL price: %s", response.Error)
	}
	if response.RplPrice == nil {
		response.RplPrice = big.NewInt(0)
	}
	if response.MinPer8EthMinipoolRplStake == nil {
		response.MinPer8EthMinipoolRplStake = big.NewInt(0)
	}
	if response.MinPer16EthMinipoolRplStake == nil {
		response.MinPer16EthMinipoolRplStake = big.NewInt(0)
	}
	return response, nil
}

// Get network stats
func (c *Client) NetworkStats() (api.NetworkStatsResponse, error) {
	responseBytes, err := c.callAPI("network stats")
	if err != nil {
		return api.NetworkStatsResponse{}, fmt.Errorf("Could not get network stats: %w", err)
	}
	var response api.NetworkStatsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkStatsResponse{}, fmt.Errorf("Could not decode network stats response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkStatsResponse{}, fmt.Errorf("Could not get network stats: %s", response.Error)
	}
	return response, nil
}

// Get the timezone map
func (c *Client) TimezoneMap() (api.NetworkTimezonesResponse, error) {
	responseBytes, err := c.callAPI("network timezone-map")
	if err != nil {
		return api.NetworkTimezonesResponse{}, fmt.Errorf("Could not get network timezone map: %w", err)
	}
	var response api.NetworkTimezonesResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkTimezonesResponse{}, fmt.Errorf("Could not decode network timezone map response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkTimezonesResponse{}, fmt.Errorf("Could not get network timezone map: %s", response.Error)
	}
	return response, nil
}

// Check if the rewards tree for the provided interval can be generated
func (c *Client) CanGenerateRewardsTree(index uint64) (api.CanNetworkGenerateRewardsTreeResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("network can-generate-rewards-tree %d", index))
	if err != nil {
		return api.CanNetworkGenerateRewardsTreeResponse{}, fmt.Errorf("Could not check rewards tree generation status: %w", err)
	}
	var response api.CanNetworkGenerateRewardsTreeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNetworkGenerateRewardsTreeResponse{}, fmt.Errorf("Could not decode rewards tree generation status response: %w", err)
	}
	if response.Error != "" {
		return api.CanNetworkGenerateRewardsTreeResponse{}, fmt.Errorf("Could not check rewards tree generation status: %s", response.Error)
	}
	return response, nil
}

// Set a request marker for the watchtower to generate the rewards tree for the given interval
func (c *Client) GenerateRewardsTree(index uint64) (api.NetworkGenerateRewardsTreeResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("network generate-rewards-tree %d", index))
	if err != nil {
		return api.NetworkGenerateRewardsTreeResponse{}, fmt.Errorf("Could not initialize rewards tree generation: %w", err)
	}
	var response api.NetworkGenerateRewardsTreeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkGenerateRewardsTreeResponse{}, fmt.Errorf("Could not decode rewards tree generation response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkGenerateRewardsTreeResponse{}, fmt.Errorf("Could not initialize rewards tree generation: %s", response.Error)
	}
	return response, nil
}

// GetActiveDAOProposals fetches information about active DAO proposals
func (c *Client) GetActiveDAOProposals() (api.NetworkDAOProposalsResponse, error) {
	responseBytes, err := c.callAPI("network dao-proposals")
	if err != nil {
		return api.NetworkDAOProposalsResponse{}, fmt.Errorf("could not request active DAO proposals: %w", err)
	}
	var response api.NetworkDAOProposalsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkDAOProposalsResponse{}, fmt.Errorf("could not decode dao proposals response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkDAOProposalsResponse{}, fmt.Errorf("error after requesting dao proposals: %s", response.Error)
	}
	return response, nil
}

// Download a rewards info file from IPFS for the given interval
func (c *Client) DownloadRewardsFile(interval uint64) (api.DownloadRewardsFileResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("network download-rewards-file %d", interval))
	if err != nil {
		return api.DownloadRewardsFileResponse{}, fmt.Errorf("could not download rewards file: %w", err)
	}
	var response api.DownloadRewardsFileResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.DownloadRewardsFileResponse{}, fmt.Errorf("could not decode download-rewards-file response: %w", err)
	}
	if response.Error != "" {
		return api.DownloadRewardsFileResponse{}, fmt.Errorf("error after downloading rewards file: %s", response.Error)
	}
	return response, nil
}

// Check if Houston has been deployed yet
func (c *Client) IsHoustonDeployed() (api.IsHoustonDeployedResponse, error) {
	responseBytes, err := c.callAPI("network is-houston-deployed")
	if err != nil {
		return api.IsHoustonDeployedResponse{}, fmt.Errorf("could not check if Houston is deployed: %w", err)
	}
	var response api.IsHoustonDeployedResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.IsHoustonDeployedResponse{}, fmt.Errorf("could not decode is-houston-deployed response: %w", err)
	}
	if response.Error != "" {
		return api.IsHoustonDeployedResponse{}, fmt.Errorf("could not check if Houston is deployed: %s", response.Error)
	}
	return response, nil
}

// Get the address of the latest minipool delegate contract
func (c *Client) GetLatestDelegate() (api.GetLatestDelegateResponse, error) {
	responseBytes, err := c.callAPI("network latest-delegate")
	if err != nil {
		return api.GetLatestDelegateResponse{}, fmt.Errorf("could not get latest delegate: %w", err)
	}
	var response api.GetLatestDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetLatestDelegateResponse{}, fmt.Errorf("could not decode get-latest-delegate response: %w", err)
	}
	if response.Error != "" {
		return api.GetLatestDelegateResponse{}, fmt.Errorf("could not get latest delegate: %s", response.Error)
	}
	return response, nil
}

// CanInitializeVoting fetches whether the node's is initialized for on-chain voting
func (c *Client) CanInitializeVoting() (api.NetworkCanInitializeVotingResponse, error) {
	responseBytes, err := c.callAPI("network can-initialize-voting")
	if err != nil {
		return api.NetworkCanInitializeVotingResponse{}, fmt.Errorf("could not call can-initialize-voting: %w", err)
	}
	var response api.NetworkCanInitializeVotingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkCanInitializeVotingResponse{}, fmt.Errorf("could not decode can-initialize-voting response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkCanInitializeVotingResponse{}, fmt.Errorf("error after requesting can-initialize-voting: %s", response.Error)
	}
	return response, nil
}

// InitializeVoting unlocks the node's voting power
func (c *Client) InitializeVoting() (api.NetworkInitializeVotingResponse, error) {
	responseBytes, err := c.callAPI("network initialize-voting")
	if err != nil {
		return api.NetworkInitializeVotingResponse{}, fmt.Errorf("could not call initialize-voting: %w", err)
	}
	var response api.NetworkInitializeVotingResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkInitializeVotingResponse{}, fmt.Errorf("could not decode initialize-voting response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkInitializeVotingResponse{}, fmt.Errorf("error after requesting initialize-voting: %s", response.Error)
	}
	return response, nil
}

// CanSetVotingDelegate estimates the gas required to set an on-chain voting delegate
func (c *Client) EstimateSetVotingDelegateGas(address common.Address) (api.NetworkCanSetVotingDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("network estimate-set-voting-delegate-gas %s", address.Hex()))
	if err != nil {
		return api.NetworkCanSetVotingDelegateResponse{}, fmt.Errorf("could not call estimate-set-voting-delegate-gas: %w", err)
	}
	var response api.NetworkCanSetVotingDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkCanSetVotingDelegateResponse{}, fmt.Errorf("could not decode estimate-set-voting-delegate-gas response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkCanSetVotingDelegateResponse{}, fmt.Errorf("error after requesting estimate-set-voting-delegate-gas: %s", response.Error)
	}
	return response, nil
}

// SetVotingDelegate set an on-chain voting delegate for the node
func (c *Client) SetVotingDelegate(address common.Address) (api.NetworkSetVotingDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("network set-voting-delegate %s", address.Hex()))
	if err != nil {
		return api.NetworkSetVotingDelegateResponse{}, fmt.Errorf("could not call set-voting-delegate: %w", err)
	}
	var response api.NetworkSetVotingDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkSetVotingDelegateResponse{}, fmt.Errorf("could not decode set-voting-delegate response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkSetVotingDelegateResponse{}, fmt.Errorf("error after requesting set-voting-delegate: %s", response.Error)
	}
	return response, nil
}

// GetActiveDAOProposals fetches information about active DAO proposals
func (c *Client) GetCurrentVotingDelegate() (api.NetworkCurrentVotingDelegateResponse, error) {
	responseBytes, err := c.callAPI("network get-current-voting-delegate")
	if err != nil {
		return api.NetworkCurrentVotingDelegateResponse{}, fmt.Errorf("could not request get-current-voting-delegate: %w", err)
	}
	var response api.NetworkCurrentVotingDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkCurrentVotingDelegateResponse{}, fmt.Errorf("could not decode get-current-voting-delegate: %w", err)
	}
	if response.Error != "" {
		return api.NetworkCurrentVotingDelegateResponse{}, fmt.Errorf("error after requesting get-current-voting-delegate: %s", response.Error)
	}
	return response, nil
}
