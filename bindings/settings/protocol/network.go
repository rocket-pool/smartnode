package protocol

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
)

// Config
const (
	NetworkSettingsContractName                        string = "rocketDAOProtocolSettingsNetwork"
	NodeConsensusThresholdSettingPath                  string = "network.consensus.threshold"
	SubmitBalancesEnabledSettingPath                   string = "network.submit.balances.enabled"
	SubmitBalancesFrequencySettingPath                 string = "network.submit.balances.frequency"
	SubmitPricesEnabledSettingPath                     string = "network.submit.prices.enabled"
	SubmitPricesFrequencySettingPath                   string = "network.submit.prices.frequency"
	MinimumNodeFeeSettingPath                          string = "network.node.fee.minimum"
	TargetNodeFeeSettingPath                           string = "network.node.fee.target"
	MaximumNodeFeeSettingPath                          string = "network.node.fee.maximum"
	NodeFeeDemandRangeSettingPath                      string = "network.node.fee.demand.range"
	NodeComissionShareSecurityCouncilAdder             string = "network.node.commission.share.security.council.adder"
	TargetRethCollateralRateSettingPath                string = "network.reth.collateral.target"
	NetworkPenaltyThresholdSettingPath                 string = "network.penalty.threshold"
	NetworkPenaltyPerRateSettingPath                   string = "network.penalty.per.rate"
	SubmitRewardsEnabledSettingPath                    string = "network.submit.rewards.enabled"
	NetworkAllowListedControllersPath                  string = "network.allow.listed.controllers"
	NetworkNodeCommissionSharePath                     string = "network.node.commission.share"
	NetworkNodeCommissionShareSecurityCouncilAdderPath string = "network.node.commission.share.security.council.adder"
	NetworkVoterSharePath                              string = "network.voter.share"
	NetworkPDAOSharePath                               string = "network.pdao.share"
	NetworkMaxNodeShareSecurityCouncilAdderPath        string = "network.max.node.commission.share.council.adder"
	NetworkMaxRethBalanceDeltaPath                     string = "network.max.reth.balance.delta"
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

// The threshold of trusted nodes that must reach consensus on oracle data to commit it
func GetNodeConsensusThresholdRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getNodeConsensusThreshold"); err != nil {
		return nil, fmt.Errorf("error getting trusted node consensus threshold: %w", err)
	}
	return *value, nil
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

// The frequency in seconds at which network balances should be submitted by trusted nodes
func GetSubmitBalancesFrequency(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getSubmitBalancesFrequency"); err != nil {
		return 0, fmt.Errorf("error getting network balance submission frequency: %w", err)
	}
	return time.Duration((*value).Uint64()) * time.Second, nil
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

// The frequency in seconds at which network prices should be submitted by trusted nodes
func GetSubmitPricesFrequency(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getSubmitPricesFrequency"); err != nil {
		return 0, fmt.Errorf("error getting network price submission frequency: %w", err)
	}
	return time.Duration((*value).Uint64()) * time.Second, nil
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

// Minimum node commission rate
func GetMinimumNodeFeeRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getMinimumNodeFee"); err != nil {
		return nil, fmt.Errorf("error getting minimum node fee: %w", err)
	}
	return *value, nil
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

// Target node commission rate
func GetTargetNodeFeeRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getTargetNodeFee"); err != nil {
		return nil, fmt.Errorf("error getting target node fee: %w", err)
	}
	return *value, nil
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

// Maximum node commission rate
func GetMaximumNodeFeeRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getMaximumNodeFee"); err != nil {
		return nil, fmt.Errorf("error getting maximum node fee: %w", err)
	}
	return *value, nil
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

// The target collateralization rate for the rETH contract as a fraction
func GetTargetRethCollateralRateRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getTargetRethCollateralRate"); err != nil {
		return nil, fmt.Errorf("error getting target rETH contract collateralization rate: %w", err)
	}
	return *value, nil
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

// The number of oDAO members that have to vote for a penalty expressed as a percentage
func GetNetworkPenaltyThresholdRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getNodePenaltyThreshold"); err != nil {
		return nil, fmt.Errorf("error getting network penalty threshold: %w", err)
	}
	return *value, nil
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

// The amount a node operator is penalised for each penalty as a percentage
func GetNetworkPenaltyPerRateRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getPerPenaltyRate"); err != nil {
		return nil, fmt.Errorf("error getting network penalty per rate: %w", err)
	}
	return *value, nil
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

// Returns a list of allow listed controller addresses
func GetAllowListedControllers(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]common.Address, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new([]common.Address)
	if err := networkSettingsContract.Call(opts, value, "getAllowListedControllers"); err != nil {
		return nil, fmt.Errorf("error getting network allow listed controllers list: %w", err)
	}
	return *value, nil
}

