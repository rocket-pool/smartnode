package protocol

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/transactions/gaslimit"
	"github.com/rocket-pool/smartnode/bindings/types"
)

// Config
const (
	PerformanceExitsEnabledSettingPath    string = "network.performance.exits.enabled"
	PerformancePeriodSettingPath          string = "network.performance.period"
	ProofBufferSettingPath                string = "network.performance.proof.buffer"
	PerformanceThresholdSettingPath       string = "network.performance.threshold"
	PerformanceChallengePeriodSettingPath string = "network.performance.challenge.period"
	PerformanceChallengeBondSettingPath   string = "network.performance.challenge.bond"
)

// Performance exits currently enabled
func GetPerformanceExitsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := networkSettingsContract.Call(opts, value, "getPerformanceExitsEnabled"); err != nil {
		return false, fmt.Errorf("error getting performance exits enabled status: %w", err)
	}
	return *value, nil
}
func ProposePerformanceExitsEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", PerformanceExitsEnabledSettingPath), NetworkSettingsContractName, PerformanceExitsEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformanceExitsEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", PerformanceExitsEnabledSettingPath), NetworkSettingsContractName, PerformanceExitsEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// Number of epochs over which attestation performance is measured
func GetPerformancePeriod(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getPerformancePeriod"); err != nil {
		return 0, fmt.Errorf("error getting performance period: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposePerformancePeriod(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", PerformancePeriodSettingPath), NetworkSettingsContractName, PerformancePeriodSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformancePeriodGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", PerformancePeriodSettingPath), NetworkSettingsContractName, PerformancePeriodSettingPath, value, blockNumber, treeNodes, opts)
}

// Buffer to detect underperformance and generate proofs before a validator can be challenged (epochs)
func GetProofBuffer(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getPerformanceProofBuffer"); err != nil {
		return 0, fmt.Errorf("error getting proof buffer: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeProofBuffer(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", ProofBufferSettingPath), NetworkSettingsContractName, ProofBufferSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeProofBufferGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ProofBufferSettingPath), NetworkSettingsContractName, ProofBufferSettingPath, value, blockNumber, treeNodes, opts)
}

// Minimum target attestation timeliness percentage required to avoid exit
func GetPerformanceThreshold(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getPerformanceThreshold"); err != nil {
		return nil, fmt.Errorf("error getting performance threshold: %w", err)
	}
	return *value, nil
}
func ProposePerformanceThreshold(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", PerformanceThresholdSettingPath), NetworkSettingsContractName, PerformanceThresholdSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformanceThresholdGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", PerformanceThresholdSettingPath), NetworkSettingsContractName, PerformanceThresholdSettingPath, value, blockNumber, treeNodes, opts)
}

// How long a performance exit challenge remains open (stored on-chain in seconds)
func GetPerformanceChallengePeriod(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getPerformanceChallengePeriod"); err != nil {
		return 0, fmt.Errorf("error getting performance challenge period: %w", err)
	}
	return time.Duration((*value).Int64()) * time.Second, nil
}
func ProposePerformanceChallengePeriod(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", PerformanceChallengePeriodSettingPath), NetworkSettingsContractName, PerformanceChallengePeriodSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformanceChallengePeriodGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", PerformanceChallengePeriodSettingPath), NetworkSettingsContractName, PerformanceChallengePeriodSettingPath, value, blockNumber, treeNodes, opts)
}

// RPL bond required to propose a performance exit
func GetPerformanceChallengeBond(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getPerformanceChallengeBond"); err != nil {
		return nil, fmt.Errorf("error getting performance challenge bond: %w", err)
	}
	return *value, nil
}
func ProposePerformanceChallengeBond(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", PerformanceChallengeBondSettingPath), NetworkSettingsContractName, PerformanceChallengeBondSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformanceChallengeBondGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", PerformanceChallengeBondSettingPath), NetworkSettingsContractName, PerformanceChallengeBondSettingPath, value, blockNumber, treeNodes, opts)
}
