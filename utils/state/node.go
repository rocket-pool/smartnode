package state

import (
	"context"
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	legacyNodeBatchSize  int = 100
	nodeAddressBatchSize int = 1000
)

// Complete details for a node
type NativeNodeDetails struct {
	Exists                           bool
	RegistrationTime                 *big.Int
	TimezoneLocation                 string
	FeeDistributorInitialised        bool
	FeeDistributorAddress            common.Address
	RewardNetwork                    *big.Int
	RplStake                         *big.Int
	EffectiveRPLStake                *big.Int
	MinimumRPLStake                  *big.Int
	MaximumRPLStake                  *big.Int
	EthMatched                       *big.Int
	EthMatchedLimit                  *big.Int
	MinipoolCount                    *big.Int
	BalanceETH                       *big.Int
	BalanceRETH                      *big.Int
	BalanceRPL                       *big.Int
	BalanceOldRPL                    *big.Int
	DepositCreditBalance             *big.Int
	DistributorBalanceUserETH        *big.Int // Must call CalculateAverageFeeAndDistributorShares to get this
	DistributorBalanceNodeETH        *big.Int // Must call CalculateAverageFeeAndDistributorShares to get this
	WithdrawalAddress                common.Address
	PendingWithdrawalAddress         common.Address
	SmoothingPoolRegistrationState   bool
	SmoothingPoolRegistrationChanged *big.Int
	NodeAddress                      common.Address
	AverageNodeFee                   *big.Int // Must call CalculateAverageFeeAndDistributorShares to get this
	CollateralisationRatio           *big.Int
	DistributorBalance               *big.Int
}

// Gets the details for a node using the efficient multicall contract
func GetNativeNodeDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, nodeAddress common.Address, isAtlasDeployed bool) (NativeNodeDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}
	details := NativeNodeDetails{
		NodeAddress:               nodeAddress,
		AverageNodeFee:            big.NewInt(0),
		CollateralisationRatio:    big.NewInt(0),
		DistributorBalanceUserETH: big.NewInt(0),
		DistributorBalanceNodeETH: big.NewInt(0),
	}

	addNodeDetailsCalls(contracts, contracts.Multicaller, &details, nodeAddress, isAtlasDeployed)

	_, err := contracts.Multicaller.FlexibleCall(true, opts)
	if err != nil {
		return NativeNodeDetails{}, fmt.Errorf("error executing multicall: %w", err)
	}

	// Get the node's ETH balance
	details.BalanceETH, err = rp.Client.BalanceAt(context.Background(), nodeAddress, opts.BlockNumber)
	if err != nil {
		return NativeNodeDetails{}, err
	}

	// Get the distributor balance
	distributorBalance, err := rp.Client.BalanceAt(context.Background(), details.FeeDistributorAddress, opts.BlockNumber)
	if err != nil {
		return NativeNodeDetails{}, err
	}

	// Do some postprocessing on the node data
	details.DistributorBalance = distributorBalance

	// Fix the effective stake
	if details.EffectiveRPLStake.Cmp(details.MinimumRPLStake) == -1 {
		details.EffectiveRPLStake.SetUint64(0)
	}

	return details, nil
}

// Gets the details for all nodes using the efficient multicall contract
func GetAllNativeNodeDetails(rp *rocketpool.RocketPool, contracts *NetworkContracts, isAtlasDeployed bool) ([]NativeNodeDetails, error) {
	opts := &bind.CallOpts{
		BlockNumber: contracts.ElBlockNumber,
	}

	// Get the list of node addresses
	addresses, err := getNodeAddressesFast(rp, contracts, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}
	count := len(addresses)
	nodeDetails := make([]NativeNodeDetails, count)

	// Sync
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	// Run the getters in batches
	for i := 0; i < count; i += legacyNodeBatchSize {
		i := i
		max := i + legacyNodeBatchSize
		if max > count {
			max = count
		}

		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				address := addresses[j]
				details := &nodeDetails[j]
				details.NodeAddress = address
				details.AverageNodeFee = big.NewInt(0)
				details.DistributorBalanceUserETH = big.NewInt(0)
				details.DistributorBalanceNodeETH = big.NewInt(0)

				if !isAtlasDeployed {
					// Before Atlas, all node's had a 1:1 collateralisation ratio
					details.CollateralisationRatio = eth.EthToWei(2)
				} else {
					details.CollateralisationRatio = big.NewInt(0)
				}

				addNodeDetailsCalls(contracts, mc, details, address, isAtlasDeployed)
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("error getting node details: %w", err)
	}

	// Get the balances of the nodes
	distributorAddresses := make([]common.Address, count)
	balances, err := contracts.BalanceBatcher.GetEthBalances(addresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node balances: %w", err)
	}
	for i, details := range nodeDetails {
		nodeDetails[i].BalanceETH = balances[i]
		distributorAddresses[i] = details.FeeDistributorAddress
	}

	// Get the balances of the distributors
	balances, err = contracts.BalanceBatcher.GetEthBalances(distributorAddresses, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting distributor balances: %w", err)
	}

	// Do some postprocessing on the node data
	for i := range nodeDetails {
		details := &nodeDetails[i]
		details.DistributorBalance = balances[i]

		// Fix the effective stake
		if details.EffectiveRPLStake.Cmp(details.MinimumRPLStake) == -1 {
			details.EffectiveRPLStake.SetUint64(0)
		}
	}

	return nodeDetails, nil
}

