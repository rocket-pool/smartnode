package rocketpool

import (
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

type MinipoolRequester struct {
	client *http.Client
}

func NewMinipoolRequester(client *http.Client) *MinipoolRequester {
	return &MinipoolRequester{
		client: client,
	}
}

func (r *MinipoolRequester) GetName() string {
	return "Minipool"
}
func (r *MinipoolRequester) GetRoute() string {
	return "minipool"
}
func (r *MinipoolRequester) GetClient() *http.Client {
	return r.client
}

// Get begin reduce bond details
func (r *MinipoolRequester) GetBeginReduceBondDetails() (*api.ApiResponse[api.MinipoolBeginReduceBondDetailsData], error) {
	return sendGetRequest[api.MinipoolBeginReduceBondDetailsData](r, "begin-reduce-bond/details", "GetBeginReduceBondDetails", nil)
}

// Begin reduce bond on minipools
func (r *MinipoolRequester) BeginReduceBond(addresses []common.Address, newBondAmount *big.Int) (*api.ApiResponse[api.BatchTxInfoData], error) {
	args := map[string]string{
		"new-bond-amount": newBondAmount.String(),
	}
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "begin-reduce-bond", "BeginReduceBond", addresses, args)
}

// Verify that migrating a solo validator's withdrawal creds to a minipool address is possible
func (r *MinipoolRequester) CanChangeWithdrawalCredentials(address common.Address, mnemonic string) (*api.ApiResponse[api.MinipoolCanChangeWithdrawalCredentialsData], error) {
	args := map[string]string{
		"address":  address.Hex(),
		"mnemonic": mnemonic,
	}
	return sendGetRequest[api.MinipoolCanChangeWithdrawalCredentialsData](r, "change-withdrawal-creds/verify", "CanChangeWithdrawalCredentials", args)
}

// Migrate a solo validator's withdrawal creds to a minipool address
func (r *MinipoolRequester) ChangeWithdrawalCredentials(address common.Address, mnemonic string) (*api.ApiResponse[api.SuccessData], error) {
	args := map[string]string{
		"address":  address.Hex(),
		"mnemonic": mnemonic,
	}
	return sendGetRequest[api.SuccessData](r, "change-withdrawal-creds", "ChangeWithdrawalCredentials", args)
}

// Get close details
func (r *MinipoolRequester) GetCloseDetails() (*api.ApiResponse[api.MinipoolCloseDetailsData], error) {
	return sendGetRequest[api.MinipoolCloseDetailsData](r, "close/details", "GetCloseDetails", nil)
}

// Close minipools
func (r *MinipoolRequester) Close(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "close", "Close", addresses, nil)
}

// Get delegate details
func (r *MinipoolRequester) GetDelegateDetails() (*api.ApiResponse[api.MinipoolDelegateDetailsData], error) {
	return sendGetRequest[api.MinipoolDelegateDetailsData](r, "delegate/details", "GetDelegateDetails", nil)
}

// Rollback minipool delegates
func (r *MinipoolRequester) RollbackDelegates(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "delegate/rollback", "RollbackDelegates", addresses, nil)
}

// Set the use-latest-delegate setting for minipools
func (r *MinipoolRequester) SetUseLatestDelegates(addresses []common.Address, setting bool) (*api.ApiResponse[api.BatchTxInfoData], error) {
	args := map[string]string{
		"setting": fmt.Sprint(setting),
	}
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "delegate/set-use-latest", "SetUseLatestDelegates", addresses, args)
}

// Upgrade minipool delegates
func (r *MinipoolRequester) UpgradeDelegates(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "delegate/upgrade", "UpgradeDelegates", addresses, nil)
}

// Get distribute minipool balances details
func (r *MinipoolRequester) GetDistributeDetails() (*api.ApiResponse[api.MinipoolDistributeDetailsData], error) {
	return sendGetRequest[api.MinipoolDistributeDetailsData](r, "distribute/details", "GetDistributeDetails", nil)
}

// Distribute minipool balances
func (r *MinipoolRequester) Distribute(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "distribute", "Distribute", addresses, nil)
}

// Get dissolve details
func (r *MinipoolRequester) GetDissolveDetails() (*api.ApiResponse[api.MinipoolDissolveDetailsData], error) {
	return sendGetRequest[api.MinipoolDissolveDetailsData](r, "dissolve/details", "GetDissolveDetails", nil)
}

// Dissolve minipools
func (r *MinipoolRequester) Dissolve(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "dissolve", "Dissolve", addresses, nil)
}

