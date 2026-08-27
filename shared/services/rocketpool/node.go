package rocketpool

import (
	"context"
	"encoding/hex"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

func zeroIfNil(in **big.Int) {
	if *in == nil {
		*in = big.NewInt(0)
	}
}

// Get node status
func (c *Client) NodeStatus() (api.NodeStatusResponse, error) {
	response, err := c.callAPI[api.NodeStatusResponse]("GET", "/api/node/status", nil, "Could not get node status")
	if err != nil {
		return response, err
	}
	zeroIfNil(&response.TotalRplStake)
	zeroIfNil(&response.RplStakeMegapool)
	zeroIfNil(&response.RplStakeLegacy)
	zeroIfNil(&response.RplStakeThreshold)
	zeroIfNil(&response.AccountBalances.ETH)
	zeroIfNil(&response.AccountBalances.RPL)
	zeroIfNil(&response.AccountBalances.RETH)
	zeroIfNil(&response.AccountBalances.FixedSupplyRPL)
	zeroIfNil(&response.PrimaryWithdrawalBalances.ETH)
	zeroIfNil(&response.PrimaryWithdrawalBalances.RPL)
	zeroIfNil(&response.PrimaryWithdrawalBalances.RETH)
	zeroIfNil(&response.PrimaryWithdrawalBalances.FixedSupplyRPL)
	zeroIfNil(&response.NodeRPLLocked)
	zeroIfNil(&response.RPLWithdrawalBalances.ETH)
	zeroIfNil(&response.RPLWithdrawalBalances.RPL)
	zeroIfNil(&response.RPLWithdrawalBalances.RETH)
	zeroIfNil(&response.RPLWithdrawalBalances.FixedSupplyRPL)
	zeroIfNil(&response.PendingMinimumRplStake)
	zeroIfNil(&response.PendingMaximumRplStake)
	zeroIfNil(&response.EthBorrowed)
	zeroIfNil(&response.EthBorrowedLimit)
	zeroIfNil(&response.PendingBorrowAmount)
	zeroIfNil(&response.CreditBalance)
	zeroIfNil(&response.FeeDistributorBalance)
	return response, nil
}

// Get active alerts from Alertmanager.
// Uses a short 2-second timeout: alerts are informational and displayed after
// every command, so they must never block the user if the daemon is not yet up.
func (c *Client) NodeAlerts() (api.NodeAlertsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return c.callAPICtx[api.NodeAlertsResponse](ctx, "GET", "/api/node/alerts", nil, "could not get node alerts")
}

// Check whether the node can be registered
func (c *Client) CanRegisterNode(timezoneLocation string) (api.CanRegisterNodeResponse, error) {
	return c.callAPI[api.CanRegisterNodeResponse]("GET", "/api/node/can-register", url.Values{"timezoneLocation": {timezoneLocation}}, "Could not get can register node status")
}

// Register the node
func (c *Client) RegisterNode(timezoneLocation string) (api.RegisterNodeResponse, error) {
	return c.callAPI[api.RegisterNodeResponse]("POST", "/api/node/register", url.Values{"timezoneLocation": {timezoneLocation}}, "Could not register node")
}

// Checks if the node's primary withdrawal address can be set
func (c *Client) CanSetNodePrimaryWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (api.CanSetNodePrimaryWithdrawalAddressResponse, error) {
	return c.callAPI[api.CanSetNodePrimaryWithdrawalAddressResponse]("GET", "/api/node/can-set-primary-withdrawal-address", url.Values{
		"address": {withdrawalAddress.Hex()},
		"confirm": {strconv.FormatBool(confirm)},
	}, "Could not get can set node primary withdrawal address")
}

// Set the node's primary withdrawal address
func (c *Client) SetNodePrimaryWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (api.SetNodePrimaryWithdrawalAddressResponse, error) {
	return c.callAPI[api.SetNodePrimaryWithdrawalAddressResponse]("POST", "/api/node/set-primary-withdrawal-address", url.Values{
		"address": {withdrawalAddress.Hex()},
		"confirm": {strconv.FormatBool(confirm)},
	}, "Could not set node primary withdrawal address")
}

// Checks if the node's primary withdrawal address can be confirmed
func (c *Client) CanConfirmNodePrimaryWithdrawalAddress() (api.CanSetNodePrimaryWithdrawalAddressResponse, error) {
	return c.callAPI[api.CanSetNodePrimaryWithdrawalAddressResponse]("GET", "/api/node/can-confirm-primary-withdrawal-address", nil, "Could not get can confirm node primary withdrawal address")
}

// Confirm the node's primary withdrawal address
func (c *Client) ConfirmNodePrimaryWithdrawalAddress() (api.SetNodePrimaryWithdrawalAddressResponse, error) {
	return c.callAPI[api.SetNodePrimaryWithdrawalAddressResponse]("POST", "/api/node/confirm-primary-withdrawal-address", nil, "Could not confirm node primary withdrawal address")
}

// Checks if the node's RPL withdrawal address can be set
func (c *Client) CanSetNodeRPLWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (api.CanSetNodeRPLWithdrawalAddressResponse, error) {
	return c.callAPI[api.CanSetNodeRPLWithdrawalAddressResponse]("GET", "/api/node/can-set-rpl-withdrawal-address", url.Values{
		"address": {withdrawalAddress.Hex()},
		"confirm": {strconv.FormatBool(confirm)},
	}, "Could not get can set node RPL withdrawal address")
}

// Set the node's RPL withdrawal address
func (c *Client) SetNodeRPLWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (api.SetNodeRPLWithdrawalAddressResponse, error) {
	return c.callAPI[api.SetNodeRPLWithdrawalAddressResponse]("POST", "/api/node/set-rpl-withdrawal-address", url.Values{
		"address": {withdrawalAddress.Hex()},
		"confirm": {strconv.FormatBool(confirm)},
	}, "Could not set node RPL withdrawal address")
}

// Checks if the node's RPL withdrawal address can be confirmed
func (c *Client) CanConfirmNodeRPLWithdrawalAddress() (api.CanSetNodeRPLWithdrawalAddressResponse, error) {
	return c.callAPI[api.CanSetNodeRPLWithdrawalAddressResponse]("GET", "/api/node/can-confirm-rpl-withdrawal-address", nil, "Could not get can confirm node RPL withdrawal address")
}

// Confirm the node's RPL withdrawal address
func (c *Client) ConfirmNodeRPLWithdrawalAddress() (api.SetNodeRPLWithdrawalAddressResponse, error) {
	return c.callAPI[api.SetNodeRPLWithdrawalAddressResponse]("POST", "/api/node/confirm-rpl-withdrawal-address", nil, "Could not confirm node RPL withdrawal address")
}

// Checks if the node's timezone location can be set
func (c *Client) CanSetNodeTimezone(timezoneLocation string) (api.CanSetNodeTimezoneResponse, error) {
	return c.callAPI[api.CanSetNodeTimezoneResponse]("GET", "/api/node/can-set-timezone", url.Values{"timezoneLocation": {timezoneLocation}}, "Could not get can set node timezone")
}

// Set the node's timezone location
func (c *Client) SetNodeTimezone(timezoneLocation string) (api.SetNodeTimezoneResponse, error) {
	return c.callAPI[api.SetNodeTimezoneResponse]("POST", "/api/node/set-timezone", url.Values{"timezoneLocation": {timezoneLocation}}, "Could not set node timezone")
}

// Check whether the node can swap RPL tokens
func (c *Client) CanNodeSwapRpl(amountWei *big.Int) (api.CanNodeSwapRplResponse, error) {
	return c.callAPI[api.CanNodeSwapRplResponse]("GET", "/api/node/can-swap-rpl", url.Values{"amountWei": {amountWei.String()}}, "Could not get can node swap RPL status")
}

// Get the gas estimate for approving legacy RPL interaction
func (c *Client) NodeSwapRplApprovalGas(amountWei *big.Int) (api.NodeSwapRplApproveGasResponse, error) {
	return c.callAPI[api.NodeSwapRplApproveGasResponse]("GET", "/api/node/get-swap-rpl-approval-gas", url.Values{"amountWei": {amountWei.String()}}, "Could not get old RPL approval gas")
}

// Approves old RPL for a token swap
func (c *Client) NodeSwapRplApprove(amountWei *big.Int) (api.NodeSwapRplApproveResponse, error) {
	return c.callAPI[api.NodeSwapRplApproveResponse]("POST", "/api/node/swap-rpl-approve-rpl", url.Values{"amountWei": {amountWei.String()}}, "Could not approve old RPL")
}

// Swap node's old RPL tokens for new RPL tokens, waiting for the approval to be included in a block first
func (c *Client) NodeWaitAndSwapRpl(amountWei *big.Int, approvalTxHash common.Hash) (api.NodeSwapRplSwapResponse, error) {
	return c.callAPI[api.NodeSwapRplSwapResponse]("POST", "/api/node/wait-and-swap-rpl", url.Values{
		"amountWei":      {amountWei.String()},
		"approvalTxHash": {approvalTxHash.Hex()},
	}, "Could not swap node's RPL tokens")
}

// Swap node's old RPL tokens for new RPL tokens
func (c *Client) NodeSwapRpl(amountWei *big.Int) (api.NodeSwapRplSwapResponse, error) {
	return c.callAPI[api.NodeSwapRplSwapResponse]("POST", "/api/node/swap-rpl", url.Values{"amountWei": {amountWei.String()}}, "Could not swap node's RPL tokens")
}

// Get a node's legacy RPL allowance for swapping on the new RPL contract
func (c *Client) GetNodeSwapRplAllowance() (api.NodeSwapRplAllowanceResponse, error) {
	return c.callAPI[api.NodeSwapRplAllowanceResponse]("GET", "/api/node/swap-rpl-allowance", nil, "Could not get node swap RPL allowance")
}

// Check whether the node can stake RPL
func (c *Client) CanNodeStakeRpl(amountWei *big.Int) (api.CanNodeStakeRplResponse, error) {
	return c.callAPI[api.CanNodeStakeRplResponse]("GET", "/api/node/can-stake-rpl", url.Values{"amountWei": {amountWei.String()}}, "Could not get can node stake RPL status")
}

// Get the gas estimate for approving new RPL interaction
func (c *Client) NodeStakeRplApprovalGas(amountWei *big.Int) (api.NodeStakeRplApproveGasResponse, error) {
	return c.callAPI[api.NodeStakeRplApproveGasResponse]("GET", "/api/node/get-stake-rpl-approval-gas", url.Values{"amountWei": {amountWei.String()}}, "Could not get new RPL approval gas")
}

// Approve RPL for staking against the node
func (c *Client) NodeStakeRplApprove(amountWei *big.Int) (api.NodeStakeRplApproveResponse, error) {
	return c.callAPI[api.NodeStakeRplApproveResponse]("POST", "/api/node/stake-rpl-approve-rpl", url.Values{"amountWei": {amountWei.String()}}, "Could not approve RPL for staking")
}

// Stake RPL against the node waiting for approvalTxHash to be included in a block first
func (c *Client) NodeWaitAndStakeRpl(amountWei *big.Int, approvalTxHash common.Hash) (api.NodeStakeRplStakeResponse, error) {
	return c.callAPI[api.NodeStakeRplStakeResponse]("POST", "/api/node/wait-and-stake-rpl", url.Values{
		"amountWei":      {amountWei.String()},
		"approvalTxHash": {approvalTxHash.Hex()},
	}, "Could not stake node RPL")
}

// Stake RPL against the node
func (c *Client) NodeStakeRpl(amountWei *big.Int) (api.NodeStakeRplStakeResponse, error) {
	return c.callAPI[api.NodeStakeRplStakeResponse]("POST", "/api/node/stake-rpl", url.Values{"amountWei": {amountWei.String()}}, "Could not stake node RPL")
}

// Get a node's RPL allowance for the staking contract
func (c *Client) GetNodeStakeRplAllowance() (api.NodeStakeRplAllowanceResponse, error) {
	return c.callAPI[api.NodeStakeRplAllowanceResponse]("GET", "/api/node/stake-rpl-allowance", nil, "Could not get node stake RPL allowance")
}

// Checks if the node operator can set RPL locking allowed
func (c *Client) CanSetRPLLockingAllowed(allowed bool) (api.CanSetRplLockingAllowedResponse, error) {
	return c.callAPI[api.CanSetRplLockingAllowedResponse]("GET", "/api/node/can-set-rpl-locking-allowed", url.Values{"allowed": {strconv.FormatBool(allowed)}}, "Could not get can set RPL locking allowed")
}

// Sets the allow state for the node to lock RPL
func (c *Client) SetRPLLockingAllowed(allowed bool) (api.SetRplLockingAllowedResponse, error) {
	return c.callAPI[api.SetRplLockingAllowedResponse]("POST", "/api/node/set-rpl-locking-allowed", url.Values{"allowed": {strconv.FormatBool(allowed)}}, "Could not set RPL locking allowed")
}

// Checks if the node operator can set RPL stake for allowed
func (c *Client) CanSetStakeRPLForAllowed(caller common.Address, allowed bool) (api.CanSetStakeRplForAllowedResponse, error) {
	return c.callAPI[api.CanSetStakeRplForAllowedResponse]("GET", "/api/node/can-set-stake-rpl-for-allowed", url.Values{
		"caller":  {caller.Hex()},
		"allowed": {strconv.FormatBool(allowed)},
	}, "Could not get can set stake RPL for allowed")
}

// Sets the allow state of another address staking on behalf of the node
func (c *Client) SetStakeRPLForAllowed(caller common.Address, allowed bool) (api.SetStakeRplForAllowedResponse, error) {
	return c.callAPI[api.SetStakeRplForAllowedResponse]("POST", "/api/node/set-stake-rpl-for-allowed", url.Values{
		"caller":  {caller.Hex()},
		"allowed": {strconv.FormatBool(allowed)},
	}, "Could not set stake RPL for allowed")
}

// Check whether the node can withdraw RPL
func (c *Client) CanNodeWithdrawRpl() (api.CanNodeWithdrawRplResponse, error) {
	return c.callAPI[api.CanNodeWithdrawRplResponse]("GET", "/api/node/can-withdraw-rpl", nil, "Could not get can node withdraw RPL status")
}

// Withdraw RPL staked against the node
func (c *Client) NodeWithdrawRpl() (api.NodeWithdrawRplResponse, error) {
	return c.callAPI[api.NodeWithdrawRplResponse]("POST", "/api/node/withdraw-rpl", nil, "Could not withdraw node RPL")
}

// Check whether the node can unstake legacy RPL
func (c *Client) CanNodeUnstakeLegacyRpl(amountWei *big.Int) (api.CanNodeUnstakeLegacyRplResponse, error) {
	return c.callAPI[api.CanNodeUnstakeLegacyRplResponse]("GET", "/api/node/can-unstake-legacy-rpl", url.Values{"amountWei": {amountWei.String()}}, "Could not get can node unstake legacy RPL status")
}

// Unstake legacy RPL staked against the node
func (c *Client) NodeUnstakeLegacyRpl(amountWei *big.Int) (api.NodeUnstakeLegacyRplResponse, error) {
	return c.callAPI[api.NodeUnstakeLegacyRplResponse]("POST", "/api/node/unstake-legacy-rpl", url.Values{"amountWei": {amountWei.String()}}, "Could not unstake node legacy RPL")
}

// Check whether the node can withdraw RPL
// Used if saturn is not deployed (v1.3.1)
func (c *Client) CanNodeWithdrawRplV1_3_1(amountWei *big.Int) (api.CanNodeWithdrawRplv1_3_1Response, error) {
	return c.callAPI[api.CanNodeWithdrawRplv1_3_1Response]("GET", "/api/node/can-withdraw-rpl-v131", url.Values{"amountWei": {amountWei.String()}}, "Could not get can node withdraw RPL status")
}

// Withdraw RPL staked against the node
// Used if saturn is not deployed (v1.3.1)
func (c *Client) NodeWithdrawRplV1_3_1(amountWei *big.Int) (api.NodeWithdrawRplResponse, error) {
	return c.callAPI[api.NodeWithdrawRplResponse]("POST", "/api/node/withdraw-rpl-v131", url.Values{"amountWei": {amountWei.String()}}, "Could not withdraw node RPL")
}

// Check whether the node can unstake RPL
func (c *Client) CanNodeUnstakeRpl(amountWei *big.Int) (api.CanNodeUnstakeRplResponse, error) {
	return c.callAPI[api.CanNodeUnstakeRplResponse]("GET", "/api/node/can-unstake-rpl", url.Values{"amountWei": {amountWei.String()}}, "Could not get can node unstake RPL status")
}

// Unstake RPL staked against the node
func (c *Client) NodeUnstakeRpl(amountWei *big.Int) (api.NodeUnstakeRplResponse, error) {
	return c.callAPI[api.NodeUnstakeRplResponse]("POST", "/api/node/unstake-rpl", url.Values{"amountWei": {amountWei.String()}}, "Could not unstake node RPL")
}

// Check whether we can withdraw ETH staked on behalf of the node
func (c *Client) CanNodeWithdrawEth(amountWei *big.Int) (api.CanNodeWithdrawEthResponse, error) {
	return c.callAPI[api.CanNodeWithdrawEthResponse]("GET", "/api/node/can-withdraw-eth", url.Values{"amountWei": {amountWei.String()}}, "Could not get can node withdraw ETH status")
}

// Withdraw ETH staked on behalf of the node
func (c *Client) NodeWithdrawEth(amountWei *big.Int) (api.NodeWithdrawEthResponse, error) {
	return c.callAPI[api.NodeWithdrawEthResponse]("POST", "/api/node/withdraw-eth", url.Values{"amountWei": {amountWei.String()}}, "Could not withdraw node ETH")
}

// Check whether we can withdraw credit from the node
func (c *Client) CanNodeWithdrawCredit(amountWei *big.Int) (api.CanNodeWithdrawCreditResponse, error) {
	return c.callAPI[api.CanNodeWithdrawCreditResponse]("GET", "/api/node/can-withdraw-credit", url.Values{"amountWei": {amountWei.String()}}, "Could not get can node withdraw credit status")
}

// Withdraw credit from the node as rETH
func (c *Client) NodeWithdrawCredit(amountWei *big.Int) (api.NodeWithdrawCreditResponse, error) {
	return c.callAPI[api.NodeWithdrawCreditResponse]("POST", "/api/node/withdraw-credit", url.Values{"amountWei": {amountWei.String()}}, "Could not withdraw credit")
}

// Check whether the node can make multiple deposits
func (c *Client) CanNodeDeposits(count uint64, amountWei *big.Int, minFee float64, salt *big.Int, expressTickets uint64) (api.CanNodeDepositsResponse, error) {
	return c.callAPI[api.CanNodeDepositsResponse]("GET", "/api/node/can-deposit", url.Values{
		"count":          {strconv.FormatUint(count, 10)},
		"amountWei":      {amountWei.String()},
		"minFee":         {strconv.FormatFloat(minFee, 'f', -1, 64)},
		"salt":           {salt.String()},
		"expressTickets": {strconv.FormatUint(expressTickets, 10)},
	}, "Could not get can node deposits status")
}

// Make multiple node deposits
func (c *Client) NodeDeposits(count uint64, amountWei *big.Int, minFee float64, salt *big.Int, useCreditBalance bool, expressTickets uint64, submit bool) (api.NodeDepositsResponse, error) {
	return c.callAPI[api.NodeDepositsResponse]("POST", "/api/node/deposit", url.Values{
		"count":            {strconv.FormatUint(count, 10)},
		"amountWei":        {amountWei.String()},
		"minFee":           {strconv.FormatFloat(minFee, 'f', -1, 64)},
		"salt":             {salt.String()},
		"expressTickets":   {strconv.FormatUint(expressTickets, 10)},
		"useCreditBalance": {strconv.FormatBool(useCreditBalance)},
		"submit":           {strconv.FormatBool(submit)},
	}, "Could not make node deposits")
}

// Check whether the node can send tokens
func (c *Client) CanNodeSend(amountRaw float64, token string, toAddress common.Address) (api.CanNodeSendResponse, error) {
	return c.callAPI[api.CanNodeSendResponse]("GET", "/api/node/can-send", url.Values{
		"amountRaw": {strconv.FormatFloat(amountRaw, 'f', 10, 64)},
		"token":     {token},
		"to":        {toAddress.Hex()},
	}, "Could not get can node send status")
}

// Send tokens from the node to an address
func (c *Client) NodeSend(amountRaw float64, token string, toAddress common.Address) (api.NodeSendResponse, error) {
	return c.callAPI[api.NodeSendResponse]("POST", "/api/node/send", url.Values{
		"amountRaw": {strconv.FormatFloat(amountRaw, 'f', 10, 64)},
		"token":     {token},
		"to":        {toAddress.Hex()},
	}, "Could not send tokens from node")
}

// Send all tokens of the given type from the node to an address.
// Uses the exact on-chain *big.Int balance to avoid float64 rounding errors.
func (c *Client) NodeSendAll(token string, toAddress common.Address) (api.NodeSendResponse, error) {
	return c.callAPI[api.NodeSendResponse]("POST", "/api/node/send-all", url.Values{
		"token": {token},
		"to":    {toAddress.Hex()},
	}, "Could not send tokens from node")
}

// Check whether the node can burn tokens
func (c *Client) CanNodeBurn(amountWei *big.Int, token string) (api.CanNodeBurnResponse, error) {
	return c.callAPI[api.CanNodeBurnResponse]("GET", "/api/node/can-burn", url.Values{
		"amountWei": {amountWei.String()},
		"token":     {token},
	}, "Could not get can node burn status")
}

// Burn tokens owned by the node for ETH
func (c *Client) NodeBurn(amountWei *big.Int, token string) (api.NodeBurnResponse, error) {
	return c.callAPI[api.NodeBurnResponse]("POST", "/api/node/burn", url.Values{
		"amountWei": {amountWei.String()},
		"token":     {token},
	}, "Could not burn tokens owned by node")
}

// Get node sync progress
func (c *Client) NodeSync() (api.NodeSyncProgressResponse, error) {
	return c.callAPI[api.NodeSyncProgressResponse]("GET", "/api/node/sync", nil, "Could not get node sync")
}

// Check whether the node has RPL rewards available to claim
func (c *Client) CanNodeClaimRpl() (api.CanNodeClaimRplResponse, error) {
	return c.callAPI[api.CanNodeClaimRplResponse]("GET", "/api/node/can-claim-rpl-rewards", nil, "Could not get can node claim rpl rewards status")
}

// Claim available RPL rewards
func (c *Client) NodeClaimRpl() (api.NodeClaimRplResponse, error) {
	return c.callAPI[api.NodeClaimRplResponse]("POST", "/api/node/claim-rpl-rewards", nil, "Could not claim rpl rewards")
}

// Get node RPL rewards status
func (c *Client) NodeRewards() (api.NodeRewardsResponse, error) {
	return c.callAPI[api.NodeRewardsResponse]("GET", "/api/node/rewards", nil, "Could not get node rewards")
}

// Get the deposit contract info for Rocket Pool and the Beacon Client
func (c *Client) DepositContractInfo() (api.DepositContractInfoResponse, error) {
	return c.callAPI[api.DepositContractInfoResponse]("GET", "/api/node/deposit-contract-info", nil, "Could not get deposit contract info")
}

// Get the initialization status of the fee distributor contract
func (c *Client) IsFeeDistributorInitialized() (api.NodeIsFeeDistributorInitializedResponse, error) {
	return c.callAPI[api.NodeIsFeeDistributorInitializedResponse]("GET", "/api/node/is-fee-distributor-initialized", nil, "Could not get fee distributor initialization status")
}

// Get the gas cost for initializing the fee distributor contract
func (c *Client) GetInitializeFeeDistributorGas() (api.NodeInitializeFeeDistributorGasResponse, error) {
	return c.callAPI[api.NodeInitializeFeeDistributorGasResponse]("GET", "/api/node/get-initialize-fee-distributor-gas", nil, "Could not get initialize fee distributor gas")
}

// Initialize the fee distributor contract
func (c *Client) InitializeFeeDistributor() (api.NodeInitializeFeeDistributorResponse, error) {
	return c.callAPI[api.NodeInitializeFeeDistributorResponse]("POST", "/api/node/initialize-fee-distributor", nil, "Could not initialize fee distributor")
}

// Check if distributing ETH from the node's fee distributor is possible
func (c *Client) CanDistribute() (api.NodeCanDistributeResponse, error) {
	return c.callAPI[api.NodeCanDistributeResponse]("GET", "/api/node/can-distribute", nil, "Could not get can distribute")
}

// Distribute ETH from the node's fee distributor
func (c *Client) Distribute() (api.NodeDistributeResponse, error) {
	return c.callAPI[api.NodeDistributeResponse]("POST", "/api/node/distribute", nil, "Could not distribute ETH")
}

// Get info about your eligible rewards periods, including balances and Merkle proofs
func (c *Client) GetRewardsInfo() (api.NodeGetRewardsInfoResponse, error) {
	return c.callAPI[api.NodeGetRewardsInfoResponse]("GET", "/api/node/get-rewards-info", nil, "Could not get rewards info")
}

// Check if the rewards for the given intervals can be claimed
func (c *Client) CanNodeClaimRewards(indices []uint64) (api.CanNodeClaimRewardsResponse, error) {
	indexStrings := make([]string, len(indices))
	for i, idx := range indices {
		indexStrings[i] = strconv.FormatUint(idx, 10)
	}
	return c.callAPI[api.CanNodeClaimRewardsResponse]("GET", "/api/node/can-claim-rewards", url.Values{"indices": {strings.Join(indexStrings, ",")}}, "Could not check if can claim rewards")
}

// Claim rewards for the given reward intervals
func (c *Client) NodeClaimRewards(indices []uint64) (api.NodeClaimRewardsResponse, error) {
	indexStrings := make([]string, len(indices))
	for i, idx := range indices {
		indexStrings[i] = strconv.FormatUint(idx, 10)
	}
	return c.callAPI[api.NodeClaimRewardsResponse]("POST", "/api/node/claim-rewards", url.Values{"indices": {strings.Join(indexStrings, ",")}}, "Could not claim rewards")
}

// Check if the rewards for the given intervals can be claimed, and RPL restaked automatically
func (c *Client) CanNodeClaimAndStakeRewards(indices []uint64, stakeAmountWei *big.Int) (api.CanNodeClaimAndStakeRewardsResponse, error) {
	indexStrings := make([]string, len(indices))
	for i, idx := range indices {
		indexStrings[i] = strconv.FormatUint(idx, 10)
	}
	return c.callAPI[api.CanNodeClaimAndStakeRewardsResponse]("GET", "/api/node/can-claim-and-stake-rewards", url.Values{
		"indices":     {strings.Join(indexStrings, ",")},
		"stakeAmount": {stakeAmountWei.String()},
	}, "Could not check if can claim and stake rewards")
}

// Claim rewards for the given reward intervals and restake RPL automatically
func (c *Client) NodeClaimAndStakeRewards(indices []uint64, stakeAmountWei *big.Int) (api.NodeClaimAndStakeRewardsResponse, error) {
	indexStrings := make([]string, len(indices))
	for i, idx := range indices {
		indexStrings[i] = strconv.FormatUint(idx, 10)
	}
	return c.callAPI[api.NodeClaimAndStakeRewardsResponse]("POST", "/api/node/claim-and-stake-rewards", url.Values{
		"indices":     {strings.Join(indexStrings, ",")},
		"stakeAmount": {stakeAmountWei.String()},
	}, "Could not claim and stake rewards")
}

// Check whether or not the node is opted into the Smoothing Pool
func (c *Client) NodeGetSmoothingPoolRegistrationStatus() (api.GetSmoothingPoolRegistrationStatusResponse, error) {
	return c.callAPI[api.GetSmoothingPoolRegistrationStatusResponse]("GET", "/api/node/get-smoothing-pool-registration-status", nil, "Could not get smoothing pool registration status")
}

// Check if the node's Smoothing Pool status can be changed
func (c *Client) CanNodeSetSmoothingPoolStatus(status bool) (api.CanSetSmoothingPoolRegistrationStatusResponse, error) {
	return c.callAPI[api.CanSetSmoothingPoolRegistrationStatusResponse]("GET", "/api/node/can-set-smoothing-pool-status", url.Values{"status": {strconv.FormatBool(status)}}, "Could not get can-set-smoothing-pool-status")
}

// Sets the node's Smoothing Pool opt-in status
func (c *Client) NodeSetSmoothingPoolStatus(status bool) (api.SetSmoothingPoolRegistrationStatusResponse, error) {
	return c.callAPI[api.SetSmoothingPoolRegistrationStatusResponse]("POST", "/api/node/set-smoothing-pool-status", url.Values{"status": {strconv.FormatBool(status)}}, "Could not set smoothing pool status")
}

func (c *Client) ResolveEnsName(name string) (api.ResolveEnsNameResponse, error) {
	return c.callAPI[api.ResolveEnsNameResponse]("GET", "/api/node/resolve-ens-name", url.Values{"name": {name}}, "Could not resolve ENS name")
}

func (c *Client) ReverseResolveEnsName(name string) (api.ResolveEnsNameResponse, error) {
	return c.callAPI[api.ResolveEnsNameResponse]("GET", "/api/node/reverse-resolve-ens-name", url.Values{"address": {name}}, "Could not reverse resolve ENS name")
}

// Use the node private key to sign an arbitrary message
func (c *Client) SignMessage(message string) (api.NodeSignResponse, error) {
	return c.callAPI[api.NodeSignResponse]("POST", "/api/node/sign-message", url.Values{"message": {message}}, "Could not sign message")
}

// Get the node's collateral info, including pending bond reductions
func (c *Client) CheckCollateral() (api.CheckCollateralResponse, error) {
	return c.callAPI[api.CheckCollateralResponse]("GET", "/api/node/check-collateral", nil, "Could not get check-collateral status")
}

// Get the ETH balance of the node address
func (c *Client) GetEthBalance() (api.NodeEthBalanceResponse, error) {
	return c.callAPI[api.NodeEthBalanceResponse]("GET", "/api/node/get-eth-balance", nil, "Could not get get-eth-balance status")
}

// Estimates the gas for sending a zero-value message with a payload
func (c *Client) CanSendMessage(address common.Address, message []byte) (api.CanNodeSendMessageResponse, error) {
	return c.callAPI[api.CanNodeSendMessageResponse]("GET", "/api/node/can-send-message", url.Values{
		"address": {address.Hex()},
		"message": {hex.EncodeToString(message)},
	}, "Could not get can-send-message response")
}

// Sends a zero-value message with a payload
func (c *Client) SendMessage(address common.Address, message []byte) (api.NodeSendMessageResponse, error) {
	return c.callAPI[api.NodeSendMessageResponse]("POST", "/api/node/send-message", url.Values{
		"address": {address.Hex()},
		"message": {hex.EncodeToString(message)},
	}, "Could not get send-message response")
}

// Get the number of express tickets available for the node
func (c *Client) GetExpressTicketCount() (api.GetExpressTicketCountResponse, error) {
	return c.callAPI[api.GetExpressTicketCountResponse]("GET", "/api/node/get-express-ticket-count", nil, "Could not get express ticket count")
}

// Check if the node's express tickets have been provisioned
func (c *Client) GetExpressTicketsProvisioned() (api.GetExpressTicketsProvisionedResponse, error) {
	return c.callAPI[api.GetExpressTicketsProvisionedResponse]("GET", "/api/node/get-express-tickets-provisioned", nil, "Could not get express tickets provisioned")
}

func (c *Client) CanProvisionExpressTickets() (api.CanProvisionExpressTicketsResponse, error) {
	return c.callAPI[api.CanProvisionExpressTicketsResponse]("GET", "/api/node/can-provision-express-tickets", nil, "Could not get can-provision-express-tickets response")
}

func (c *Client) ProvisionExpressTickets() (api.ProvisionExpressTicketsResponse, error) {
	return c.callAPI[api.ProvisionExpressTicketsResponse]("POST", "/api/node/provision-express-tickets", nil, "Could not get provision-express-tickets response")
}

// Check whether the node can claim unclaimed rewards
func (c *Client) CanClaimUnclaimedRewards(nodeAddress common.Address) (api.CanClaimUnclaimedRewardsResponse, error) {
	return c.callAPI[api.CanClaimUnclaimedRewardsResponse]("GET", "/api/node/can-claim-unclaimed-rewards", url.Values{"nodeAddress": {nodeAddress.Hex()}}, "Could not get can-claim-unclaimed-rewards response")
}

// Send unclaimed rewards to a node operator's withdrawal address
func (c *Client) ClaimUnclaimedRewards(nodeAddress common.Address) (api.ClaimUnclaimedRewardsResponse, error) {
	return c.callAPI[api.ClaimUnclaimedRewardsResponse]("POST", "/api/node/claim-unclaimed-rewards", url.Values{"nodeAddress": {nodeAddress.Hex()}}, "Could not get claim-unclaimed-rewards response")
}

// Get the bond requirement for a number of validators
func (c *Client) GetBondRequirement(numValidators uint64) (api.GetBondRequirementResponse, error) {
	return c.callAPI[api.GetBondRequirementResponse]("GET", "/api/node/get-bond-requirement", url.Values{"numValidators": {strconv.FormatUint(numValidators, 10)}}, "Could not get get-bond-requirement response")
}