// Calculate the average node fee and user/node shares of the distributor's balance
func CalculateAverageFeeAndDistributorShares_Legacy(rp *rocketpool.RocketPool, contracts *NetworkContracts, node NativeNodeDetails, minipoolDetails []*NativeMinipoolDetails) error {

	// Calculate the total of all fees for staking minipools that aren't finalized
	totalFee := big.NewInt(0)
	eligibleMinipools := int64(0)
	for _, mpd := range minipoolDetails {
		if mpd.Status == types.Staking && !mpd.Finalised {
			totalFee.Add(totalFee, mpd.NodeFee)
			eligibleMinipools++
		}
	}

	// Get the average fee (0 if there aren't any minipools)
	if eligibleMinipools > 0 {
		node.AverageNodeFee.Div(totalFee, big.NewInt(eligibleMinipools))
	}

	// Get the user and node portions of the distributor balance
	distributorBalance := big.NewInt(0).Set(node.DistributorBalance)
	if distributorBalance.Cmp(big.NewInt(0)) > 0 {
		halfBalance := big.NewInt(0)
		halfBalance.Div(distributorBalance, two)

		if eligibleMinipools == 0 {
			// Split it 50/50 if there are no minipools
			node.DistributorBalanceNodeETH = big.NewInt(0).Set(halfBalance)
			node.DistributorBalanceUserETH = big.NewInt(0).Sub(distributorBalance, halfBalance)
		} else {
			// Amount of ETH given to the NO as a commission
			commissionEth := big.NewInt(0)
			commissionEth.Mul(halfBalance, node.AverageNodeFee)
			commissionEth.Div(commissionEth, big.NewInt(1e18))

			node.DistributorBalanceNodeETH.Add(halfBalance, commissionEth)                         // Node gets half + commission
			node.DistributorBalanceUserETH.Sub(distributorBalance, node.DistributorBalanceNodeETH) // User gets balance - node share
		}

	} else {
		// No distributor balance
		node.DistributorBalanceNodeETH = big.NewInt(0)
		node.DistributorBalanceUserETH = big.NewInt(0)
	}

	return nil
}

// Calculate the average node fee and user/node shares of the distributor's balance
func CalculateAverageFeeAndDistributorShares_New(rp *rocketpool.RocketPool, contracts *NetworkContracts, node NativeNodeDetails, minipoolDetails []*NativeMinipoolDetails) error {

	// Calculate the total of all fees for staking minipools that aren't finalized
	totalFee := big.NewInt(0)
	eligibleMinipools := int64(0)
	for _, mpd := range minipoolDetails {
		if mpd.Status == types.Staking && !mpd.Finalised {
			totalFee.Add(totalFee, mpd.NodeFee)
			eligibleMinipools++
		}
	}

	// Get the average fee (0 if there aren't any minipools)
	if eligibleMinipools > 0 {
		node.AverageNodeFee.Div(totalFee, big.NewInt(eligibleMinipools))
	}

	// Get the user and node portions of the distributor balance
	distributorBalance := big.NewInt(0).Set(node.DistributorBalance)
	if distributorBalance.Cmp(big.NewInt(0)) > 0 {
		nodeBalance := big.NewInt(0)
		nodeBalance.Mul(distributorBalance, big.NewInt(1e18))
		nodeBalance.Div(nodeBalance, node.CollateralisationRatio)

		userBalance := big.NewInt(0)
		userBalance.Sub(distributorBalance, nodeBalance)

		if eligibleMinipools == 0 {
			// Split it based solely on the collateralisation ratio if there are no minipools (and hence no average fee)
			node.DistributorBalanceNodeETH = big.NewInt(0).Set(nodeBalance)
			node.DistributorBalanceUserETH = big.NewInt(0).Sub(distributorBalance, nodeBalance)
		} else {
			// Amount of ETH given to the NO as a commission
			commissionEth := big.NewInt(0)
			commissionEth.Mul(userBalance, node.AverageNodeFee)
			commissionEth.Div(commissionEth, big.NewInt(1e18))

			node.DistributorBalanceNodeETH.Add(nodeBalance, commissionEth)                         // Node gets their portion + commission on user portion
			node.DistributorBalanceUserETH.Sub(distributorBalance, node.DistributorBalanceNodeETH) // User gets balance - node share
		}

	} else {
		// No distributor balance
		node.DistributorBalanceNodeETH = big.NewInt(0)
		node.DistributorBalanceUserETH = big.NewInt(0)
	}

	return nil
}

