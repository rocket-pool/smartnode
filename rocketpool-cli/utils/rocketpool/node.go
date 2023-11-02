package rocketpool

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

type NodeRequester struct {
	client *http.Client
}

func NewNodeRequester(client *http.Client) *NodeRequester {
	return &NodeRequester{
		client: client,
	}
}

func (r *NodeRequester) GetName() string {
	return "Auction"
}
func (r *NodeRequester) GetRoute() string {
	return "auction"
}
func (r *NodeRequester) GetClient() *http.Client {
	return r.client
}

// Make a node deposit
func (r *NodeRequester) Deposit(amount *big.Int, minFee float64, salt *big.Int) (*api.ApiResponse[api.NodeDepositData], error) {
	args := map[string]string{
		"amount":       amount.String(),
		"min-node-fee": fmt.Sprint(minFee),
		"salt":         salt.String(),
	}
	return sendGetRequest[api.NodeDepositData](r, "deposit", "Deposit", args)
}

// Register the node
func (r *NodeRequester) Register(timezoneLocation string) (*api.ApiResponse[api.NodeRegisterData], error) {
	args := map[string]string{
		"timezone": timezoneLocation,
	}
	return sendGetRequest[api.NodeRegisterData](r, "register", "Register", args)
}

// Confirm the node's withdrawal address
func (r *NodeRequester) ConfirmPrimaryWithdrawalAddress() (*api.ApiResponse[api.NodeConfirmPrimaryWithdrawalAddressData], error) {
	return sendGetRequest[api.NodeConfirmPrimaryWithdrawalAddressData](r, "primary-withdrawal-address/confirm", "ConfirmPrimaryWithdrawalAddress", nil)
}

// Set the node's primary withdrawal address
func (r *NodeRequester) SetPrimaryWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (*api.ApiResponse[api.NodeSetPrimaryWithdrawalAddressData], error) {
	args := map[string]string{
		"address": withdrawalAddress.Hex(),
		"confirm": fmt.Sprint(confirm),
	}
	return sendGetRequest[api.NodeSetPrimaryWithdrawalAddressData](r, "primary-withdrawal-address/set", "SetPrimaryWithdrawalAddress", args)
}

// Confirm the node's RPL address
func (r *NodeRequester) ConfirmRplWithdrawalAddress() (*api.ApiResponse[api.NodeConfirmRplWithdrawalAddressData], error) {
	return sendGetRequest[api.NodeConfirmRplWithdrawalAddressData](r, "rpl-withdrawal-address/confirm", "ConfirmRplWithdrawalAddress", nil)
}

// Send tokens from the node to an address
func (r *NodeRequester) Send(amount *big.Int, token string, recipient common.Address) (*api.ApiResponse[api.NodeSendData], error) {
	args := map[string]string{
		"amount":    amount.String(),
		"token":     token,
		"recipient": recipient.Hex(),
	}
	return sendGetRequest[api.NodeSendData](r, "send", "Send", args)
}

// Set the node's RPL withdrawal address
func (r *NodeRequester) SetRplWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (*api.ApiResponse[api.NodeSetRplWithdrawalAddressData], error) {
	args := map[string]string{
		"address": withdrawalAddress.Hex(),
		"confirm": fmt.Sprint(confirm),
	}
	return sendGetRequest[api.NodeSetRplWithdrawalAddressData](r, "rpl-withdrawal-address/set", "SetRplWithdrawalAddress", args)
}

// Sets the allow state of another address staking on behalf of the node
func (r *NodeRequester) SetStakeRplForAllowed(caller common.Address, allowed bool) (*api.ApiResponse[api.NodeSetStakeRplForAllowedData], error) {
	args := map[string]string{
		"caller":  caller.Hex(),
		"allowed": fmt.Sprint(allowed),
	}
	return sendGetRequest[api.NodeSetStakeRplForAllowedData](r, "set-stake-rpl-for-allowed", "SetStakeRplForAllowed", args)
}

