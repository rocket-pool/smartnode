package rocketpool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"

	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Get minipool status
func (c *Client) MinipoolStatus() (api.MinipoolStatusResponse, error) {
	responseBytes, err := c.callAPI("minipool status")
	if err != nil {
		return api.MinipoolStatusResponse{}, fmt.Errorf("Could not get minipool status: %w", err)
	}
	var response api.MinipoolStatusResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.MinipoolStatusResponse{}, fmt.Errorf("Could not decode minipool status response: %w", err)
	}
	if response.Error != "" {
		return api.MinipoolStatusResponse{}, fmt.Errorf("Could not get minipool status: %s", response.Error)
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
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-refund %s", address.Hex()))
	if err != nil {
		return api.CanRefundMinipoolResponse{}, fmt.Errorf("Could not get can refund minipool status: %w", err)
	}
	var response api.CanRefundMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanRefundMinipoolResponse{}, fmt.Errorf("Could not decode can refund minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CanRefundMinipoolResponse{}, fmt.Errorf("Could not get can refund minipool status: %s", response.Error)
	}
	return response, nil
}

// Refund ETH from a minipool
func (c *Client) RefundMinipool(address common.Address) (api.RefundMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool refund %s", address.Hex()))
	if err != nil {
		return api.RefundMinipoolResponse{}, fmt.Errorf("Could not refund minipool: %w", err)
	}
	var response api.RefundMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.RefundMinipoolResponse{}, fmt.Errorf("Could not decode refund minipool response: %w", err)
	}
	if response.Error != "" {
		return api.RefundMinipoolResponse{}, fmt.Errorf("Could not refund minipool: %s", response.Error)
	}
	return response, nil
}

// Check whether a minipool is eligible for staking
func (c *Client) CanStakeMinipool(address common.Address) (api.CanStakeMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-stake %s", address.Hex()))
	if err != nil {
		return api.CanStakeMinipoolResponse{}, fmt.Errorf("Could not get can stake minipool status: %w", err)
	}
	var response api.CanStakeMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanStakeMinipoolResponse{}, fmt.Errorf("Could not decode can stake minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CanStakeMinipoolResponse{}, fmt.Errorf("Could not get can stake minipool status: %s", response.Error)
	}
	return response, nil
}

// Stake a minipool
func (c *Client) StakeMinipool(address common.Address) (api.StakeMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool stake %s", address.Hex()))
	if err != nil {
		return api.StakeMinipoolResponse{}, fmt.Errorf("Could not stake minipool: %w", err)
	}
	var response api.StakeMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.StakeMinipoolResponse{}, fmt.Errorf("Could not decode stake minipool response: %w", err)
	}
	if response.Error != "" {
		return api.StakeMinipoolResponse{}, fmt.Errorf("Could not stake minipool: %s", response.Error)
	}
	return response, nil
}

// Check whether a minipool is eligible for promotion
func (c *Client) CanPromoteMinipool(address common.Address) (api.CanPromoteMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-promote %s", address.Hex()))
	if err != nil {
		return api.CanPromoteMinipoolResponse{}, fmt.Errorf("Could not get can promote minipool status: %w", err)
	}
	var response api.CanPromoteMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanPromoteMinipoolResponse{}, fmt.Errorf("Could not decode can promote minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CanPromoteMinipoolResponse{}, fmt.Errorf("Could not get can promote minipool status: %s", response.Error)
	}
	return response, nil
}

// Promote a minipool
func (c *Client) PromoteMinipool(address common.Address) (api.PromoteMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool promote %s", address.Hex()))
	if err != nil {
		return api.PromoteMinipoolResponse{}, fmt.Errorf("Could not promote minipool: %w", err)
	}
	var response api.PromoteMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.PromoteMinipoolResponse{}, fmt.Errorf("Could not decode promote minipool response: %w", err)
	}
	if response.Error != "" {
		return api.PromoteMinipoolResponse{}, fmt.Errorf("Could not promote minipool: %s", response.Error)
	}
	return response, nil
}

