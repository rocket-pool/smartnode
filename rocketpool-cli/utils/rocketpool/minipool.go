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
	route  string
}

func NewMinipoolRequester(client *http.Client) *MinipoolRequester {
	return &MinipoolRequester{
		client: client,
		route:  "minipool",
	}
}

// Get begin reduce bond details
func (r *MinipoolRequester) GetBeginReduceBondDetails() (*api.ApiResponse[api.MinipoolBeginReduceBondDetailsData], error) {
	return sendDetailsRequest[api.MinipoolBeginReduceBondDetailsData](r, "begin-reduce-bond", "BeginReduceBond")
}

// Begin reduce bond on minipools
func (r *MinipoolRequester) BeginReduceBond(addresses []common.Address, newBondAmount *big.Int) (*api.ApiResponse[api.BatchTxInfoData], error) {
	args := map[string]string{
		"newBondAmount": newBondAmount.String(),
	}
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "begin-reduce-bond", "BeginReduceBond", addresses, args)
}

// Verify that migrating a solo validator's withdrawal creds to a minipool address is possible
func (r *MinipoolRequester) CanChangeWithdrawalCredentials(address common.Address, mnemonic string) (*api.ApiResponse[api.MinipoolCanChangeWithdrawalCredentialsData], error) {
	method := "change-withdrawal-creds/verify"
	args := map[string]string{
		"address":  address.Hex(),
		"mnemonic": mnemonic,
	}
	response, err := SendGetRequest[api.MinipoolCanChangeWithdrawalCredentialsData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Minipool CanChangeWithdrawalCredentials request: %w", err)
	}
	return response, nil
}

// Migrate a solo validator's withdrawal creds to a minipool address
func (r *MinipoolRequester) ChangeWithdrawalCredentials(address common.Address, mnemonic string) (*api.ApiResponse[api.SuccessData], error) {
	method := "change-withdrawal-creds"
	args := map[string]string{
		"address":  address.Hex(),
		"mnemonic": mnemonic,
	}
	response, err := SendGetRequest[api.SuccessData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Minipool ChangeWithdrawalCredentials request: %w", err)
	}
	return response, nil
}

// Get close details
func (r *MinipoolRequester) GetCloseDetails() (*api.ApiResponse[api.MinipoolCloseDetailsData], error) {
	return sendDetailsRequest[api.MinipoolCloseDetailsData](r, "close", "Close")
}

// Close minipools
func (r *MinipoolRequester) Close(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "close", "Close", addresses, nil)
}

// Get delegate details
func (r *MinipoolRequester) GetDelegateDetails() (*api.ApiResponse[api.MinipoolDelegateDetailsData], error) {
	return sendDetailsRequest[api.MinipoolDelegateDetailsData](r, "delegate", "Delegate")
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
	return sendDetailsRequest[api.MinipoolDistributeDetailsData](r, "distribute", "Distribute")
}

// Distribute minipool balances
func (r *MinipoolRequester) Distribute(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "distribute", "Distribute", addresses, nil)
}

// Get dissolve details
func (r *MinipoolRequester) GetDissolveDetails() (*api.ApiResponse[api.MinipoolDissolveDetailsData], error) {
	return sendDetailsRequest[api.MinipoolDissolveDetailsData](r, "dissolve", "Dissolve")
}

// Dissolve minipools
func (r *MinipoolRequester) Dissolve(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "dissolve", "Dissolve", addresses, nil)
}

// Get exit details
func (r *MinipoolRequester) GetExitDetails() (*api.ApiResponse[api.MinipoolExitDetailsData], error) {
	return sendDetailsRequest[api.MinipoolExitDetailsData](r, "exit", "Exit")
}

// Exit minipools
func (r *MinipoolRequester) Exit(addresses []common.Address) (*api.ApiResponse[api.SuccessData], error) {
	return sendMultiMinipoolRequest[api.SuccessData](r, "exit", "Exit", addresses, nil)
}

// Import a validator private key for a vacant minipool
func (r *MinipoolRequester) ImportKey(address common.Address, mnemonic string) (*api.ApiResponse[api.SuccessData], error) {
	method := "import-key"
	args := map[string]string{
		"address":  address.Hex(),
		"mnemonic": mnemonic,
	}
	response, err := SendGetRequest[api.SuccessData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Minipool ImportKey request: %w", err)
	}
	return response, nil
}

