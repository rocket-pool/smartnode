package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Config
const (
	NetworkSettingsContractName         string = "rocketDAOProtocolSettingsNetwork"
	NodeConsensusThresholdSettingPath   string = "network.consensus.threshold"
	SubmitBalancesEnabledSettingPath    string = "network.submit.balances.enabled"
	SubmitBalancesFrequencySettingPath  string = "network.submit.balances.frequency"
	SubmitPricesEnabledSettingPath      string = "network.submit.prices.enabled"
	SubmitPricesFrequencySettingPath    string = "network.submit.prices.frequency"
	MinimumNodeFeeSettingPath           string = "network.node.fee.minimum"
	TargetNodeFeeSettingPath            string = "network.node.fee.target"
	MaximumNodeFeeSettingPath           string = "network.node.fee.maximum"
	NodeFeeDemandRangeSettingPath       string = "network.node.fee.demand.range"
	TargetRethCollateralRateSettingPath string = "network.reth.collateral.target"
	NetworkPenaltyThresholdSettingPath  string = "network.penalty.threshold"
	NetworkPenaltyPerRateSettingPath    string = "network.penalty.per.rate"
	SubmitRewardsEnabledSettingPath     string = "network.submit.rewards.enabled"
)

// The threshold of trusted nodes that must reach consensus on oracle data to commit it
func GetNodeConsensusThreshold(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getNodeConsensusThreshold"); err != nil {
		return 0, fmt.Errorf("error getting trusted node consensus threshold: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeNodeConsensusThreshold(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NodeConsensusThresholdSettingPath), NetworkSettingsContractName, NodeConsensusThresholdSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeNodeConsensusThresholdGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NodeConsensusThresholdSettingPath), NetworkSettingsContractName, NodeConsensusThresholdSettingPath, value, blockNumber, treeNodes, opts)
}

// Network balance submissions currently enabled
func GetSubmitBalancesEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := networkSettingsContract.Call(opts, value, "getSubmitBalancesEnabled"); err != nil {
		return false, fmt.Errorf("error getting network balance submissions enabled status: %w", err)
	}
	return *value, nil
}
func ProposeSubmitBalancesEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", SubmitBalancesEnabledSettingPath), NetworkSettingsContractName, SubmitBalancesEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSubmitBalancesEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", SubmitBalancesEnabledSettingPath), NetworkSettingsContractName, SubmitBalancesEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// The frequency in blocks at which network balances should be submitted by trusted nodes
func GetSubmitBalancesFrequency(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getSubmitBalancesFrequency"); err != nil {
		return 0, fmt.Errorf("error getting network balance submission frequency: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeSubmitBalancesFrequency(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", SubmitBalancesFrequencySettingPath), NetworkSettingsContractName, SubmitBalancesFrequencySettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSubmitBalancesFrequencyGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", SubmitBalancesFrequencySettingPath), NetworkSettingsContractName, SubmitBalancesFrequencySettingPath, value, blockNumber, treeNodes, opts)
}

// Network price submissions currently enabled
func GetSubmitPricesEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := networkSettingsContract.Call(opts, value, "getSubmitPricesEnabled"); err != nil {
		return false, fmt.Errorf("error getting network price submissions enabled status: %w", err)
	}
	return *value, nil
}
func ProposeSubmitPricesEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", SubmitPricesEnabledSettingPath), NetworkSettingsContractName, SubmitPricesEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSubmitPricesEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", SubmitPricesEnabledSettingPath), NetworkSettingsContractName, SubmitPricesEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// The frequency in blocks at which network prices should be submitted by trusted nodes
func GetSubmitPricesFrequency(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getSubmitPricesFrequency"); err != nil {
		return 0, fmt.Errorf("error getting network price submission frequency: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeSubmitPricesFrequency(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", SubmitPricesFrequencySettingPath), NetworkSettingsContractName, SubmitPricesFrequencySettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSubmitPricesFrequencyGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", SubmitPricesFrequencySettingPath), NetworkSettingsContractName, SubmitPricesFrequencySettingPath, value, blockNumber, treeNodes, opts)
}

