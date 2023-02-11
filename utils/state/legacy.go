package state

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
)

// Gets the details for a node using the efficient multicall contract
func GetNativeNodeDetails_Legacy(rp *rocketpool.RocketPool, nodeAddress common.Address, multicallerAddress common.Address, opts *bind.CallOpts) (node.NativeNodeDetails, error) {
	// Get contracts
	rocketNodeManager, err := rp.GetContract("rocketNodeManager", nil)
	if err != nil {
		return node.NativeNodeDetails{}, err
	}
	rocketMinipoolManager, err := rp.GetContract("rocketMinipoolManager", nil)
	if err != nil {
		return node.NativeNodeDetails{}, err
	}
	rocketNodeDistributorFactory, err := rp.GetContract("rocketNodeDistributorFactory", nil)
	if err != nil {
		return node.NativeNodeDetails{}, err
	}
	rocketTokenRETH, err := rp.GetContract("rocketTokenRETH", nil)
	if err != nil {
		return node.NativeNodeDetails{}, err
	}
	rocketTokenRPL, err := rp.GetContract("rocketTokenRPL", nil)
	if err != nil {
		return node.NativeNodeDetails{}, err
	}
	rocketTokenRPLFixedSupply, err := rp.GetContract("rocketTokenRPLFixedSupply", nil)
	if err != nil {
		return node.NativeNodeDetails{}, err
	}

	details := node.NativeNodeDetails{}
	details.NodeAddress = nodeAddress
	mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
	if err != nil {
		return node.NativeNodeDetails{}, err
	}

	avgFee := big.NewInt(0)
	mc.AddCall("GetNodeExists", rocketNodeManager, &details.Exists, "getNodeExists", nodeAddress)
	mc.AddCall("GetNodeRegistrationTimeRaw", rocketNodeManager, &details.RegistrationTime, "getNodeRegistrationTime", nodeAddress)
	mc.AddCall("GetNodeTimezoneLocation", rocketNodeManager, &details.TimezoneLocation, "getNodeTimezoneLocation", nodeAddress)
	mc.AddCall("GetFeeDistributorInitialized", rocketNodeManager, &details.FeeDistributorInitialised, "getFeeDistributorInitialised", nodeAddress)
	mc.AddCall("GetDistributorAddress", rocketNodeDistributorFactory, &details.FeeDistributorAddress, "getProxyAddress", nodeAddress)
	mc.AddCall("GetNodeAverageFeeRaw", rocketNodeManager, &avgFee, "getAverageNodeFee", nodeAddress)
	mc.AddCall("GetRewardNetworkRaw", rocketNodeManager, &details.RewardNetwork, "getRewardNetwork", nodeAddress)
	mc.AddCall("GetNodeRPLStake", rocketNodeManager, &details.RplStake, "getNodeRPLStake", nodeAddress)
	mc.AddCall("GetNodeEffectiveRPLStake", rocketNodeManager, &details.EffectiveRPLStake, "getNodeEffectiveRPLStake", nodeAddress)
	mc.AddCall("GetNodeMinimumRPLStake", rocketNodeManager, &details.MinimumRPLStake, "getNodeMinimumRPLStake", nodeAddress)
	mc.AddCall("GetNodeMaximumRPLStake", rocketNodeManager, &details.MaximumRPLStake, "getNodeMaximumRPLStake", nodeAddress)
	mc.AddCall("GetNodeMinipoolCountRaw", rocketMinipoolManager, &details.MinipoolCount, "getNodeMinipoolCount", nodeAddress)
	mc.AddCall("GetBalanceRETH", rocketTokenRETH, &details.BalanceRETH, "balanceOf", nodeAddress)
	mc.AddCall("GetBalanceRPL", rocketTokenRPL, &details.BalanceRPL, "balanceOf", nodeAddress)
	mc.AddCall("GetBalanceOldRPL", rocketTokenRPLFixedSupply, &details.BalanceOldRPL, "balanceOf", nodeAddress)
	mc.AddCall("GetNodeWithdrawalAddress", rp.RocketStorageContract, &details.WithdrawalAddress, "getNodeWithdrawalAddress", nodeAddress)
	mc.AddCall("GetNodePendingWithdrawalAddress", rp.RocketStorageContract, &details.PendingWithdrawalAddress, "getNodePendingWithdrawalAddress", nodeAddress)
	mc.AddCall("GetSmoothingPoolRegistrationState", rocketNodeManager, &details.SmoothingPoolRegistrationState, "getSmoothingPoolRegistrationState", nodeAddress)
	mc.AddCall("GetSmoothingPoolRegistrationChangedRaw", rocketNodeManager, &details.SmoothingPoolRegistrationChanged, "getSmoothingPoolRegistrationChanged", nodeAddress)
	_, err = mc.FlexibleCall(true)
	if err != nil {
		return node.NativeNodeDetails{}, fmt.Errorf("error executing multicall: %w", err)
	}

	// Get the node's ETH balance
	details.BalanceETH, err = rp.Client.BalanceAt(context.Background(), nodeAddress, opts.BlockNumber)
	if err != nil {
		return node.NativeNodeDetails{}, err
	}

	// Fix the effective stake
	if details.EffectiveRPLStake.Cmp(details.MinimumRPLStake) == -1 {
		details.EffectiveRPLStake.SetUint64(0)
	}

	// Get the user and node portions of the distributor balance
	distributorBalance, err := rp.Client.BalanceAt(context.Background(), details.FeeDistributorAddress, opts.BlockNumber)
	if err != nil {
		return node.NativeNodeDetails{}, err
	}
	if distributorBalance.Cmp(big.NewInt(0)) > 0 {
		halfBalance := big.NewInt(0)
		halfBalance.Div(distributorBalance, big.NewInt(2))
		nodeShare := big.NewInt(0)
		nodeShare.Mul(halfBalance, avgFee)
		nodeShare.Div(nodeShare, eth.EthToWei(1))
		nodeShare.Add(nodeShare, halfBalance)
		details.DistributorBalanceNodeETH = nodeShare
		userShare := big.NewInt(0)
		userShare.Sub(distributorBalance, nodeShare)
		details.DistributorBalanceUserETH = userShare
	}

	return details, nil
}