// Get promote details
func (r *MinipoolRequester) GetPromoteDetails() (*api.ApiResponse[api.MinipoolPromoteDetailsData], error) {
	return sendDetailsRequest[api.MinipoolPromoteDetailsData](r, "promote", "Promote")
}

// Promote minipools
func (r *MinipoolRequester) Promote(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "promote", "Promote", addresses, nil)
}

// Get reduce bond details
func (r *MinipoolRequester) GetReduceBondDetails() (*api.ApiResponse[api.MinipoolReduceBondDetailsData], error) {
	return sendDetailsRequest[api.MinipoolReduceBondDetailsData](r, "reduce-bond", "ReduceBond")
}

// Reduce bond on minipools
func (r *MinipoolRequester) ReduceBond(addresses []common.Address, newBondAmount *big.Int) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "reduce-bond", "ReduceBond", addresses, nil)
}

// Get refund details
func (r *MinipoolRequester) GetRefundDetails() (*api.ApiResponse[api.MinipoolRefundDetailsData], error) {
	return sendDetailsRequest[api.MinipoolRefundDetailsData](r, "refund", "Refund")
}

// Refund ETH from minipools
func (r *MinipoolRequester) Refund(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "refund", "Refund", addresses, nil)
}

// Get rescue dissolved details
func (r *MinipoolRequester) GetRescueDissolvedDetails() (*api.ApiResponse[api.MinipoolRescueDissolvedDetailsData], error) {
	return sendDetailsRequest[api.MinipoolRescueDissolvedDetailsData](r, "rescue-dissolved", "RescueDissolved")
}

// Rescue dissolved minipools
func (r *MinipoolRequester) RescueDissolved(addresses []common.Address, depositAmounts []*big.Int) (*api.ApiResponse[api.BatchTxInfoData], error) {
	amounts := make([]string, len(depositAmounts))
	for i, amount := range depositAmounts {
		amounts[i] = amount.String()
	}
	args := map[string]string{
		"depositAmounts": strings.Join(amounts, ","),
	}
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "rescue-dissolved", "RescueDissolved", addresses, args)
}

// Get stake details
func (r *MinipoolRequester) GetStakeDetails() (*api.ApiResponse[api.MinipoolStakeDetailsData], error) {
	return sendDetailsRequest[api.MinipoolStakeDetailsData](r, "stake", "Stake")
}

// Stake minipools
func (r *MinipoolRequester) Stake(addresses []common.Address) (*api.ApiResponse[api.BatchTxInfoData], error) {
	return sendMultiMinipoolRequest[api.BatchTxInfoData](r, "stake", "Stake", addresses, nil)
}

// Get minipool status
func (r *MinipoolRequester) Status() (*api.ApiResponse[api.MinipoolStatusData], error) {
	method := "status"
	args := map[string]string{}
	response, err := SendGetRequest[api.MinipoolStatusData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Minipool Status request: %w", err)
	}
	return response, nil
}

// Get the artifacts necessary for vanity address searching
func (r *MinipoolRequester) GetVanityArtifacts(nodeAddress common.Address) (*api.ApiResponse[api.MinipoolVanityArtifactsData], error) {
	method := "vanity-artifacts"
	args := map[string]string{
		"nodeAddress": nodeAddress.Hex(),
	}
	response, err := SendGetRequest[api.MinipoolVanityArtifactsData](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Minipool GetVanityArtifacts request: %w", err)
	}
	return response, nil
}

// Submit a minipool request that asks for details of some route and returns whatever type is requested
func sendDetailsRequest[DataType any](r *MinipoolRequester, method string, requestName string) (*api.ApiResponse[DataType], error) {
	args := map[string]string{}
	response, err := SendGetRequest[DataType](r.client, fmt.Sprintf("%s/%s/details", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Minipool %s Details request: %w", requestName, err)
	}
	return response, nil
}

// Submit a minipool request that takes in a list of addresses and returns whatever type is requested
func sendMultiMinipoolRequest[DataType any](r *MinipoolRequester, method string, requestName string, addresses []common.Address, args map[string]string) (*api.ApiResponse[DataType], error) {
	addressStrings := make([]string, len(addresses))
	for i, address := range addresses {
		addressStrings[i] = address.Hex()
	}
	if args == nil {
		args = map[string]string{}
	}
	args["addresses"] = strings.Join(addressStrings, ",")
	response, err := SendGetRequest[DataType](r.client, fmt.Sprintf("%s/%s", r.route, method), args)
	if err != nil {
		return nil, fmt.Errorf("error during Minipool %s request: %w", requestName, err)
	}
	return response, nil
}