// Set the node's timezone location
func (r *NodeRequester) SetTimezone(timezoneLocation string) (*api.ApiResponse[api.NodeSetTimezoneData], error) {
	args := map[string]string{
		"timezone": timezoneLocation,
	}
	return sendGetRequest[api.NodeSetTimezoneData](r, "set-timezone", "SetTimezone", args)
}

// Stake RPL against the node
func (r *NodeRequester) StakeRpl(amount *big.Int) (*api.ApiResponse[api.NodeStakeRplData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return sendGetRequest[api.NodeStakeRplData](r, "stake-rpl", "StakeRpl", args)
}

// Get node status
func (r *NodeRequester) Status() (*api.ApiResponse[api.NodeStatusData], error) {
	return sendGetRequest[api.NodeStatusData](r, "status", "Status", nil)
}

// Swap node's old RPL tokens for new RPL tokens
func (r *NodeRequester) SwapRpl(amount *big.Int) (*api.ApiResponse[api.NodeSwapRplData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return sendGetRequest[api.NodeSwapRplData](r, "swap-rpl", "SwapRpl", args)
}

// Withdraw RPL staked against the node
func (r *NodeRequester) WithdrawRpl(amount *big.Int) (*api.ApiResponse[api.NodeWithdrawRplData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return sendGetRequest[api.NodeWithdrawRplData](r, "withdraw-rpl", "WithdrawRpl", args)
}

// ================================

// Burn tokens owned by the node for ETH
func (c *Client) NodeBurn(amountWei *big.Int, token string) (api.NodeBurnResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node burn %s %s", amountWei.String(), token))
	if err != nil {
		return api.NodeBurnResponse{}, fmt.Errorf("Could not burn tokens owned by node: %w", err)
	}
	var response api.NodeBurnResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeBurnResponse{}, fmt.Errorf("Could not decode node burn response: %w", err)
	}
	if response.Error != "" {
		return api.NodeBurnResponse{}, fmt.Errorf("Could not burn tokens owned by node: %s", response.Error)
	}
	return response, nil
}

// Get node sync progress
func (c *Client) NodeSync() (api.NodeSyncProgressResponse, error) {
	responseBytes, err := c.callAPI("node sync")
	if err != nil {
		return api.NodeSyncProgressResponse{}, fmt.Errorf("Could not get node sync: %w", err)
	}
	var response api.NodeSyncProgressResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeSyncProgressResponse{}, fmt.Errorf("Could not decode node sync response: %w", err)
	}
	if response.Error != "" {
		return api.NodeSyncProgressResponse{}, fmt.Errorf("Could not get node sync: %s", response.Error)
	}
	return response, nil
}

// Check whether the node has RPL rewards available to claim
func (c *Client) CanNodeClaimRpl() (api.CanNodeClaimRplResponse, error) {
	responseBytes, err := c.callAPI("node can-claim-rpl-rewards")
	if err != nil {
		return api.CanNodeClaimRplResponse{}, fmt.Errorf("Could not get can node claim rpl rewards status: %w", err)
	}
	var response api.CanNodeClaimRplResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeClaimRplResponse{}, fmt.Errorf("Could not decode can node claim rpl rewards response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeClaimRplResponse{}, fmt.Errorf("Could not get can node claim rpl rewards status: %s", response.Error)
	}
	return response, nil
}

// Claim available RPL rewards
func (c *Client) NodeClaimRpl() (api.NodeClaimRplResponse, error) {
	responseBytes, err := c.callAPI("node claim-rpl-rewards")
	if err != nil {
		return api.NodeClaimRplResponse{}, fmt.Errorf("Could not claim rpl rewards: %w", err)
	}
	var response api.NodeClaimRplResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeClaimRplResponse{}, fmt.Errorf("Could not decode node claim rpl rewards response: %w", err)
	}
	if response.Error != "" {
		return api.NodeClaimRplResponse{}, fmt.Errorf("Could not claim rpl rewards: %s", response.Error)
	}
	return response, nil
}

