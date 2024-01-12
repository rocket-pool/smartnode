package rocketpool

import (
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"

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

// Get the node's ETH balance
func (r *NodeRequester) Balance() (*api.ApiResponse[api.NodeBalanceData], error) {
	return sendGetRequest[api.NodeBalanceData](r, "balance", "Balance", nil)
}

// Burn rETH owned by the node for ETH
func (r *NodeRequester) Burn(amount *big.Int) (*api.ApiResponse[api.NodeBurnData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return sendGetRequest[api.NodeBurnData](r, "burn", "Burn", args)
}

// Get the node's collateral info, including pending bond reductions
func (r *NodeRequester) CheckCollateral() (*api.ApiResponse[api.NodeCheckCollateralData], error) {
	return sendGetRequest[api.NodeCheckCollateralData](r, "check-collateral", "CheckCollateral", nil)
}

// Claim rewards for the given reward intervals
func (r *NodeRequester) ClaimAndStake(indices []*big.Int, stakeAmount *big.Int) (*api.ApiResponse[api.TxInfoData], error) {
	indicesStrings := make([]string, len(indices))
	for i, index := range indices {
		indicesStrings[i] = index.String()
	}
	args := map[string]string{
		"indices":      strings.Join(indicesStrings, ","),
		"stake-amount": stakeAmount.String(),
	}
	return sendGetRequest[api.TxInfoData](r, "claim-and-stake", "ClaimAndStake", args)
}

// Create a vacant minipool, which can be used to migrate a solo staker
func (r *NodeRequester) CreateVacantMinipool(amount *big.Int, minFee float64, salt *big.Int, pubkey types.ValidatorPubkey) (*api.ApiResponse[api.NodeCreateVacantMinipoolData], error) {
	args := map[string]string{
		"amount":       amount.String(),
		"min-node-fee": fmt.Sprint(minFee),
		"salt":         salt.String(),
		"pubkey":       pubkey.Hex(),
	}
	return sendGetRequest[api.NodeCreateVacantMinipoolData](r, "create-vacant-minipool", "CreateVacantMinipool", args)
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

// Distribute ETH from the node's fee distributor
func (r *NodeRequester) Distribute() (*api.ApiResponse[api.NodeDistributeData], error) {
	return sendGetRequest[api.NodeDistributeData](r, "distribute", "Distribute", nil)
}

// Get info about your eligible rewards periods, including balances and Merkle proofs
func (r *NodeRequester) GetRewardsInfo() (*api.ApiResponse[api.NodeGetRewardsInfoData], error) {
	return sendGetRequest[api.NodeGetRewardsInfoData](r, "get-rewards-info", "GetRewardsInfo", nil)
}

// Initialize the fee distributor contract
func (r *NodeRequester) InitializeFeeDistributor() (*api.ApiResponse[api.NodeInitializeFeeDistributorData], error) {
	return sendGetRequest[api.NodeInitializeFeeDistributorData](r, "initialize-fee-distributor", "InitializeFeeDistributor", nil)
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

// Register the node
func (r *NodeRequester) Register(timezoneLocation string) (*api.ApiResponse[api.NodeRegisterData], error) {
	args := map[string]string{
		"timezone": timezoneLocation,
	}
	return sendGetRequest[api.NodeRegisterData](r, "register", "Register", args)
}

// Resolves an ENS name or reserve resolves an address
func (r *NodeRequester) ResolveEns(address common.Address, name string) (*api.ApiResponse[api.NodeResolveEnsData], error) {
	args := map[string]string{
		"address": address.Hex(),
		"name":    name,
	}
	return sendGetRequest[api.NodeResolveEnsData](r, "resolve-ens", "ResolveEns", args)
}

// Get node rewards status
func (r *NodeRequester) Rewards() (*api.ApiResponse[api.NodeRewardsData], error) {
	return sendGetRequest[api.NodeRewardsData](r, "rewards", "Rewards", nil)
}

// Confirm the node's RPL address
func (r *NodeRequester) ConfirmRplWithdrawalAddress() (*api.ApiResponse[api.NodeConfirmRplWithdrawalAddressData], error) {
	return sendGetRequest[api.NodeConfirmRplWithdrawalAddressData](r, "rpl-withdrawal-address/confirm", "ConfirmRplWithdrawalAddress", nil)
}

// Set the node's RPL withdrawal address
func (r *NodeRequester) SetRplWithdrawalAddress(withdrawalAddress common.Address, confirm bool) (*api.ApiResponse[api.NodeSetRplWithdrawalAddressData], error) {
	args := map[string]string{
		"address": withdrawalAddress.Hex(),
		"confirm": fmt.Sprint(confirm),
	}
	return sendGetRequest[api.NodeSetRplWithdrawalAddressData](r, "rpl-withdrawal-address/set", "SetRplWithdrawalAddress", args)
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

// Sets whether or not the node is allowed to lock RPL for Protocol DAO proposal or challenge bonds
func (r *NodeRequester) SetRplLockingAllowed(allowed bool) (*api.ApiResponse[api.NodeSetRplLockingAllowedData], error) {
	args := map[string]string{
		"allowed": fmt.Sprint(allowed),
	}
	return sendGetRequest[api.NodeSetRplLockingAllowedData](r, "set-rpl-locking-allowed", "SetRplLockingAllowed", args)
}

// Sets the node's Smoothing Pool opt-in status
func (r *NodeRequester) SetSmoothingPoolRegistrationState(optIn bool) (*api.ApiResponse[api.NodeSetSmoothingPoolRegistrationStatusData], error) {
	args := map[string]string{
		"opt-in": fmt.Sprint(optIn),
	}
	return sendGetRequest[api.NodeSetSmoothingPoolRegistrationStatusData](r, "set-smoothing-pool-registration-state", "SetSmoothingPoolRegistrationState", args)
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
func (r *NodeRequester) SetTimezone(timezoneLocation string) (*api.ApiResponse[api.TxInfoData], error) {
	args := map[string]string{
		"timezone": timezoneLocation,
	}
	return sendGetRequest[api.TxInfoData](r, "set-timezone", "SetTimezone", args)
}

// Clear the node's voting snapshot delegate
func (r *NodeRequester) ClearSnapshotDelegate() (*api.ApiResponse[api.TxInfoData], error) {
	return sendGetRequest[api.TxInfoData](r, "snapshot-delegate/clear", "ClearSnapshotDelegate", nil)
}

// Set a voting snapshot delegate for the node
func (r *NodeRequester) SetSnapshotDelegate(delegate common.Address) (*api.ApiResponse[api.TxInfoData], error) {
	args := map[string]string{
		"delegate": delegate.Hex(),
	}
	return sendGetRequest[api.TxInfoData](r, "snapshot-delegate/set", "SetSnapshotDelegate", args)
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

// Withdraw ETH staked on behalf of the node
func (r *NodeRequester) WithdrawEth(amount *big.Int) (*api.ApiResponse[api.NodeWithdrawEthData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return sendGetRequest[api.NodeWithdrawEthData](r, "withdraw-eth", "WithdrawEth", args)
}

// Withdraw RPL staked against the node
func (r *NodeRequester) WithdrawRpl(amount *big.Int) (*api.ApiResponse[api.NodeWithdrawRplData], error) {
	args := map[string]string{
		"amount": amount.String(),
	}
	return sendGetRequest[api.NodeWithdrawRplData](r, "withdraw-rpl", "WithdrawRpl", args)
}