// Minimum node commission rate
func GetMinimumNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getMinimumNodeFee"); err != nil {
		return 0, fmt.Errorf("error getting minimum node fee: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeMinimumNodeFee(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MinimumNodeFeeSettingPath), NetworkSettingsContractName, MinimumNodeFeeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMinimumNodeFeeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MinimumNodeFeeSettingPath), NetworkSettingsContractName, MinimumNodeFeeSettingPath, value, blockNumber, treeNodes, opts)
}

// Target node commission rate
func GetTargetNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getTargetNodeFee"); err != nil {
		return 0, fmt.Errorf("error getting target node fee: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeTargetNodeFee(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", TargetNodeFeeSettingPath), NetworkSettingsContractName, TargetNodeFeeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeTargetNodeFeeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", TargetNodeFeeSettingPath), NetworkSettingsContractName, TargetNodeFeeSettingPath, value, blockNumber, treeNodes, opts)
}

// Maximum node commission rate
func GetMaximumNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getMaximumNodeFee"); err != nil {
		return 0, fmt.Errorf("error getting maximum node fee: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeMaximumNodeFee(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MaximumNodeFeeSettingPath), NetworkSettingsContractName, MaximumNodeFeeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMaximumNodeFeeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MaximumNodeFeeSettingPath), NetworkSettingsContractName, MaximumNodeFeeSettingPath, value, blockNumber, treeNodes, opts)
}

// The range of node demand values to base fee calculations on
func GetNodeFeeDemandRange(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getNodeFeeDemandRange"); err != nil {
		return nil, fmt.Errorf("error getting node fee demand range: %w", err)
	}
	return *value, nil
}
func ProposeNodeFeeDemandRange(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NodeFeeDemandRangeSettingPath), NetworkSettingsContractName, NodeFeeDemandRangeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeNodeFeeDemandRangeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NodeFeeDemandRangeSettingPath), NetworkSettingsContractName, NodeFeeDemandRangeSettingPath, value, blockNumber, treeNodes, opts)
}

// The target collateralization rate for the rETH contract as a fraction
func GetTargetRethCollateralRate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getTargetRethCollateralRate"); err != nil {
		return 0, fmt.Errorf("error getting target rETH contract collateralization rate: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeTargetRethCollateralRate(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", TargetRethCollateralRateSettingPath), NetworkSettingsContractName, TargetRethCollateralRateSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeTargetRethCollateralRateGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", TargetRethCollateralRateSettingPath), NetworkSettingsContractName, TargetRethCollateralRateSettingPath, value, blockNumber, treeNodes, opts)
}

// The number of oDAO members that have to vote for a penalty expressed as a percentage
func GetNetworkPenaltyThreshold(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getNodePenaltyThreshold"); err != nil {
		return 0, fmt.Errorf("error getting network penalty threshold: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeNetworkPenaltyThreshold(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NetworkPenaltyThresholdSettingPath), NetworkSettingsContractName, NetworkPenaltyThresholdSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeNetworkPenaltyThresholdGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NetworkPenaltyThresholdSettingPath), NetworkSettingsContractName, NetworkPenaltyThresholdSettingPath, value, blockNumber, treeNodes, opts)
}

// The amount a node operator is penalised for each penalty as a percentage
func GetNetworkPenaltyPerRate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getPerPenaltyRate"); err != nil {
		return 0, fmt.Errorf("error getting network penalty per rate: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeNetworkPenaltyPerRate(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NetworkPenaltyPerRateSettingPath), NetworkSettingsContractName, NetworkPenaltyPerRateSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeNetworkPenaltyPerRateGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NetworkPenaltyPerRateSettingPath), NetworkSettingsContractName, NetworkPenaltyPerRateSettingPath, value, blockNumber, treeNodes, opts)
}

// Rewards submissions currently enabled
func GetSubmitRewardsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := networkSettingsContract.Call(opts, value, "getSubmitRewardsEnabled"); err != nil {
		return false, fmt.Errorf("error getting rewards submissions enabled status: %w", err)
	}
	return *value, nil
}
func ProposeSubmitRewardsEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", SubmitRewardsEnabledSettingPath), NetworkSettingsContractName, SubmitRewardsEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSubmitRewardsEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", SubmitRewardsEnabledSettingPath), NetworkSettingsContractName, SubmitRewardsEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var networkSettingsContractLock sync.Mutex

func getNetworkSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	networkSettingsContractLock.Lock()
	defer networkSettingsContractLock.Unlock()
	return rp.GetContract(NetworkSettingsContractName, opts)
}