// Get node RPL rewards status
func (c *Client) NodeRewards() (api.NodeRewardsResponse, error) {
	responseBytes, err := c.callAPI("node rewards")
	if err != nil {
		return api.NodeRewardsResponse{}, fmt.Errorf("Could not get node rewards: %w", err)
	}
	var response api.NodeRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeRewardsResponse{}, fmt.Errorf("Could not decode node rewards response: %w", err)
	}
	if response.Error != "" {
		return api.NodeRewardsResponse{}, fmt.Errorf("Could not get node rewards: %s", response.Error)
	}
	return response, nil
}

// Get the deposit contract info for Rocket Pool and the Beacon Client
func (c *Client) DepositContractInfo() (api.NetworkDepositContractInfoData, error) {
	responseBytes, err := c.callAPI("node deposit-contract-info")
	if err != nil {
		return api.NetworkDepositContractInfoData{}, fmt.Errorf("Could not get deposit contract info: %w", err)
	}
	var response api.NetworkDepositContractInfoData
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NetworkDepositContractInfoData{}, fmt.Errorf("Could not decode deposit contract info response: %w", err)
	}
	if response.Error != "" {
		return api.NetworkDepositContractInfoData{}, fmt.Errorf("Could not get deposit contract info: %s", response.Error)
	}
	return response, nil
}

// Estimate the gas required to set a voting snapshot delegate
func (c *Client) EstimateSetSnapshotDelegateGas(address common.Address) (api.EstimateSetSnapshotDelegateGasResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node estimate-set-snapshot-delegate-gas %s", address.Hex()))
	if err != nil {
		return api.EstimateSetSnapshotDelegateGasResponse{}, fmt.Errorf("Could not get estimate-set-snapshot-delegate-gas response: %w", err)
	}
	var response api.EstimateSetSnapshotDelegateGasResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.EstimateSetSnapshotDelegateGasResponse{}, fmt.Errorf("Could not decode estimate-set-snapshot-delegate-gas response: %w", err)
	}
	if response.Error != "" {
		return api.EstimateSetSnapshotDelegateGasResponse{}, fmt.Errorf("Could not get estimate-set-snapshot-delegate-gas response: %s", response.Error)
	}
	return response, nil
}

// Set a voting snapshot delegate for the node
func (c *Client) SetSnapshotDelegate(address common.Address) (api.SetSnapshotDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node set-snapshot-delegate %s", address.Hex()))
	if err != nil {
		return api.SetSnapshotDelegateResponse{}, fmt.Errorf("Could not get set-snapshot-delegate response: %w", err)
	}
	var response api.SetSnapshotDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetSnapshotDelegateResponse{}, fmt.Errorf("Could not decode set-snapshot-delegate response: %w", err)
	}
	if response.Error != "" {
		return api.SetSnapshotDelegateResponse{}, fmt.Errorf("Could not get set-snapshot-delegate response: %s", response.Error)
	}
	return response, nil
}

// Estimate the gas required to clear the node's voting snapshot delegate
func (c *Client) EstimateClearSnapshotDelegateGas() (api.EstimateClearSnapshotDelegateGasResponse, error) {
	responseBytes, err := c.callAPI("node estimate-clear-snapshot-delegate-gas")
	if err != nil {
		return api.EstimateClearSnapshotDelegateGasResponse{}, fmt.Errorf("Could not get estimate-clear-snapshot-delegate-gas response: %w", err)
	}
	var response api.EstimateClearSnapshotDelegateGasResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.EstimateClearSnapshotDelegateGasResponse{}, fmt.Errorf("Could not decode estimate-clear-snapshot-delegate-gas response: %w", err)
	}
	if response.Error != "" {
		return api.EstimateClearSnapshotDelegateGasResponse{}, fmt.Errorf("Could not get estimate-clear-snapshot-delegate-gas response: %s", response.Error)
	}
	return response, nil
}

