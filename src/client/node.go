package client

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/node-manager-core/beacon"

	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type NodeRequester struct {
	context *client.RequesterContext
}

func NewNodeRequester(context *client.RequesterContext) *NodeRequester {
	return &NodeRequester{
		context: context,
	}
}

func (r *NodeRequester) GetName() string {
	return "Node"
}
func (r *NodeRequester) GetRoute() string {
	return "node"
}
func (r *NodeRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Get the node's ETH balance
func (r *NodeRequester) Balance() (*types.ApiResponse[api.NodeBalanceData], error) {
	return client.SendGetRequest[api.NodeBalanceData](r, "balance", "Balance", nil)
}

// Burn rETH owned by the node for ETH
func (r *NodeRequester) Burn(amount *big.Int) (*types.ApiResponse[api.NodeBurnData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return client.SendGetRequest[api.NodeBurnData](r, "burn", "Burn", args)
}

// Get the node's collateral info, including pending bond reductions
func (r *NodeRequester) CheckCollateral() (*types.ApiResponse[api.NodeCheckCollateralData], error) {
	return client.SendGetRequest[api.NodeCheckCollateralData](r, "check-collateral", "CheckCollateral", nil)
}

// Claim rewards for the given reward intervals
func (r *NodeRequester) ClaimAndStake(indices []*big.Int, stakeAmount *big.Int) (*types.ApiResponse[types.TxInfoData], error) {
	indicesStrings := make([]string, len(indices))
	for i, index := range indices {
		indicesStrings[i] = index.String()
	}
	args := map[string]string{
		"indices":      strings.Join(indicesStrings, ","),
		"stake-amount": stakeAmount.String(),
	}
	return client.SendGetRequest[types.TxInfoData](r, "claim-and-stake", "ClaimAndStake", args)
}

// Create a vacant minipool, which can be used to migrate a solo staker
func (r *NodeRequester) CreateVacantMinipool(amount *big.Int, minFee float64, salt *big.Int, pubkey beacon.ValidatorPubkey) (*types.ApiResponse[api.NodeCreateVacantMinipoolData], error) {
	args := map[string]string{
		"amount":       amount.String(),
		"min-node-fee": fmt.Sprint(minFee),
		"salt":         salt.String(),
		"pubkey":       pubkey.Hex(),
	}
	return client.SendGetRequest[api.NodeCreateVacantMinipoolData](r, "create-vacant-minipool", "CreateVacantMinipool", args)
}

// Make a node deposit
func (r *NodeRequester) Deposit(amount *big.Int, minFee float64, salt *big.Int) (*types.ApiResponse[api.NodeDepositData], error) {
	args := map[string]string{
		"amount":       amount.String(),
		"min-node-fee": fmt.Sprint(minFee),
		"salt":         salt.String(),
	}
	return client.SendGetRequest[api.NodeDepositData](r, "deposit", "Deposit", args)
}

// Distribute ETH from the node's fee distributor
func (r *NodeRequester) Distribute() (*types.ApiResponse[api.NodeDistributeData], error) {
	return client.SendGetRequest[api.NodeDistributeData](r, "distribute", "Distribute", nil)
}

// Get info about your eligible rewards periods, including balances and Merkle proofs
func (r *NodeRequester) GetRewardsInfo() (*types.ApiResponse[api.NodeGetRewardsInfoData], error) {
	return client.SendGetRequest[api.NodeGetRewardsInfoData](r, "get-rewards-info", "GetRewardsInfo", nil)
}

// Initialize the fee distributor contract
func (r *NodeRequester) InitializeFeeDistributor() (*types.ApiResponse[api.NodeInitializeFeeDistributorData], error) {
	return client.SendGetRequest[api.NodeInitializeFeeDistributorData](r, "initialize-fee-distributor", "InitializeFeeDistributor", nil)
}

// Confirm the node's withdrawal address
func (r *NodeRequester) ConfirmPrimaryWithdrawalAddress() (*types.ApiResponse[api.NodeConfirmPrimaryWithdrawalAddressData], error) {
	return client.SendGetRequest[api.NodeConfirmPrimaryWithdrawalAddressData](r, "primary-withdrawal-address/confirm", "ConfirmPrimaryWithdrawalAddress", nil)
}

// Set the node's primary withdrawal address
func (r *NodeRequester) SetPrimaryWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (*types.ApiResponse[api.NodeSetPrimaryWithdrawalAddressData], error) {
	args := map[string]string{
		"address": withdrawalAddress.Hex(),
		"confirm": fmt.Sprint(confirm),
	}
	return client.SendGetRequest[api.NodeSetPrimaryWithdrawalAddressData](r, "primary-withdrawal-address/set", "SetPrimaryWithdrawalAddress", args)
}

// Register the node
func (r *NodeRequester) Register(timezoneLocation string) (*types.ApiResponse[api.NodeRegisterData], error) {
	args := map[string]string{
		"timezone": timezoneLocation,
	}
	return client.SendGetRequest[api.NodeRegisterData](r, "register", "Register", args)
}

// Resolves an ENS name or reserve resolves an address
func (r *NodeRequester) ResolveEns(address common.Address, name string) (*types.ApiResponse[api.NodeResolveEnsData], error) {
	args := map[string]string{
		"address": address.Hex(),
		"name":    name,
	}
	return client.SendGetRequest[api.NodeResolveEnsData](r, "resolve-ens", "ResolveEns", args)
}

// Get node rewards status
func (r *NodeRequester) Rewards() (*types.ApiResponse[api.NodeRewardsData], error) {
	return client.SendGetRequest[api.NodeRewardsData](r, "rewards", "Rewards", nil)
}

// Confirm the node's RPL address
func (r *NodeRequester) ConfirmRplWithdrawalAddress() (*types.ApiResponse[api.NodeConfirmRplWithdrawalAddressData], error) {
	return client.SendGetRequest[api.NodeConfirmRplWithdrawalAddressData](r, "rpl-withdrawal-address/confirm", "ConfirmRplWithdrawalAddress", nil)
}

// Set the node's RPL withdrawal address
func (r *NodeRequester) SetRplWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (*types.ApiResponse[api.NodeSetRplWithdrawalAddressData], error) {
	args := map[string]string{
		"address": withdrawalAddress.Hex(),
		"confirm": fmt.Sprint(confirm),
	}
	return client.SendGetRequest[api.NodeSetRplWithdrawalAddressData](r, "rpl-withdrawal-address/set", "SetRplWithdrawalAddress", args)
}

// Send tokens from the node to an address
func (r *NodeRequester) Send(amount *big.Int, token string, recipient common.Address) (*types.ApiResponse[api.NodeSendData], error) {
	args := map[string]string{
		"amount":    amount.String(),
		"token":     token,
		"recipient": recipient.Hex(),
	}
	return client.SendGetRequest[api.NodeSendData](r, "send", "Send", args)
}

// Sets whether or not the node is allowed to lock RPL for Protocol DAO proposal or challenge bonds
func (r *NodeRequester) SetRplLockingAllowed(allowed bool) (*types.ApiResponse[api.NodeSetRplLockingAllowedData], error) {
	args := map[string]string{
		"allowed": fmt.Sprint(allowed),
	}
	return client.SendGetRequest[api.NodeSetRplLockingAllowedData](r, "set-rpl-locking-allowed", "SetRplLockingAllowed", args)
}

// Sets the node's Smoothing Pool opt-in status
func (r *NodeRequester) SetSmoothingPoolRegistrationState(optIn bool) (*types.ApiResponse[api.NodeSetSmoothingPoolRegistrationStatusData], error) {
	args := map[string]string{
		"opt-in": fmt.Sprint(optIn),
	}
	return client.SendGetRequest[api.NodeSetSmoothingPoolRegistrationStatusData](r, "set-smoothing-pool-registration-state", "SetSmoothingPoolRegistrationState", args)
}

// Sets the allow state of another address staking on behalf of the node
func (r *NodeRequester) SetStakeRplForAllowed(caller common.Address, allowed bool) (*types.ApiResponse[api.NodeSetStakeRplForAllowedData], error) {
	args := map[string]string{
		"caller":  caller.Hex(),
		"allowed": fmt.Sprint(allowed),
	}
	return client.SendGetRequest[api.NodeSetStakeRplForAllowedData](r, "set-stake-rpl-for-allowed", "SetStakeRplForAllowed", args)
}

// Set the node's timezone location
func (r *NodeRequester) SetTimezone(timezoneLocation string) (*types.ApiResponse[types.TxInfoData], error) {
	args := map[string]string{
		"timezone": timezoneLocation,
	}
	return client.SendGetRequest[types.TxInfoData](r, "set-timezone", "SetTimezone", args)
}

// Clear the node's voting snapshot delegate
func (r *NodeRequester) ClearSnapshotDelegate() (*types.ApiResponse[types.TxInfoData], error) {
	return client.SendGetRequest[types.TxInfoData](r, "snapshot-delegate/clear", "ClearSnapshotDelegate", nil)
}

// Set a voting snapshot delegate for the node
func (r *NodeRequester) SetSnapshotDelegate(delegate common.Address) (*types.ApiResponse[types.TxInfoData], error) {
	args := map[string]string{
		"delegate": delegate.Hex(),
	}
	return client.SendGetRequest[types.TxInfoData](r, "snapshot-delegate/set", "SetSnapshotDelegate", args)
}

// Stake RPL against the node
func (r *NodeRequester) StakeRpl(amount *big.Int) (*types.ApiResponse[api.NodeStakeRplData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return client.SendGetRequest[api.NodeStakeRplData](r, "stake-rpl", "StakeRpl", args)
}

// Get node status
func (r *NodeRequester) Status() (*types.ApiResponse[api.NodeStatusData], error) {
	return client.SendGetRequest[api.NodeStatusData](r, "status", "Status", nil)
}

// Swap node's old RPL tokens for new RPL tokens
func (r *NodeRequester) SwapRpl(amount *big.Int) (*types.ApiResponse[api.NodeSwapRplData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return client.SendGetRequest[api.NodeSwapRplData](r, "swap-rpl", "SwapRpl", args)
}

// Withdraw ETH staked on behalf of the node
func (r *NodeRequester) WithdrawEth(amount *big.Int) (*types.ApiResponse[api.NodeWithdrawEthData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return client.SendGetRequest[api.NodeWithdrawEthData](r, "withdraw-eth", "WithdrawEth", args)
}

// Withdraw RPL staked against the node
func (r *NodeRequester) WithdrawRpl(amount *big.Int) (*types.ApiResponse[api.NodeWithdrawRplData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return client.SendGetRequest[api.NodeWithdrawRplData](r, "withdraw-rpl", "WithdrawRpl", args)
}