// Check whether a minipool can be dissolved
func (c *Client) CanDissolveMinipool(address common.Address) (api.CanDissolveMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-dissolve %s", address.Hex()))
	if err != nil {
		return api.CanDissolveMinipoolResponse{}, fmt.Errorf("Could not get can dissolve minipool status: %w", err)
	}
	var response api.CanDissolveMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanDissolveMinipoolResponse{}, fmt.Errorf("Could not decode can dissolve minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CanDissolveMinipoolResponse{}, fmt.Errorf("Could not get can dissolve minipool status: %s", response.Error)
	}
	return response, nil
}

// Dissolve a minipool
func (c *Client) DissolveMinipool(address common.Address) (api.DissolveMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool dissolve %s", address.Hex()))
	if err != nil {
		return api.DissolveMinipoolResponse{}, fmt.Errorf("Could not dissolve minipool: %w", err)
	}
	var response api.DissolveMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.DissolveMinipoolResponse{}, fmt.Errorf("Could not decode dissolve minipool response: %w", err)
	}
	if response.Error != "" {
		return api.DissolveMinipoolResponse{}, fmt.Errorf("Could not dissolve minipool: %s", response.Error)
	}
	return response, nil
}

// Check whether a minipool can be exited
func (c *Client) CanExitMinipool(address common.Address) (api.CanExitMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-exit %s", address.Hex()))
	if err != nil {
		return api.CanExitMinipoolResponse{}, fmt.Errorf("Could not get can exit minipool status: %w", err)
	}
	var response api.CanExitMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanExitMinipoolResponse{}, fmt.Errorf("Could not decode can exit minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CanExitMinipoolResponse{}, fmt.Errorf("Could not get can exit minipool status: %s", response.Error)
	}
	return response, nil
}

// Exit a minipool
func (c *Client) ExitMinipool(address common.Address) (api.ExitMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool exit %s", address.Hex()))
	if err != nil {
		return api.ExitMinipoolResponse{}, fmt.Errorf("Could not exit minipool: %w", err)
	}
	var response api.ExitMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ExitMinipoolResponse{}, fmt.Errorf("Could not decode exit minipool response: %w", err)
	}
	if response.Error != "" {
		return api.ExitMinipoolResponse{}, fmt.Errorf("Could not exit minipool: %s", response.Error)
	}
	return response, nil
}

// Check all of the node's minipools for closure eligibility, and return the details of the closeable ones
func (c *Client) GetMinipoolCloseDetailsForNode() (api.GetMinipoolCloseDetailsForNodeResponse, error) {
	responseBytes, err := c.callAPI("minipool get-minipool-close-details-for-node")
	if err != nil {
		return api.GetMinipoolCloseDetailsForNodeResponse{}, fmt.Errorf("Could not get get-minipool-close-details-for-node status: %w", err)
	}
	var response api.GetMinipoolCloseDetailsForNodeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetMinipoolCloseDetailsForNodeResponse{}, fmt.Errorf("Could not decode get-minipool-close-details-for-node response: %w", err)
	}
	if response.Error != "" {
		return api.GetMinipoolCloseDetailsForNodeResponse{}, fmt.Errorf("Could not get get-minipool-close-details-for-node status: %s", response.Error)
	}
	return response, nil
}

// Close a minipool
func (c *Client) CloseMinipool(address common.Address) (api.CloseMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool close %s", address.Hex()))
	if err != nil {
		return api.CloseMinipoolResponse{}, fmt.Errorf("Could not close minipool: %w", err)
	}
	var response api.CloseMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CloseMinipoolResponse{}, fmt.Errorf("Could not decode close minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CloseMinipoolResponse{}, fmt.Errorf("Could not close minipool: %s", response.Error)
	}
	return response, nil
}

// Check whether a minipool can have its delegate upgraded
func (c *Client) CanDelegateUpgradeMinipool(address common.Address) (api.CanDelegateUpgradeResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-delegate-upgrade %s", address.Hex()))
	if err != nil {
		return api.CanDelegateUpgradeResponse{}, fmt.Errorf("Could not get can delegate upgrade minipool status: %w", err)
	}
	var response api.CanDelegateUpgradeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanDelegateUpgradeResponse{}, fmt.Errorf("Could not decode can delegate upgrade minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CanDelegateUpgradeResponse{}, fmt.Errorf("Could not get can delegate upgrade minipool status: %s", response.Error)
	}
	return response, nil
}