// Clear the node's voting snapshot delegate
func (c *Client) ClearSnapshotDelegate() (api.ClearSnapshotDelegateResponse, error) {
	responseBytes, err := c.callAPI("node clear-snapshot-delegate")
	if err != nil {
		return api.ClearSnapshotDelegateResponse{}, fmt.Errorf("Could not get clear-snapshot-delegate response: %w", err)
	}
	var response api.ClearSnapshotDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ClearSnapshotDelegateResponse{}, fmt.Errorf("Could not decode clear-snapshot-delegate response: %w", err)
	}
	if response.Error != "" {
		return api.ClearSnapshotDelegateResponse{}, fmt.Errorf("Could not get clear-snapshot-delegate response: %s", response.Error)
	}
	return response, nil
}

// Get the initialization status of the fee distributor contract
func (c *Client) IsFeeDistributorInitialized() (api.NodeFeeDistributorStatusResponse, error) {
	responseBytes, err := c.callAPI("node is-fee-distributor-initialized")
	if err != nil {
		return api.NodeFeeDistributorStatusResponse{}, fmt.Errorf("Could not get fee distributor initialization status: %w", err)
	}
	var response api.NodeFeeDistributorStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeFeeDistributorStatusResponse{}, fmt.Errorf("Could not decode fee distributor initialization status response: %w", err)
	}
	if response.Error != "" {
		return api.NodeFeeDistributorStatusResponse{}, fmt.Errorf("Could not get fee distributor initialization status: %s", response.Error)
	}
	return response, nil
}

// Get the gas cost for initializing the fee distributor contract
func (c *Client) GetInitializeFeeDistributorGas() (api.NodeInitializeFeeDistributorGasResponse, error) {
	responseBytes, err := c.callAPI("node get-initialize-fee-distributor-gas")
	if err != nil {
		return api.NodeInitializeFeeDistributorGasResponse{}, fmt.Errorf("Could not get initialize fee distributor gas: %w", err)
	}
	var response api.NodeInitializeFeeDistributorGasResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeInitializeFeeDistributorGasResponse{}, fmt.Errorf("Could not decode initialize fee distributor gas response: %w", err)
	}
	if response.Error != "" {
		return api.NodeInitializeFeeDistributorGasResponse{}, fmt.Errorf("Could not get initialize fee distributor gas: %s", response.Error)
	}
	return response, nil
}

// Initialize the fee distributor contract
func (c *Client) InitializeFeeDistributor() (api.NodeInitializeFeeDistributorResponse, error) {
	responseBytes, err := c.callAPI("node initialize-fee-distributor")
	if err != nil {
		return api.NodeInitializeFeeDistributorResponse{}, fmt.Errorf("Could not initialize fee distributor: %w", err)
	}
	var response api.NodeInitializeFeeDistributorResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeInitializeFeeDistributorResponse{}, fmt.Errorf("Could not decode initialize fee distributor response: %w", err)
	}
	if response.Error != "" {
		return api.NodeInitializeFeeDistributorResponse{}, fmt.Errorf("Could not initialize fee distributor: %s", response.Error)
	}
	return response, nil
}

// Check if distributing ETH from the node's fee distributor is possible
func (c *Client) CanDistribute() (api.NodeCanDistributeResponse, error) {
	responseBytes, err := c.callAPI("node can-distribute")
	if err != nil {
		return api.NodeCanDistributeResponse{}, fmt.Errorf("Could not get can distribute: %w", err)
	}
	var response api.NodeCanDistributeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeCanDistributeResponse{}, fmt.Errorf("Could not decode can distribute response: %w", err)
	}
	if response.Error != "" {
		return api.NodeCanDistributeResponse{}, fmt.Errorf("Could not get can distribute: %s", response.Error)
	}
	return response, nil
}

// Distribute ETH from the node's fee distributor
func (c *Client) Distribute() (api.NodeDistributeResponse, error) {
	responseBytes, err := c.callAPI("node distribute")
	if err != nil {
		return api.NodeDistributeResponse{}, fmt.Errorf("Could not distribute ETH: %w", err)
	}
	var response api.NodeDistributeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeDistributeResponse{}, fmt.Errorf("Could not decode distribute response: %w", err)
	}
	if response.Error != "" {
		return api.NodeDistributeResponse{}, fmt.Errorf("Could not distribute ETH: %s", response.Error)
	}
	return response, nil
}