func ProposeAllowListedControllers(rp *rocketpool.RocketPool, value []common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetAddressList(rp, fmt.Sprintf("set %s", NetworkAllowListedControllersPath), NetworkSettingsContractName, NetworkAllowListedControllersPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeAllowListedControllersGas(rp *rocketpool.RocketPool, value []common.Address, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetAddressListGas(rp, fmt.Sprintf("set %s", NetworkAllowListedControllersPath), NetworkSettingsContractName, NetworkAllowListedControllersPath, value, blockNumber, treeNodes, opts)
}

// Get the network.node.commission.share setting
func GetNodeShare(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeShare := new(*big.Int)
	if err := networkSettingsContract.Call(opts, nodeShare, "getNodeShare"); err != nil {
		return nil, fmt.Errorf("error getting network node commission share %w", err)
	}
	return *nodeShare, nil
}

func ProposeNodeShare(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NetworkNodeCommissionSharePath), NetworkSettingsContractName, NetworkNodeCommissionSharePath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeNodeShareGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NetworkNodeCommissionSharePath), NetworkSettingsContractName, NetworkNodeCommissionSharePath, value, blockNumber, treeNodes, opts)
}

// Get the network.node.commission.share.security.council.adder setting
func GetNodeShareSecurityCouncilAdder(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeShareSecurityCouncilAdder := new(*big.Int)
	if err := networkSettingsContract.Call(opts, nodeShareSecurityCouncilAdder, "getNodeShareSecurityCouncilAdder"); err != nil {
		return nil, fmt.Errorf("error getting network node commission share %w", err)
	}
	return *nodeShareSecurityCouncilAdder, nil
}

func ProposeNodeShareSecurityCouncilAdder(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NetworkNodeCommissionShareSecurityCouncilAdderPath), NetworkSettingsContractName, NetworkNodeCommissionShareSecurityCouncilAdderPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeNodeShareSecurityCouncilAdderGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NetworkNodeCommissionShareSecurityCouncilAdderPath), NetworkSettingsContractName, NetworkNodeCommissionShareSecurityCouncilAdderPath, value, blockNumber, treeNodes, opts)
}

// Get the network.voter.share setting
func GetVoterShare(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	voterShare := new(*big.Int)
	if err := networkSettingsContract.Call(opts, voterShare, "getVoterShare"); err != nil {
		return nil, fmt.Errorf("error getting network node commission share %w", err)
	}
	return *voterShare, nil
}

func ProposeVoterShare(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NetworkVoterSharePath), NetworkSettingsContractName, NetworkVoterSharePath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeVoterShareGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NetworkVoterSharePath), NetworkSettingsContractName, NetworkVoterSharePath, value, blockNumber, treeNodes, opts)
}

// Get the network.pdao.share setting
func GetProtocolDAOShare(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	pdaoShare := new(*big.Int)
	if err := networkSettingsContract.Call(opts, pdaoShare, "getProtocolDAOShare"); err != nil {
		return nil, fmt.Errorf("error getting network pdao commission share %w", err)
	}
	return *pdaoShare, nil
}

func ProposeProtocolDAOShare(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NetworkPDAOSharePath), NetworkSettingsContractName, NetworkPDAOSharePath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeProtocolDAOShare(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NetworkPDAOSharePath), NetworkSettingsContractName, NetworkPDAOSharePath, value, blockNumber, treeNodes, opts)
}

// Get the network.max.node.commission.share.council.adder setting
func GetMaxNodeShareSecurityCouncilAdder(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	maxNodeShareSecurityCouncilAdder := new(*big.Int)
	if err := networkSettingsContract.Call(opts, maxNodeShareSecurityCouncilAdder, "getMaxNodeShareSecurityCouncilAdder"); err != nil {
		return nil, fmt.Errorf("error getting network node commission share %w", err)
	}
	return *maxNodeShareSecurityCouncilAdder, nil
}

func ProposeMaxNodeShareSecurityCouncilAdder(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NetworkMaxNodeShareSecurityCouncilAdderPath), NetworkSettingsContractName, NetworkMaxNodeShareSecurityCouncilAdderPath, value, blockNumber, treeNodes, opts)
}
func EstimateMaxNodeShareSecurityCouncilAdder(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NetworkMaxNodeShareSecurityCouncilAdderPath), NetworkSettingsContractName, NetworkMaxNodeShareSecurityCouncilAdderPath, value, blockNumber, treeNodes, opts)
}

// Get the network.max.reth.balance.delta setting
func GetMaxRethDelta(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	maxRethDelta := new(*big.Int)
	if err := networkSettingsContract.Call(opts, maxRethDelta, "getMaxRethDelta"); err != nil {
		return nil, fmt.Errorf("error getting network node commission share %w", err)
	}
	return *maxRethDelta, nil
}

func ProposeMaxRethDelta(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NetworkMaxRethBalanceDeltaPath), NetworkSettingsContractName, NetworkMaxRethBalanceDeltaPath, value, blockNumber, treeNodes, opts)
}
func EstimateMaxRethDeltaGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NetworkMaxRethBalanceDeltaPath), NetworkSettingsContractName, NetworkMaxRethBalanceDeltaPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var networkSettingsContractLock sync.Mutex

func getNetworkSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	networkSettingsContractLock.Lock()
	defer networkSettingsContractLock.Unlock()
	return rp.GetContract(NetworkSettingsContractName, opts)
}
