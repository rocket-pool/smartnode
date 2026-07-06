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
)

// Config
const (
	PerformanceSettingsContractName       string = "rocketDAOProtocolSettingsPerformance"
	PerformanceExitsEnabledSettingPath    string = "performance.exits.enabled"
	PerformancePeriodSettingPath          string = "performance.period"
	PerformanceProofBufferSettingPath     string = "performance.proof.buffer"
	PerformanceThresholdSettingPath       string = "performance.threshold"
	PerformanceChallengePeriodSettingPath string = "performance.challenge.period"
	PerformanceChallengeBondSettingPath   string = "performance.challenge.bond"
)

// Performance exits currently enabled
func GetPerformanceExitsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	performanceSettingsContract, err := getPerformanceSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := performanceSettingsContract.Call(opts, value, "getPerformanceExitsEnabled"); err != nil {
		return false, fmt.Errorf("error getting performance exits enabled status: %w", err)
	}
	return *value, nil
}
func ProposePerformanceExitsEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", PerformanceExitsEnabledSettingPath), PerformanceSettingsContractName, PerformanceExitsEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformanceExitsEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", PerformanceExitsEnabledSettingPath), PerformanceSettingsContractName, PerformanceExitsEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// Number of epochs over which attestation performance is measured
func GetPerformancePeriod(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	performanceSettingsContract, err := getPerformanceSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := performanceSettingsContract.Call(opts, value, "getPerformancePeriod"); err != nil {
		return 0, fmt.Errorf("error getting performance period: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposePerformancePeriod(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", PerformancePeriodSettingPath), PerformanceSettingsContractName, PerformancePeriodSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformancePeriodGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", PerformancePeriodSettingPath), PerformanceSettingsContractName, PerformancePeriodSettingPath, value, blockNumber, treeNodes, opts)
}

// Time buffer to detect underperformance and generate proofs before a validator can be challenged
func GetPerformanceProofBuffer(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	performanceSettingsContract, err := getPerformanceSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := performanceSettingsContract.Call(opts, value, "getPerformanceProofBuffer"); err != nil {
		return 0, fmt.Errorf("error getting performance proof buffer: %w", err)
	}
	return time.Duration((*value).Int64()) * time.Hour, nil
}
func ProposePerformanceProofBuffer(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", PerformanceProofBufferSettingPath), PerformanceSettingsContractName, PerformanceProofBufferSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformanceProofBufferGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", PerformanceProofBufferSettingPath), PerformanceSettingsContractName, PerformanceProofBufferSettingPath, value, blockNumber, treeNodes, opts)
}

// Minimum target attestation timeliness percentage required to avoid exit
func GetPerformanceThreshold(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	performanceSettingsContract, err := getPerformanceSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := performanceSettingsContract.Call(opts, value, "getPerformanceThreshold"); err != nil {
		return nil, fmt.Errorf("error getting performance threshold: %w", err)
	}
	return *value, nil
}
func ProposePerformanceThreshold(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", PerformanceThresholdSettingPath), PerformanceSettingsContractName, PerformanceThresholdSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformanceThresholdGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", PerformanceThresholdSettingPath), PerformanceSettingsContractName, PerformanceThresholdSettingPath, value, blockNumber, treeNodes, opts)
}

// How long a performance exit challenge remains open
func GetPerformanceChallengePeriod(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	performanceSettingsContract, err := getPerformanceSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := performanceSettingsContract.Call(opts, value, "getPerformanceChallengePeriod"); err != nil {
		return 0, fmt.Errorf("error getting performance challenge period: %w", err)
	}
	return time.Duration((*value).Int64()) * time.Hour, nil
}
func ProposePerformanceChallengePeriod(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", PerformanceChallengePeriodSettingPath), PerformanceSettingsContractName, PerformanceChallengePeriodSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformanceChallengePeriodGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", PerformanceChallengePeriodSettingPath), PerformanceSettingsContractName, PerformanceChallengePeriodSettingPath, value, blockNumber, treeNodes, opts)
}

// RPL bond required to propose a performance exit
func GetPerformanceChallengeBond(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	performanceSettingsContract, err := getPerformanceSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := performanceSettingsContract.Call(opts, value, "getPerformanceChallengeBond"); err != nil {
		return nil, fmt.Errorf("error getting performance challenge bond: %w", err)
	}
	return *value, nil
}
func ProposePerformanceChallengeBond(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", PerformanceChallengeBondSettingPath), PerformanceSettingsContractName, PerformanceChallengeBondSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposePerformanceChallengeBondGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", PerformanceChallengeBondSettingPath), PerformanceSettingsContractName, PerformanceChallengeBondSettingPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var performanceSettingsContractLock sync.Mutex

func getPerformanceSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	performanceSettingsContractLock.Lock()
	defer performanceSettingsContractLock.Unlock()
	return rp.GetContract(PerformanceSettingsContractName, opts)
}