// Get info about your eligible rewards periods, including balances and Merkle proofs
func (c *Client) GetRewardsInfo() (api.NodeGetRewardsInfoResponse, error) {
	responseBytes, err := c.callAPI("node get-rewards-info")
	if err != nil {
		return api.NodeGetRewardsInfoResponse{}, fmt.Errorf("Could not get rewards info: %w", err)
	}
	var response api.NodeGetRewardsInfoResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeGetRewardsInfoResponse{}, fmt.Errorf("Could not decode get rewards info response: %w", err)
	}
	if response.Error != "" {
		return api.NodeGetRewardsInfoResponse{}, fmt.Errorf("Could not get rewards info: %s", response.Error)
	}
	return response, nil
}

// Check if the rewards for the given intervals can be claimed
func (c *Client) CanNodeClaimRewards(indices []uint64) (api.CanNodeClaimRewardsResponse, error) {
	indexStrings := []string{}
	for _, index := range indices {
		indexStrings = append(indexStrings, fmt.Sprint(index))
	}
	responseBytes, err := c.callAPI("node can-claim-rewards", strings.Join(indexStrings, ","))
	if err != nil {
		return api.CanNodeClaimRewardsResponse{}, fmt.Errorf("Could not check if can claim rewards: %w", err)
	}
	var response api.CanNodeClaimRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeClaimRewardsResponse{}, fmt.Errorf("Could not decode can claim rewards response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeClaimRewardsResponse{}, fmt.Errorf("Could not check if can claim rewards: %s", response.Error)
	}
	return response, nil
}

// Claim rewards for the given reward intervals
func (c *Client) NodeClaimRewards(indices []uint64) (api.NodeClaimRewardsResponse, error) {
	indexStrings := []string{}
	for _, index := range indices {
		indexStrings = append(indexStrings, fmt.Sprint(index))
	}
	responseBytes, err := c.callAPI("node claim-rewards", strings.Join(indexStrings, ","))
	if err != nil {
		return api.NodeClaimRewardsResponse{}, fmt.Errorf("Could not claim rewards: %w", err)
	}
	var response api.NodeClaimRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeClaimRewardsResponse{}, fmt.Errorf("Could not decode claim rewards response: %w", err)
	}
	if response.Error != "" {
		return api.NodeClaimRewardsResponse{}, fmt.Errorf("Could not claim rewards: %s", response.Error)
	}
	return response, nil
}

// Check if the rewards for the given intervals can be claimed, and RPL restaked automatically
func (c *Client) CanNodeClaimAndStakeRewards(indices []uint64, stakeAmountWei *big.Int) (api.CanNodeClaimAndStakeRewardsResponse, error) {
	indexStrings := []string{}
	for _, index := range indices {
		indexStrings = append(indexStrings, fmt.Sprint(index))
	}
	responseBytes, err := c.callAPI("node can-claim-and-stake-rewards", strings.Join(indexStrings, ","), stakeAmountWei.String())
	if err != nil {
		return api.CanNodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not check if can claim and stake rewards: %w", err)
	}
	var response api.CanNodeClaimAndStakeRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not decode can claim and stake rewards response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not check if can claim and stake rewards: %s", response.Error)
	}
	return response, nil
}

// Claim rewards for the given reward intervals and restake RPL automatically
func (c *Client) NodeClaimAndStakeRewards(indices []uint64, stakeAmountWei *big.Int) (api.NodeClaimAndStakeRewardsResponse, error) {
	indexStrings := []string{}
	for _, index := range indices {
		indexStrings = append(indexStrings, fmt.Sprint(index))
	}
	responseBytes, err := c.callAPI("node claim-and-stake-rewards", strings.Join(indexStrings, ","), stakeAmountWei.String())
	if err != nil {
		return api.NodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not claim and stake rewards: %w", err)
	}
	var response api.NodeClaimAndStakeRewardsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not decode claim and stake rewards response: %w", err)
	}
	if response.Error != "" {
		return api.NodeClaimAndStakeRewardsResponse{}, fmt.Errorf("Could not claim and stake rewards: %s", response.Error)
	}
	return response, nil
}