// Upgrade a minipool delegate
func (c *Client) DelegateUpgradeMinipool(address common.Address) (api.DelegateUpgradeResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool delegate-upgrade %s", address.Hex()))
	if err != nil {
		return api.DelegateUpgradeResponse{}, fmt.Errorf("Could not upgrade delegate for minipool: %w", err)
	}
	var response api.DelegateUpgradeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.DelegateUpgradeResponse{}, fmt.Errorf("Could not decode upgrade delegate minipool response: %w", err)
	}
	if response.Error != "" {
		return api.DelegateUpgradeResponse{}, fmt.Errorf("Could not upgrade delegate for minipool: %s", response.Error)
	}
	return response, nil
}

// Check whether a minipool can have its auto-upgrade setting changed
func (c *Client) CanSetUseLatestDelegateMinipool(address common.Address) (api.CanSetUseLatestDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-set-use-latest-delegate %s", address.Hex()))
	if err != nil {
		return api.CanSetUseLatestDelegateResponse{}, fmt.Errorf("Could not get can set use latest delegate for minipool status: %w", err)
	}
	var response api.CanSetUseLatestDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanSetUseLatestDelegateResponse{}, fmt.Errorf("Could not decode can set use latest delegate for minipool response: %w", err)
	}
	if response.Error != "" {
		return api.CanSetUseLatestDelegateResponse{}, fmt.Errorf("Could not get can set use latest delegate for minipool status: %s", response.Error)
	}
	return response, nil
}

// Change a minipool's auto-upgrade setting
func (c *Client) SetUseLatestDelegateMinipool(address common.Address) (api.SetUseLatestDelegateResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool set-use-latest-delegate %s", address.Hex()))
	if err != nil {
		return api.SetUseLatestDelegateResponse{}, fmt.Errorf("Could not set use latest delegate for minipool: %w", err)
	}
	var response api.SetUseLatestDelegateResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.SetUseLatestDelegateResponse{}, fmt.Errorf("Could not decode set use latest delegate for minipool response: %w", err)
	}
	if response.Error != "" {
		return api.SetUseLatestDelegateResponse{}, fmt.Errorf("Could not set use latest delegate for minipool: %s", response.Error)
	}
	return response, nil
}

// Get the artifacts necessary for vanity address searching
func (c *Client) GetVanityArtifacts(depositAmount *big.Int, nodeAddress string) (api.GetVanityArtifactsResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool get-vanity-artifacts %s %s", depositAmount.String(), nodeAddress))
	if err != nil {
		return api.GetVanityArtifactsResponse{}, fmt.Errorf("Could not get vanity artifacts: %w", err)
	}
	var response api.GetVanityArtifactsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetVanityArtifactsResponse{}, fmt.Errorf("Could not decode get vanity artifacts response: %w", err)
	}
	if response.Error != "" {
		return api.GetVanityArtifactsResponse{}, fmt.Errorf("Could not get vanity artifacts: %s", response.Error)
	}
	return response, nil
}

// Check whether the minipool can begin the bond reduction process
func (c *Client) CanBeginReduceBondAmount(address common.Address, newBondAmountWei *big.Int) (api.CanBeginReduceBondAmountResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-begin-reduce-bond-amount %s %s", address.Hex(), newBondAmountWei.String()))
	if err != nil {
		return api.CanBeginReduceBondAmountResponse{}, fmt.Errorf("Could not get can begin reduce bond amount status: %w", err)
	}
	var response api.CanBeginReduceBondAmountResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanBeginReduceBondAmountResponse{}, fmt.Errorf("Could not decode can begin reduce bond status amount response: %w", err)
	}
	if response.Error != "" {
		return api.CanBeginReduceBondAmountResponse{}, fmt.Errorf("Could not get can begin reduce bond amount status: %s", response.Error)
	}
	return response, nil
}

