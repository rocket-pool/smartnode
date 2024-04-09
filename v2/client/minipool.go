package client

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/node-manager-core/api/client"
	"github.com/rocket-pool/node-manager-core/api/types"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

type MinipoolRequester struct {
	context *client.RequesterContext
}

func NewMinipoolRequester(context *client.RequesterContext) *MinipoolRequester {
	return &MinipoolRequester{
		context: context,
	}
}

func (r *MinipoolRequester) GetName() string {
	return "Minipool"
}
func (r *MinipoolRequester) GetRoute() string {
	return "minipool"
}
func (r *MinipoolRequester) GetContext() *client.RequesterContext {
	return r.context
}

// Get begin reduce bond details
func (r *MinipoolRequester) GetBeginReduceBondDetails() (*types.ApiResponse[api.MinipoolBeginReduceBondDetailsData], error) {
	return client.SendGetRequest[api.MinipoolBeginReduceBondDetailsData](r, "begin-reduce-bond/details", "GetBeginReduceBondDetails", nil)
}

// Begin reduce bond on minipools
func (r *MinipoolRequester) BeginReduceBond(addresses []common.Address, newBondAmount *big.Int) (*types.ApiResponse[types.BatchTxInfoData], error) {
	args := map[string]string{
		"new-bond-amount": newBondAmount.String(),
	}
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "begin-reduce-bond", "BeginReduceBond", addresses, args)
}

// Verify that migrating a solo validator's withdrawal creds to a minipool address is possible
func (r *MinipoolRequester) CanChangeWithdrawalCredentials(address common.Address, mnemonic string) (*types.ApiResponse[api.MinipoolCanChangeWithdrawalCredentialsData], error) {
	args := map[string]string{
		"address":  address.Hex(),
		"mnemonic": mnemonic,
	}
	return client.SendGetRequest[api.MinipoolCanChangeWithdrawalCredentialsData](r, "change-withdrawal-creds/verify", "CanChangeWithdrawalCredentials", args)
}

// Migrate a solo validator's withdrawal creds to a minipool address
func (r *MinipoolRequester) ChangeWithdrawalCredentials(address common.Address, mnemonic string) (*types.ApiResponse[types.SuccessData], error) {
	args := map[string]string{
		"address":  address.Hex(),
		"mnemonic": mnemonic,
	}
	return client.SendGetRequest[types.SuccessData](r, "change-withdrawal-creds", "ChangeWithdrawalCredentials", args)
}

// Get close details
func (r *MinipoolRequester) GetCloseDetails() (*types.ApiResponse[api.MinipoolCloseDetailsData], error) {
	return client.SendGetRequest[api.MinipoolCloseDetailsData](r, "close/details", "GetCloseDetails", nil)
}

// Close minipools
func (r *MinipoolRequester) Close(addresses []common.Address) (*types.ApiResponse[types.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "close", "Close", addresses, nil)
}

// Get delegate details
func (r *MinipoolRequester) GetDelegateDetails() (*types.ApiResponse[api.MinipoolDelegateDetailsData], error) {
	return client.SendGetRequest[api.MinipoolDelegateDetailsData](r, "delegate/details", "GetDelegateDetails", nil)
}

// Rollback minipool delegates
func (r *MinipoolRequester) RollbackDelegates(addresses []common.Address) (*types.ApiResponse[types.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "delegate/rollback", "RollbackDelegates", addresses, nil)
}

// Set the use-latest-delegate setting for minipools
func (r *MinipoolRequester) SetUseLatestDelegates(addresses []common.Address, setting bool) (*types.ApiResponse[types.BatchTxInfoData], error) {
	args := map[string]string{
		"setting": fmt.Sprint(setting),
	}
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "delegate/set-use-latest", "SetUseLatestDelegates", addresses, args)
}

// Upgrade minipool delegates
func (r *MinipoolRequester) UpgradeDelegates(addresses []common.Address) (*types.ApiResponse[types.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "delegate/upgrade", "UpgradeDelegates", addresses, nil)
}

// Get distribute minipool balances details
func (r *MinipoolRequester) GetDistributeDetails() (*types.ApiResponse[api.MinipoolDistributeDetailsData], error) {
	return client.SendGetRequest[api.MinipoolDistributeDetailsData](r, "distribute/details", "GetDistributeDetails", nil)
}

// Distribute minipool balances
func (r *MinipoolRequester) Distribute(addresses []common.Address) (*types.ApiResponse[types.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "distribute", "Distribute", addresses, nil)
}

// Get dissolve details
func (r *MinipoolRequester) GetDissolveDetails() (*types.ApiResponse[api.MinipoolDissolveDetailsData], error) {
	return client.SendGetRequest[api.MinipoolDissolveDetailsData](r, "dissolve/details", "GetDissolveDetails", nil)
}

// Dissolve minipools
func (r *MinipoolRequester) Dissolve(addresses []common.Address) (*types.ApiResponse[types.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "dissolve", "Dissolve", addresses, nil)
}

