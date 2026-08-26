package rocketpool

import (
	"math/big"
	"net/url"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get minipool status
func (c *Client) MinipoolStatus() (api.MinipoolStatusResponse, error) {
	response, err := c.callAPI[api.MinipoolStatusResponse]("GET", "/api/minipool/status", nil, "Could not get minipool status")
	if err != nil {
		return response, err
	}
	for i := 0; i < len(response.Minipools); i++ {
		mp := &response.Minipools[i]
		if mp.Node.DepositBalance == nil {
			mp.Node.DepositBalance = big.NewInt(0)
		}
		if mp.Node.RefundBalance == nil {
			mp.Node.RefundBalance = big.NewInt(0)
		}
		if mp.User.DepositBalance == nil {
			mp.User.DepositBalance = big.NewInt(0)
		}
		if mp.Balances.ETH == nil {
			mp.Balances.ETH = big.NewInt(0)
		}
		if mp.Balances.RPL == nil {
			mp.Balances.RPL = big.NewInt(0)
		}
		if mp.Balances.RETH == nil {
			mp.Balances.RETH = big.NewInt(0)
		}
		if mp.Balances.FixedSupplyRPL == nil {
			mp.Balances.FixedSupplyRPL = big.NewInt(0)
		}
		if mp.Validator.Balance == nil {
			mp.Validator.Balance = big.NewInt(0)
		}
		if mp.Validator.NodeBalance == nil {
			mp.Validator.NodeBalance = big.NewInt(0)
		}
	}
	return response, nil
}

// Check whether a minipool is eligible for a refund
func (c *Client) CanRefundMinipool(address common.Address) (api.CanRefundMinipoolResponse, error) {
	return c.callAPI[api.CanRefundMinipoolResponse]("GET", "/api/minipool/can-refund", url.Values{"address": {address.Hex()}}, "Could not get can refund minipool status")
}

// Refund ETH from a minipool
func (c *Client) RefundMinipool(address common.Address) (api.RefundMinipoolResponse, error) {
	return c.callAPI[api.RefundMinipoolResponse]("POST", "/api/minipool/refund", url.Values{"address": {address.Hex()}}, "Could not refund minipool")
}

// Check whether a minipool is eligible for staking
func (c *Client) CanStakeMinipool(address common.Address) (api.CanStakeMinipoolResponse, error) {
	return c.callAPI[api.CanStakeMinipoolResponse]("GET", "/api/minipool/can-stake", url.Values{"address": {address.Hex()}}, "Could not get can stake minipool status")
}

// Stake a minipool
func (c *Client) StakeMinipool(address common.Address) (api.StakeMinipoolResponse, error) {
	return c.callAPI[api.StakeMinipoolResponse]("POST", "/api/minipool/stake", url.Values{"address": {address.Hex()}}, "Could not stake minipool")
}

// Check whether a minipool can be dissolved
func (c *Client) CanDissolveMinipool(address common.Address) (api.CanDissolveMinipoolResponse, error) {
	return c.callAPI[api.CanDissolveMinipoolResponse]("GET", "/api/minipool/can-dissolve", url.Values{"address": {address.Hex()}}, "Could not get can dissolve minipool status")
}

// Dissolve a minipool
func (c *Client) DissolveMinipool(address common.Address) (api.DissolveMinipoolResponse, error) {
	return c.callAPI[api.DissolveMinipoolResponse]("POST", "/api/minipool/dissolve", url.Values{"address": {address.Hex()}}, "Could not dissolve minipool")
}

// Check whether a minipool can be exited
func (c *Client) CanExitMinipool(address common.Address) (api.CanExitMinipoolResponse, error) {
	return c.callAPI[api.CanExitMinipoolResponse]("GET", "/api/minipool/can-exit", url.Values{"address": {address.Hex()}}, "Could not get can exit minipool status")
}

// Exit a minipool
func (c *Client) ExitMinipool(address common.Address) (api.ExitMinipoolResponse, error) {
	return c.callAPI[api.ExitMinipoolResponse]("POST", "/api/minipool/exit", url.Values{"address": {address.Hex()}}, "Could not exit minipool")
}

// Check all of the node's minipools for closure eligibility, and return the details of the closeable ones
func (c *Client) GetMinipoolCloseDetailsForNode() (api.GetMinipoolCloseDetailsForNodeResponse, error) {
	return c.callAPI[api.GetMinipoolCloseDetailsForNodeResponse]("GET", "/api/minipool/get-minipool-close-details-for-node", nil, "Could not get get-minipool-close-details-for-node status")
}

// Close a minipool
func (c *Client) CloseMinipool(address common.Address, bundle bool) (api.CloseMinipoolResponse, error) {
	return c.callAPI[api.CloseMinipoolResponse]("POST", "/api/minipool/close", url.Values{
		"address": {address.Hex()},
		"bundle":  {strconv.FormatBool(bundle)},
	}, "Could not close minipool")
}

// Check whether a minipool can have its delegate upgraded
func (c *Client) CanDelegateUpgradeMinipool(address common.Address) (api.CanDelegateUpgradeResponse, error) {
	return c.callAPI[api.CanDelegateUpgradeResponse]("GET", "/api/minipool/can-delegate-upgrade", url.Values{"address": {address.Hex()}}, "Could not get can delegate upgrade minipool status")
}

// Upgrade a minipool delegate
func (c *Client) DelegateUpgradeMinipool(address common.Address) (api.DelegateUpgradeResponse, error) {
	return c.callAPI[api.DelegateUpgradeResponse]("POST", "/api/minipool/delegate-upgrade", url.Values{"address": {address.Hex()}}, "Could not upgrade delegate for minipool")
}

// Check whether a minipool can have its auto-upgrade setting changed
func (c *Client) CanSetUseLatestDelegateMinipool(address common.Address, setLatest bool) (api.CanSetUseLatestDelegateResponse, error) {
	// pass setLatest as well
	return c.callAPI[api.CanSetUseLatestDelegateResponse]("GET", "/api/minipool/can-set-use-latest-delegate", url.Values{"address": {address.Hex()}, "setLatest": {strconv.FormatBool(setLatest)}}, "Could not get can set use latest delegate for minipool status")
}

// Change a minipool's auto-upgrade setting
func (c *Client) SetUseLatestDelegateMinipool(address common.Address) (api.SetUseLatestDelegateResponse, error) {
	return c.callAPI[api.SetUseLatestDelegateResponse]("POST", "/api/minipool/set-use-latest-delegate", url.Values{"address": {address.Hex()}}, "Could not set use latest delegate for minipool")
}

// Get the artifacts necessary for vanity address searching
func (c *Client) GetVanityArtifacts(depositAmount *big.Int, nodeAddress string) (api.GetVanityArtifactsResponse, error) {
	return c.callAPI[api.GetVanityArtifactsResponse]("GET", "/api/minipool/get-vanity-artifacts", url.Values{
		"depositAmount": {depositAmount.String()},
		"nodeAddress":   {nodeAddress},
	}, "Could not get vanity artifacts")
}

// Get the balance distribution details for all of the node's minipools
func (c *Client) GetDistributeBalanceDetails() (api.GetDistributeBalanceDetailsResponse, error) {
	return c.callAPI[api.GetDistributeBalanceDetailsResponse]("GET", "/api/minipool/get-distribute-balance-details", nil, "Could not get distribute balance details")
}

// Distribute a minipool's ETH balance
func (c *Client) DistributeBalance(address common.Address) (api.DistributeBalanceResponse, error) {
	return c.callAPI[api.DistributeBalanceResponse]("POST", "/api/minipool/distribute-balance", url.Values{"address": {address.Hex()}}, "Could not get distribute balance status")
}

// Import a validator private key for a vacant minipool
func (c *Client) ImportKey(address common.Address, mnemonic string) (api.ChangeWithdrawalCredentialsResponse, error) {
	return c.callAPI[api.ChangeWithdrawalCredentialsResponse]("POST", "/api/minipool/import-key", url.Values{
		"address":  {address.Hex()},
		"mnemonic": {mnemonic},
	}, "Could not import validator key")
}

// Check whether a solo validator's withdrawal creds can be migrated to a minipool address
func (c *Client) CanChangeWithdrawalCredentials(address common.Address, mnemonic string) (api.CanChangeWithdrawalCredentialsResponse, error) {
	return c.callAPI[api.CanChangeWithdrawalCredentialsResponse]("GET", "/api/minipool/can-change-withdrawal-creds", url.Values{
		"address":  {address.Hex()},
		"mnemonic": {mnemonic},
	}, "Could not get can-change-withdrawal-creds status")
}

// Migrate a solo validator's withdrawal creds to a minipool address
func (c *Client) ChangeWithdrawalCredentials(address common.Address, mnemonic string) (api.ChangeWithdrawalCredentialsResponse, error) {
	return c.callAPI[api.ChangeWithdrawalCredentialsResponse]("POST", "/api/minipool/change-withdrawal-creds", url.Values{
		"address":  {address.Hex()},
		"mnemonic": {mnemonic},
	}, "Could not change withdrawal creds")
}

// Check all of the node's minipools for rescue eligibility, and return the details of the rescuable ones
func (c *Client) GetMinipoolRescueDissolvedDetailsForNode() (api.GetMinipoolRescueDissolvedDetailsForNodeResponse, error) {
	return c.callAPI[api.GetMinipoolRescueDissolvedDetailsForNodeResponse]("GET", "/api/minipool/get-rescue-dissolved-details-for-node", nil, "Could not get get-minipool-rescue-dissolved-details-for-node status")
}

// Rescue a dissolved minipool by depositing ETH for it to the Beacon deposit contract
func (c *Client) RescueDissolvedMinipool(address common.Address, amount *big.Int, submit bool) (api.RescueDissolvedMinipoolResponse, error) {
	submitStr := "false"
	if submit {
		submitStr = "true"
	}
	return c.callAPI[api.RescueDissolvedMinipoolResponse]("POST", "/api/minipool/rescue-dissolved", url.Values{
		"address": {address.Hex()},
		"amount":  {amount.String()},
		"submit":  {submitStr},
	}, "Could not rescue dissolved minipool")
}