// Begin the bond reduction process for a minipool
func (c *Client) BeginReduceBondAmount(address common.Address, newBondAmountWei *big.Int) (api.BeginReduceBondAmountResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool begin-reduce-bond-amount %s %s", address.Hex(), newBondAmountWei.String()))
	if err != nil {
		return api.BeginReduceBondAmountResponse{}, fmt.Errorf("Could not begin reduce bond amount: %w", err)
	}
	var response api.BeginReduceBondAmountResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.BeginReduceBondAmountResponse{}, fmt.Errorf("Could not decode begin reduce bond amount response: %w", err)
	}
	if response.Error != "" {
		return api.BeginReduceBondAmountResponse{}, fmt.Errorf("Could not begin reduce bond amount: %s", response.Error)
	}
	return response, nil
}

// Check if a minipool's bond can be reduced
func (c *Client) CanReduceBondAmount(address common.Address) (api.CanReduceBondAmountResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-reduce-bond-amount %s", address.Hex()))
	if err != nil {
		return api.CanReduceBondAmountResponse{}, fmt.Errorf("Could not get can reduce bond amount status: %w", err)
	}
	var response api.CanReduceBondAmountResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanReduceBondAmountResponse{}, fmt.Errorf("Could not decode can reduce bond amount response: %w", err)
	}
	if response.Error != "" {
		return api.CanReduceBondAmountResponse{}, fmt.Errorf("Could not get can reduce bond amount status: %s", response.Error)
	}
	return response, nil
}

// Reduce a minipool's bond
func (c *Client) ReduceBondAmount(address common.Address) (api.ReduceBondAmountResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool reduce-bond-amount %s", address.Hex()))
	if err != nil {
		return api.ReduceBondAmountResponse{}, fmt.Errorf("Could not reduce bond amount: %w", err)
	}
	var response api.ReduceBondAmountResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ReduceBondAmountResponse{}, fmt.Errorf("Could not decode reduce bond amount response: %w", err)
	}
	if response.Error != "" {
		return api.ReduceBondAmountResponse{}, fmt.Errorf("Could not reduce bond amount: %s", response.Error)
	}
	return response, nil
}

// Get the balance distribution details for all of the node's minipools
func (c *Client) GetDistributeBalanceDetails() (api.GetDistributeBalanceDetailsResponse, error) {
	responseBytes, err := c.callAPI("minipool get-distribute-balance-details")
	if err != nil {
		return api.GetDistributeBalanceDetailsResponse{}, fmt.Errorf("Could not get distribute balance details: %w", err)
	}
	var response api.GetDistributeBalanceDetailsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetDistributeBalanceDetailsResponse{}, fmt.Errorf("Could not decode get distribute balance details response: %w", err)
	}
	if response.Error != "" {
		return api.GetDistributeBalanceDetailsResponse{}, fmt.Errorf("Could not get distribute balance details: %s", response.Error)
	}
	return response, nil
}

// Distribute a minipool's ETH balance
func (c *Client) DistributeBalance(address common.Address) (api.DistributeBalanceResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool distribute-balance %s", address.Hex()))
	if err != nil {
		return api.DistributeBalanceResponse{}, fmt.Errorf("Could not get distribute balance status: %w", err)
	}
	var response api.DistributeBalanceResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.DistributeBalanceResponse{}, fmt.Errorf("Could not decode distribute balance response: %w", err)
	}
	if response.Error != "" {
		return api.DistributeBalanceResponse{}, fmt.Errorf("Could not get distribute balance status: %s", response.Error)
	}
	return response, nil
}

// Import a validator private key for a vacant minipool
func (c *Client) ImportKey(address common.Address, mnemonic string) (api.ChangeWithdrawalCredentialsResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool import-key %s", address.Hex()), mnemonic)
	if err != nil {
		return api.ChangeWithdrawalCredentialsResponse{}, fmt.Errorf("Could not import validator key: %w", err)
	}
	var response api.ChangeWithdrawalCredentialsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ChangeWithdrawalCredentialsResponse{}, fmt.Errorf("Could not decode import-key response: %w", err)
	}
	if response.Error != "" {
		return api.ChangeWithdrawalCredentialsResponse{}, fmt.Errorf("Could not import validator key: %s", response.Error)
	}
	return response, nil
}