// Get exit details
func (r *MinipoolRequester) GetExitDetails() (*types.ApiResponse[api.MinipoolExitDetailsData], error) {
	return client.SendGetRequest[api.MinipoolExitDetailsData](r, "exit/details", "GetExitDetails", nil)
}

// Exit minipools
func (r *MinipoolRequester) Exit(addresses []common.Address) (*types.ApiResponse[types.SuccessData], error) {
	return sendMultiMinipoolRequest[types.SuccessData](r, "exit", "Exit", addresses, nil)
}

// Import a validator private key for a vacant minipool
func (r *MinipoolRequester) ImportKey(address common.Address, mnemonic string) (*types.ApiResponse[types.SuccessData], error) {
	args := map[string]string{
		"address":  address.Hex(),
		"mnemonic": mnemonic,
	}
	return client.SendGetRequest[types.SuccessData](r, "import-key", "ImportKey", args)
}

// Get promote details
func (r *MinipoolRequester) GetPromoteDetails() (*types.ApiResponse[api.MinipoolPromoteDetailsData], error) {
	return client.SendGetRequest[api.MinipoolPromoteDetailsData](r, "promote/details", "GetPromoteDetails", nil)
}

// Promote minipools
func (r *MinipoolRequester) Promote(addresses []common.Address) (*types.ApiResponse[types.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "promote", "Promote", addresses, nil)
}

// Get reduce bond details
func (r *MinipoolRequester) GetReduceBondDetails() (*types.ApiResponse[api.MinipoolReduceBondDetailsData], error) {
	return client.SendGetRequest[api.MinipoolReduceBondDetailsData](r, "reduce-bond/details", "GetReduceBondDetails", nil)
}

// Reduce bond on minipools
func (r *MinipoolRequester) ReduceBond(addresses []common.Address) (*types.ApiResponse[types.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "reduce-bond", "ReduceBond", addresses, nil)
}

// Get refund details
func (r *MinipoolRequester) GetRefundDetails() (*types.ApiResponse[api.MinipoolRefundDetailsData], error) {
	return client.SendGetRequest[api.MinipoolRefundDetailsData](r, "refund/details", "GetRefundDetails", nil)
}

// Refund ETH from minipools
func (r *MinipoolRequester) Refund(addresses []common.Address) (*types.ApiResponse[types.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "refund", "Refund", addresses, nil)
}

// Get rescue dissolved details
func (r *MinipoolRequester) GetRescueDissolvedDetails() (*types.ApiResponse[api.MinipoolRescueDissolvedDetailsData], error) {
	return client.SendGetRequest[api.MinipoolRescueDissolvedDetailsData](r, "rescue-dissolved/details", "GetRescueDissolvedDetails", nil)
}

// Rescue dissolved minipools
func (r *MinipoolRequester) RescueDissolved(addresses []common.Address, depositAmounts []*big.Int) (*types.ApiResponse[types.BatchTxInfoData], error) {
	amounts := make([]string, len(depositAmounts))
	for i, amount := range depositAmounts {
		amounts[i] = amount.String()
	}
	args := map[string]string{
		"deposit-amounts": strings.Join(amounts, ","),
	}
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "rescue-dissolved", "RescueDissolved", addresses, args)
}

// Get stake details
func (r *MinipoolRequester) GetStakeDetails() (*types.ApiResponse[api.MinipoolStakeDetailsData], error) {
	return client.SendGetRequest[api.MinipoolStakeDetailsData](r, "stake/details", "GetStakeDetails", nil)
}

// Stake minipools
func (r *MinipoolRequester) Stake(addresses []common.Address) (*types.ApiResponse[types.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[types.BatchTxInfoData](r, "stake", "Stake", addresses, nil)
}

// Get minipool status
func (r *MinipoolRequester) Status() (*types.ApiResponse[api.MinipoolStatusData], error) {
	return client.SendGetRequest[api.MinipoolStatusData](r, "status", "Status", nil)
}

// Get the artifacts necessary for vanity address searching
func (r *MinipoolRequester) GetVanityArtifacts(nodeAddressStr string) (*types.ApiResponse[api.MinipoolVanityArtifactsData], error) {
	args := map[string]string{
		"node-address": nodeAddressStr,
	}
	return client.SendGetRequest[api.MinipoolVanityArtifactsData](r, "vanity-artifacts", "GetVanityArtifacts", args)
}

// Submit a minipool request that takes in a list of addresses and returns whatever type is requested
func sendMultiMinipoolRequest[DataType any](r *MinipoolRequester, method string, requestName string, addresses []common.Address, args map[string]string) (*types.ApiResponse[DataType], error) {
	if args == nil {
		args = map[string]string{}
	}
	args["addresses"] = client.MakeBatchArg(addresses)
	return client.SendGetRequest[DataType](r, method, requestName, args)
}
