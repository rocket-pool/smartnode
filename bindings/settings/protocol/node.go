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
	NodeSettingsContractName                    string = "rocketDAOProtocolSettingsNode"
	NodeRegistrationEnabledSettingPath          string = "node.registration.enabled"
	SmoothingPoolRegistrationEnabledSettingPath string = "node.smoothing.pool.registration.enabled"
	NodeDepositEnabledSettingPath               string = "node.deposit.enabled"
	VacantMinipoolsEnabledSettingPath           string = "node.vacant.minipools.enabled"
	MinimumPerMinipoolStakeSettingPath          string = "node.per.minipool.stake.minimum"
	MaximumPerMinipoolStakeSettingPath          string = "node.per.minipool.stake.maximum"
)

// Node registrations currently enabled
func GetNodeRegistrationEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := nodeSettingsContract.Call(opts, value, "getRegistrationEnabled"); err != nil {
		return false, fmt.Errorf("error getting node registrations enabled status: %w", err)
	}
	return *value, nil
}
func ProposeNodeRegistrationEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", NodeRegistrationEnabledSettingPath), NodeSettingsContractName, NodeRegistrationEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeNodeRegistrationEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", NodeRegistrationEnabledSettingPath), NodeSettingsContractName, NodeRegistrationEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// Smoothing pool joining currently enabled
func GetSmoothingPoolRegistrationEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := nodeSettingsContract.Call(opts, value, "getSmoothingPoolRegistrationEnabled"); err != nil {
		return false, fmt.Errorf("error getting smoothing pool registrations enabled status: %w", err)
	}
	return *value, nil
}
func ProposeSmoothingPoolRegistrationEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", SmoothingPoolRegistrationEnabledSettingPath), NodeSettingsContractName, SmoothingPoolRegistrationEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeSmoothingPoolRegistrationEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", SmoothingPoolRegistrationEnabledSettingPath), NodeSettingsContractName, SmoothingPoolRegistrationEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// Node deposits currently enabled
func GetNodeDepositEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := nodeSettingsContract.Call(opts, value, "getDepositEnabled"); err != nil {
		return false, fmt.Errorf("error getting node deposits enabled status: %w", err)
	}
	return *value, nil
}
func ProposeNodeDepositEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", NodeDepositEnabledSettingPath), NodeSettingsContractName, NodeDepositEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeNodeDepositEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", NodeDepositEnabledSettingPath), NodeSettingsContractName, NodeDepositEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// Vacant minipools currently enabled
func GetVacantMinipoolsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := nodeSettingsContract.Call(opts, value, "getVacantMinipoolsEnabled"); err != nil {
		return false, fmt.Errorf("error getting vacant minipools enabled status: %w", err)
	}
	return *value, nil
}
func ProposeVacantMinipoolsEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", VacantMinipoolsEnabledSettingPath), NodeSettingsContractName, VacantMinipoolsEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeVacantMinipoolsEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", VacantMinipoolsEnabledSettingPath), NodeSettingsContractName, VacantMinipoolsEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// The minimum RPL stake per minipool as a fraction of assigned user ETH
func GetMinimumPerMinipoolStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMinimumPerMinipoolStake"); err != nil {
		return 0, fmt.Errorf("error getting minimum RPL stake per minipool: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeMinimumPerMinipoolStake(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MinimumPerMinipoolStakeSettingPath), NodeSettingsContractName, MinimumPerMinipoolStakeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMinimumPerMinipoolStakeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MinimumPerMinipoolStakeSettingPath), NodeSettingsContractName, MinimumPerMinipoolStakeSettingPath, value, blockNumber, treeNodes, opts)
}

// The minimum RPL stake per minipool as a fraction of assigned user ETH
func GetMinimumPerMinipoolStakeRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMinimumPerMinipoolStake"); err != nil {
		return nil, fmt.Errorf("error getting minimum RPL stake per minipool: %w", err)
	}
	return *value, nil
}

// The maximum RPL stake per minipool as a fraction of assigned user ETH
func GetMaximumPerMinipoolStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMaximumPerMinipoolStake"); err != nil {
		return 0, fmt.Errorf("error getting maximum RPL stake per minipool: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeMaximumPerMinipoolStake(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MaximumPerMinipoolStakeSettingPath), NodeSettingsContractName, MaximumPerMinipoolStakeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMaximumPerMinipoolStakeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MaximumPerMinipoolStakeSettingPath), NodeSettingsContractName, MaximumPerMinipoolStakeSettingPath, value, blockNumber, treeNodes, opts)
}

// The maximum RPL stake per minipool as a fraction of assigned user ETH
func GetMaximumPerMinipoolStakeRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMaximumPerMinipoolStake"); err != nil {
		return nil, fmt.Errorf("error getting maximum RPL stake per minipool: %w", err)
	}
	return *value, nil
}

// Get contracts
var nodeSettingsContractLock sync.Mutex

func getNodeSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	nodeSettingsContractLock.Lock()
	defer nodeSettingsContractLock.Unlock()
	return rp.GetContract(NodeSettingsContractName, opts)
}