// Get exit details
func (r *MinipoolRequester) GetExitDetails() (*api.ApiResponse[api.MinipoolExitDetailsData], error) {
	return sendGetRequest[api.MinipoolExitDetailsData](r, "exit/details", "GetExitDetails", nil)
}

// Exit minipools
func (r *MinipoolRequester) Exit(addresses []common.Address) (*api.ApiResponse[api.SuccessData], error) {
	return sendMultiMinipoolRequest[api.SuccessData](r, "exit", "Exit", addresses, nil)
}

// Import a validator private key for a vacant minipool
func (r *MinipoolRequester) ImportKey(address common.Address, mnemonic string) (*api.ApiResponse[api.SuccessData], error) {
	args := map[string]string{
		"address":  address.Hex(),
		"mnemonic": mnemonic,
	}
	return sendGetRequest[api.SuccessData](r, "import-key", "ImportKey", args)
}

// Get promote details
func (r *MinipoolRequester) GetPromoteDetails() (*api.ApiResponse[api.MinipoolPromoteDetailsData], error) {
	return sendGetRequest[api.MinipoolPromoteDetailsData](r, "promote/details", "GetPromoteDetails", nil)
}

// Promote minipools
func (r *MinipoolRequester) Promote(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "promote", "Promote", addresses, nil)
}

// Get reduce bond details
func (r *MinipoolRequester) GetReduceBondDetails() (*api.ApiResponse[api.MinipoolReduceBondDetailsData], error) {
	return sendGetRequest[api.MinipoolReduceBondDetailsData](r, "reduce-bond/details", "GetReduceBondDetails", nil)
}

// Reduce bond on minipools
func (r *MinipoolRequester) ReduceBond(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "reduce-bond", "ReduceBond", addresses, nil)
}

// Get refund details
func (r *MinipoolRequester) GetRefundDetails() (*api.ApiResponse[api.MinipoolRefundDetailsData], error) {
	return sendGetRequest[api.MinipoolRefundDetailsData](r, "refund/details", "GetRefundDetails", nil)
}

// Refund ETH from minipools
func (r *MinipoolRequester) Refund(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "refund", "Refund", addresses, nil)
}

// Get rescue dissolved details
func (r *MinipoolRequester) GetRescueDissolvedDetails() (*api.ApiResponse[api.MinipoolRescueDissolvedDetailsData], error) {
	return sendGetRequest[api.MinipoolRescueDissolvedDetailsData](r, "rescue-dissolved/details", "GetRescueDissolvedDetails", nil)
}

// Rescue dissolved minipools
func (r *MinipoolRequester) RescueDissolved(addresses []common.Address, depositAmounts []*big.Int) (*api.ApiResponse[api.BatchTxInfoData], error) {
	amounts := make([]string, len(depositAmounts))
	for i, amount := range depositAmounts {
		amounts[i] = amount.String()
	}
	args := map[string]string{
		"deposit-amounts": strings.Join(amounts, ","),
	}
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "rescue-dissolved", "RescueDissolved", addresses, args)
}

// Get stake details
func (r *MinipoolRequester) GetStakeDetails() (*api.ApiResponse[api.MinipoolStakeDetailsData], error) {
	return sendGetRequest[api.MinipoolStakeDetailsData](r, "stake/details", "GetStakeDetails", nil)
}

// Stake minipools
func (r *MinipoolRequester) Stake(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "stake", "Stake", addresses, nil)
}

// Get minipool status
func (r *MinipoolRequester) Status() (*api.ApiResponse[api.MinipoolStatusData], error) {
	return sendGetRequest[api.MinipoolStatusData](r, "status", "Status", nil)
}

// Get the artifacts necessary for vanity address searching
func (r *MinipoolRequester) GetVanityArtifacts(nodeAddressStr string) (*api.ApiResponse[api.MinipoolVanityArtifactsData], error) {
	args := map[string]string{
		"node-address": nodeAddressStr,
	}
	return sendGetRequest[api.MinipoolVanityArtifactsData](r, "vanity-artifacts", "GetVanityArtifacts", args)
}

// Submit a minipool request that takes in a list of addresses and returns whatever type is requested
func sendMultiMinipoolRequest[DataType any](r *MinipoolRequester, method string, requestName string, addresses []common.Address, args map[string]string) (*api.ApiResponse[DataType], error) {
	if args == nil {
		args = map[string]string{}
	}
	args["addresses"] = makeBatchArg(addresses)
	return sendGetRequest[DataType](r, method, requestName, args)
}
