package state

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	legacyNodeBatchSize  int = 20
	nodeAddressBatchSize int = 1000
)

// Gets the details for a node using the efficient multicall contract
func GetNativeNodeDetails_Legacy(rp *rocketpool.RocketPool, nodeAddress common.Address, multicallerAddress common.Address, opts *bind.CallOpts) (node.NativeNodeDetails, error) {
	contracts, err := NewNetworkContracts(rp, opts)
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
	addNodeDetailsCalls(contracts, mc, &details, nodeAddress, &avgFee)

	_, err = mc.FlexibleCall(true)
	if err != nil {
		return node.NativeNodeDetails{}, fmt.Errorf("error executing multicall: %w", err)
	}

	fixupNodeDetails(rp, &details, avgFee, opts)

	return details, nil
}

// Gets the details for all nodes using the efficient multicall contract
func GetAllNativeNodeDetails_Legacy(rp *rocketpool.RocketPool, multicallerAddress common.Address, opts *bind.CallOpts) ([]node.NativeNodeDetails, error) {
	contracts, err := NewNetworkContracts(rp, opts)
	if err != nil {
		return nil, err
	}

	// Get the list of node addresses
	addresses, err := getNodeAddressesFast(rp, contracts, multicallerAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}
	nodeDetails := make([]node.NativeNodeDetails, len(addresses))

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	// Run the getters in batches
	count := len(addresses)
	for i := 0; i < count; i += legacyNodeBatchSize {
		i := i
		max := i + legacyNodeBatchSize
		if max > count {
			max = count
		}

		avgFees := make([]*big.Int, legacyNodeBatchSize)
		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {

				address := addresses[j]
				details := &nodeDetails[j]
				details.NodeAddress = address

				avgFees[j-i] = big.NewInt(0)
				addNodeDetailsCalls(contracts, mc, details, address, &avgFees[j-i])
			}
			_, err = mc.FlexibleCall(true)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}

			for j := i; j < max; j++ {
				fixupNodeDetails(rp, &nodeDetails[j], avgFees[j-i], opts)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting node details: %w", err)
	}

	return nodeDetails, nil
}

// Get all node addresses using the multicaller
func getNodeAddressesFast(rp *rocketpool.RocketPool, contracts *NetworkContracts, multicallerAddress common.Address, opts *bind.CallOpts) ([]common.Address, error) {
	// Get minipool count
	nodeCount, err := node.GetNodeCount(rp, opts)
	if err != nil {
		return []common.Address{}, err
	}

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	addresses := make([]common.Address, nodeCount)

	// Run the getters in batches
	count := int(nodeCount)
	for i := 0; i < count; i += nodeAddressBatchSize {
		i := i
		max := i + nodeAddressBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, multicallerAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				mc.AddCall(contracts.RocketNodeManager, &addresses[j], "getNodeAt", big.NewInt(int64(j)))
			}
			_, err = mc.FlexibleCall(true)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}

	return addresses, nil
}

// Add all of the calls for the node details to the multicaller
func addNodeDetailsCalls(contracts *NetworkContracts, mc *multicall.MultiCaller, details *node.NativeNodeDetails, address common.Address, avgFee **big.Int) {
	mc.AddCall(contracts.RocketNodeManager, &details.Exists, "getNodeExists", address)
	mc.AddCall(contracts.RocketNodeManager, &details.RegistrationTime, "getNodeRegistrationTime", address)
	mc.AddCall(contracts.RocketNodeManager, &details.TimezoneLocation, "getNodeTimezoneLocation", address)
	mc.AddCall(contracts.RocketNodeManager, &details.FeeDistributorInitialised, "getFeeDistributorInitialised", address)
	mc.AddCall(contracts.RocketNodeDistributorFactory, &details.FeeDistributorAddress, "getProxyAddress", address)
	mc.AddCall(contracts.RocketNodeManager, avgFee, "getAverageNodeFee", address)
	mc.AddCall(contracts.RocketNodeManager, &details.RewardNetwork, "getRewardNetwork", address)
	mc.AddCall(contracts.RocketNodeManager, &details.RplStake, "getNodeRPLStake", address)
	mc.AddCall(contracts.RocketNodeManager, &details.EffectiveRPLStake, "getNodeEffectiveRPLStake", address)
	mc.AddCall(contracts.RocketNodeManager, &details.MinimumRPLStake, "getNodeMinimumRPLStake", address)
	mc.AddCall(contracts.RocketNodeManager, &details.MaximumRPLStake, "getNodeMaximumRPLStake", address)
	mc.AddCall(contracts.RocketMinipoolManager, &details.MinipoolCount, "getNodeMinipoolCount", address)
	mc.AddCall(contracts.RocketTokenRETH, &details.BalanceRETH, "balanceOf", address)
	mc.AddCall(contracts.RocketTokenRPL, &details.BalanceRPL, "balanceOf", address)
	mc.AddCall(contracts.RocketTokenRPLFixedSupply, &details.BalanceOldRPL, "balanceOf", address)
	mc.AddCall(contracts.RocketStorage, &details.WithdrawalAddress, "getNodeWithdrawalAddress", address)
	mc.AddCall(contracts.RocketStorage, &details.PendingWithdrawalAddress, "getNodePendingWithdrawalAddress", address)
	mc.AddCall(contracts.RocketNodeManager, &details.SmoothingPoolRegistrationState, "getSmoothingPoolRegistrationState", address)
	mc.AddCall(contracts.RocketNodeManager, &details.SmoothingPoolRegistrationChanged, "getSmoothingPoolRegistrationChanged", address)
}

// Fixes a legacy node details struct with supplemental logic
func fixupNodeDetails(rp *rocketpool.RocketPool, details *node.NativeNodeDetails, avgFee *big.Int, opts *bind.CallOpts) error {
	address := details.NodeAddress

	var err error

	// Get the node's ETH balance
	details.BalanceETH, err = rp.Client.BalanceAt(context.Background(), address, opts.BlockNumber)
	if err != nil {
		return err
	}

	// Fix the effective stake
	if details.EffectiveRPLStake.Cmp(details.MinimumRPLStake) == -1 {
		details.EffectiveRPLStake.SetUint64(0)
	}

	// Get the user and node portions of the distributor balance
	distributorBalance, err := rp.Client.BalanceAt(context.Background(), details.FeeDistributorAddress, opts.BlockNumber)
	if err != nil {
		return err
	}
	if distributorBalance.Cmp(zero) > 0 {
		halfBalance := big.NewInt(0)
		halfBalance.Div(distributorBalance, two)
		nodeShare := big.NewInt(0)
		nodeShare.Mul(halfBalance, avgFee)
		nodeShare.Div(nodeShare, oneInWei)
		nodeShare.Add(nodeShare, halfBalance)
		details.DistributorBalanceNodeETH = nodeShare
		userShare := big.NewInt(0)
		userShare.Sub(distributorBalance, nodeShare)
		details.DistributorBalanceUserETH = userShare
	}

	return nil
}