// Get all node addresses using the multicaller
func getNodeAddressesFast(rp *rocketpool.RocketPool, contracts *NetworkContracts, opts *bind.CallOpts) ([]common.Address, error) {
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
			mc, err := multicall.NewMultiCaller(rp.Client, contracts.Multicaller.ContractAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				mc.AddCall(contracts.RocketNodeManager, &addresses[j], "getNodeAt", big.NewInt(int64(j)))
			}
			_, err = mc.FlexibleCall(true, opts)
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
func addNodeDetailsCalls(contracts *NetworkContracts, mc *multicall.MultiCaller, details *NativeNodeDetails, address common.Address, isAtlasDeployed bool) {
	mc.AddCall(contracts.RocketNodeManager, &details.Exists, "getNodeExists", address)
	mc.AddCall(contracts.RocketNodeManager, &details.RegistrationTime, "getNodeRegistrationTime", address)
	mc.AddCall(contracts.RocketNodeManager, &details.TimezoneLocation, "getNodeTimezoneLocation", address)
	mc.AddCall(contracts.RocketNodeManager, &details.FeeDistributorInitialised, "getFeeDistributorInitialised", address)
	mc.AddCall(contracts.RocketNodeDistributorFactory, &details.FeeDistributorAddress, "getProxyAddress", address)
	mc.AddCall(contracts.RocketNodeManager, &details.RewardNetwork, "getRewardNetwork", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.RplStake, "getNodeRPLStake", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.EffectiveRPLStake, "getNodeEffectiveRPLStake", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.MinimumRPLStake, "getNodeMinimumRPLStake", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.MaximumRPLStake, "getNodeMaximumRPLStake", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.EthMatched, "getNodeETHMatched", address)
	mc.AddCall(contracts.RocketNodeStaking, &details.EthMatchedLimit, "getNodeETHMatchedLimit", address)
	mc.AddCall(contracts.RocketMinipoolManager, &details.MinipoolCount, "getNodeMinipoolCount", address)
	mc.AddCall(contracts.RocketTokenRETH, &details.BalanceRETH, "balanceOf", address)
	mc.AddCall(contracts.RocketTokenRPL, &details.BalanceRPL, "balanceOf", address)
	mc.AddCall(contracts.RocketTokenRPLFixedSupply, &details.BalanceOldRPL, "balanceOf", address)
	mc.AddCall(contracts.RocketStorage, &details.WithdrawalAddress, "getNodeWithdrawalAddress", address)
	mc.AddCall(contracts.RocketStorage, &details.PendingWithdrawalAddress, "getNodePendingWithdrawalAddress", address)
	mc.AddCall(contracts.RocketNodeManager, &details.SmoothingPoolRegistrationState, "getSmoothingPoolRegistrationState", address)
	mc.AddCall(contracts.RocketNodeManager, &details.SmoothingPoolRegistrationChanged, "getSmoothingPoolRegistrationChanged", address)

	if isAtlasDeployed {
		mc.AddCall(contracts.RocketNodeDeposit, &details.DepositCreditBalance, "getNodeDepositCredit", address)
		mc.AddCall(contracts.RocketNodeStaking, &details.CollateralisationRatio, "getNodeETHCollateralisationRatio", address)
	}
}