// Check whether or not the node is opted into the Smoothing Pool
func (c *Client) NodeGetSmoothingPoolRegistrationStatus() (api.GetSmoothingPoolRegistrationStatusResponse, error) {
	responseBytes, err := c.callAPI("node get-smoothing-pool-registration-status")
	if err != nil {
		return api.GetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not get smoothing pool registration status: %w", err)
	}
	var response api.GetSmoothingPoolRegistrationStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not decode smoothing pool registration status response: %w", err)
	}
	if response.Error != "" {
		return api.GetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not get smoothing pool registration status: %s", response.Error)
	}
	return response, nil
}

// Check if the node's Smoothing Pool status can be changed
func (c *Client) CanNodeSetSmoothingPoolStatus(status bool) (api.CanSetSmoothingPoolRegistrationStatusResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node can-set-smoothing-pool-status %t", status))
	if err != nil {
		return api.CanSetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not get can-set-smoothing-pool-status: %w", err)
	}
	var response api.CanSetSmoothingPoolRegistrationStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanSetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not decode can-set-smoothing-pool-status response: %w", err)
	}
	if response.Error != "" {
		return api.CanSetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not get can-set-smoothing-pool-status: %s", response.Error)
	}
	return response, nil
}

// Sets the node's Smoothing Pool opt-in status
func (c *Client) NodeSetSmoothingPoolStatus(status bool) (api.SetSmoothingPoolRegistrationStatusResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node set-smoothing-pool-status %t", status))
	if err != nil {
		return api.SetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not set smoothing pool status: %w", err)
	}
	var response api.SetSmoothingPoolRegistrationStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not decode set-smoothing-pool-status response: %w", err)
	}
	if response.Error != "" {
		return api.SetSmoothingPoolRegistrationStatusResponse{}, fmt.Errorf("Could not set smoothing pool status: %s", response.Error)
	}
	return response, nil
}

func (c *Client) ResolveEnsName(name string) (api.ResolveEnsNameResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node resolve-ens-name %s", name))
	if err != nil {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not resolve ENS name: %w", err)
	}
	var response api.ResolveEnsNameResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not decode resolve-ens-name: %w", err)
	}
	if response.Error != "" {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not resolve ENS name: %s", response.Error)
	}
	return response, nil
}
func (c *Client) ReverseResolveEnsName(name string) (api.ResolveEnsNameResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node reverse-resolve-ens-name %s", name))
	if err != nil {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not reverse resolve ENS name: %w", err)
	}
	var response api.ResolveEnsNameResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not decode reverse-resolve-ens-name: %w", err)
	}
	if response.Error != "" {
		return api.ResolveEnsNameResponse{}, fmt.Errorf("Could not reverse resolve ENS name: %s", response.Error)
	}
	return response, nil
}

// Use the node private key to sign an arbitrary message
func (c *Client) SignMessage(message string) (api.NodeSignResponse, error) {
	// Ignore sync status so we can sign messages even without ready clients
	c.ignoreSyncCheck = true
	responseBytes, err := c.callAPI("node sign-message", message)
	if err != nil {
		return api.NodeSignResponse{}, fmt.Errorf("Could not sign message: %w", err)
	}

	var response api.NodeSignResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeSignResponse{}, fmt.Errorf("Could not decode node sign response: %w", err)
	}
	if response.Error != "" {
		return api.NodeSignResponse{}, fmt.Errorf("Could not sign message: %s", response.Error)
	}
	return response, nil
}