// Check whether a solo validator's withdrawal creds can be migrated to a minipool address
func (c *Client) CanChangeWithdrawalCredentials(address common.Address, mnemonic string) (api.CanChangeWithdrawalCredentialsResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool can-change-withdrawal-creds %s", address.Hex()), mnemonic)
	if err != nil {
		return api.CanChangeWithdrawalCredentialsResponse{}, fmt.Errorf("Could not get can-change-withdrawal-creds status: %w", err)
	}
	var response api.CanChangeWithdrawalCredentialsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.CanChangeWithdrawalCredentialsResponse{}, fmt.Errorf("Could not decode can-change-withdrawal-creds response: %w", err)
	}
	if response.Error != "" {
		return api.CanChangeWithdrawalCredentialsResponse{}, fmt.Errorf("Could not get can-change-withdrawal-creds status: %s", response.Error)
	}
	return response, nil
}

// Migrate a solo validator's withdrawal creds to a minipool address
func (c *Client) ChangeWithdrawalCredentials(address common.Address, mnemonic string) (api.ChangeWithdrawalCredentialsResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool change-withdrawal-creds %s", address.Hex()), mnemonic)
	if err != nil {
		return api.ChangeWithdrawalCredentialsResponse{}, fmt.Errorf("Could not change withdrawal creds: %w", err)
	}
	var response api.ChangeWithdrawalCredentialsResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.ChangeWithdrawalCredentialsResponse{}, fmt.Errorf("Could not decode change-withdrawal-creds response: %w", err)
	}
	if response.Error != "" {
		return api.ChangeWithdrawalCredentialsResponse{}, fmt.Errorf("Could not change withdrawal creds: %s", response.Error)
	}
	return response, nil
}

// Check all of the node's minipools for rescue eligibility, and return the details of the rescuable ones
func (c *Client) GetMinipoolRescueDissolvedDetailsForNode() (api.GetMinipoolRescueDissolvedDetailsForNodeResponse, error) {
	responseBytes, err := c.callAPI("minipool get-rescue-dissolved-details-for-node")
	if err != nil {
		return api.GetMinipoolRescueDissolvedDetailsForNodeResponse{}, fmt.Errorf("Could not get get-minipool-rescue-dissolved-details-for-node status: %w", err)
	}
	var response api.GetMinipoolRescueDissolvedDetailsForNodeResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetMinipoolRescueDissolvedDetailsForNodeResponse{}, fmt.Errorf("Could not decode get-minipool-rescue-dissolved-details-for-node response: %w", err)
	}
	if response.Error != "" {
		return api.GetMinipoolRescueDissolvedDetailsForNodeResponse{}, fmt.Errorf("Could not get get-minipool-rescue-dissolved-details-for-node status: %s", response.Error)
	}
	return response, nil
}

// Rescue a dissolved minipool by depositing ETH for it to the Beacon deposit contract
func (c *Client) RescueDissolvedMinipool(address common.Address, amount *big.Int, submit bool) (api.RescueDissolvedMinipoolResponse, error) {
	responseBytes, err := c.callAPI(fmt.Sprintf("minipool rescue-dissolved %s %s %t", address.Hex(), amount.String(), submit))
	if err != nil {
		return api.RescueDissolvedMinipoolResponse{}, fmt.Errorf("Could not rescue dissolved minipool: %w", err)
	}
	var response api.RescueDissolvedMinipoolResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.RescueDissolvedMinipoolResponse{}, fmt.Errorf("Could not decode rescue dissolved minipool response: %w", err)
	}
	if response.Error != "" {
		return api.RescueDissolvedMinipoolResponse{}, fmt.Errorf("Could not rescue dissolved minipool: %s", response.Error)
	}
	return response, nil
}

func (c *Client) GetBondReductionEnabled() (api.GetBondReductionEnabledResponse, error) {
	responseBytes, err := c.callAPI("minipool get-bond-reduction-enabled")
	if err != nil {
		return api.GetBondReductionEnabledResponse{}, fmt.Errorf("Could not get bond reduction enabled status: %w", err)
	}
	var response api.GetBondReductionEnabledResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GetBondReductionEnabledResponse{}, fmt.Errorf("Could not decode bond reduction enabled response: %w", err)
	}
	if response.Error != "" {
		return api.GetBondReductionEnabledResponse{}, fmt.Errorf("Could not get bond reduction enabled status: %s", response.Error)
	}
	return response, nil
}