// Check whether a vacant minipool can be created for solo staker migration
func (c *Client) CanCreateVacantMinipool(amountWei *big.Int, minFee float64, salt *big.Int, pubkey types.ValidatorPubkey) (api.CanCreateVacantMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node can-create-vacant-minipool %s %f %s %s", amountWei.String(), minFee, salt.String(), pubkey.Hex()))
	if err != nil {
		return api.CanCreateVacantMinipoolResponse{}, fmt.Errorf("Could not get can create vacant minipool status: %w", err)
	}
	var response api.CanCreateVacantMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanCreateVacantMinipoolResponse{}, fmt.Errorf("Could not decode can create vacant minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CanCreateVacantMinipoolResponse{}, fmt.Errorf("Could not get can create vacant minipool status: %s", response.Error)
	}
	return response, nil
}

// Create a vacant minipool, which can be used to migrate a solo staker
func (c *Client) CreateVacantMinipool(amountWei *big.Int, minFee float64, salt *big.Int, pubkey types.ValidatorPubkey) (api.NodeCreateVacantMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node create-vacant-minipool %s %f %s %s", amountWei.String(), minFee, salt.String(), pubkey.Hex()))
	if err != nil {
		return api.NodeCreateVacantMinipoolResponse{}, fmt.Errorf("Could not get create vacant minipool status: %w", err)
	}
	var response api.NodeCreateVacantMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeCreateVacantMinipoolResponse{}, fmt.Errorf("Could not decode create vacant minipool response: %w", err)
	}
	if response.Error != "" {
		return api.NodeCreateVacantMinipoolResponse{}, fmt.Errorf("Could not get create vacant minipool status: %s", response.Error)
	}
	return response, nil
}

// Get the node's collateral info, including pending bond reductions
func (c *Client) CheckCollateral() (api.NodeCheckCollateralResponse, error) {
	responseBytes, err := c.callAPI("node check-collateral")
	if err != nil {
		return api.NodeCheckCollateralResponse{}, fmt.Errorf("Could not get check-collateral status: %w", err)
	}
	var response api.NodeCheckCollateralResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeCheckCollateralResponse{}, fmt.Errorf("Could not decode check-collateral response: %w", err)
	}
	if response.Error != "" {
		return api.NodeCheckCollateralResponse{}, fmt.Errorf("Could not get check-collateral status: %s", response.Error)
	}
	return response, nil
}

// Get the ETH balance of the node address
func (c *Client) GetEthBalance() (api.NodeEthBalanceResponse, error) {
	responseBytes, err := c.callAPI("node get-eth-balance")
	if err != nil {
		return api.NodeEthBalanceResponse{}, fmt.Errorf("Could not get get-eth-balance status: %w", err)
	}
	var response api.NodeEthBalanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeEthBalanceResponse{}, fmt.Errorf("Could not decode get-eth-balance response: %w", err)
	}
	if response.Error != "" {
		return api.NodeEthBalanceResponse{}, fmt.Errorf("Could not get get-eth-balance status: %s", response.Error)
	}
	return response, nil
}

// Estimates the gas for sending a zero-value message with a payload
func (c *Client) CanSendMessage(address common.Address, message []byte) (api.CanNodeSendMessageResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node can-send-message %s %s", address.Hex(), hex.EncodeToString(message)))
	if err != nil {
		return api.CanNodeSendMessageResponse{}, fmt.Errorf("Could not get can-send-message response: %w", err)
	}
	var response api.CanNodeSendMessageResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanNodeSendMessageResponse{}, fmt.Errorf("Could not decode can-send-message response: %w", err)
	}
	if response.Error != "" {
		return api.CanNodeSendMessageResponse{}, fmt.Errorf("Could not get can-send-message response: %s", response.Error)
	}
	return response, nil
}

// Sends a zero-value message with a payload
func (c *Client) SendMessage(address common.Address, message []byte) (api.NodeSendMessageResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("node send-message %s %s", address.Hex(), hex.EncodeToString(message)))
	if err != nil {
		return api.NodeSendMessageResponse{}, fmt.Errorf("Could not get send-message response: %w", err)
	}
	var response api.NodeSendMessageResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.NodeSendMessageResponse{}, fmt.Errorf("Could not decode send-message response: %w", err)
	}
	if response.Error != "" {
		return api.NodeSendMessageResponse{}, fmt.Errorf("Could not get send-message response: %s", response.Error)
	}
	return response, nil
}
