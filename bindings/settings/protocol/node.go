package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/dao/protocol"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
)

// Config
const (
	NodeSettingsContractName                    string = "rocketDAOProtocolSettingsNode"
	NodeRegistrationEnabledSettingPath          string = "node.registration.enabled"
	SmoothingPoolRegistrationEnabledSettingPath string = "node.smoothing.pool.registration.enabled"
	NodeDepositEnabledSettingPath               string = "node.deposit.enabled"
	VacantMinipoolsEnabledSettingPath           string = "node.vacant.minipools.enabled"
	MinimumLegacyRplStakePath                   string = "node.minimum.legacy.staked.rpl"
	ReducedBondSettingPath                      string = "reduced.bond"
	NodeUnstakingPeriodSettingPath              string = "node.unstaking.period"
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

// The amount of legacy staked RPL required by a node after unstaking as percentage of their borrowed ETH
func GetMinimumLegacyRPLStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMinimumLegacyRPLStake"); err != nil {
		return 0, fmt.Errorf("error getting minimum legacy RPL stake per node: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeMinimumLecacyRPLStake(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MinimumLegacyRplStakePath), NodeSettingsContractName, MinimumLegacyRplStakePath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMinimumLecacyRPLStakeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MinimumLegacyRplStakePath), NodeSettingsContractName, MinimumLegacyRplStakePath, value, blockNumber, treeNodes, opts)
}

// The amount of legacy staked RPL required by a node after unstaking as percentage of their borrowed ETH
func GetMinimumLegacyRPLStakeRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMinimumLegacyRPLStake"); err != nil {
		return nil, fmt.Errorf("error getting raw minimum legacy RPL stake per node: %w", err)
	}
	return *value, nil
}

// Get the `reduced_bond` variable used in bond requirements calculation as ETH
func GetReducedBond(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getReducedBond"); err != nil {
		return 0, fmt.Errorf("error getting reduced bond variable: %w", err)
	}
	return eth.WeiToEth(*value), nil
}
func ProposeReducedBond(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", ReducedBondSettingPath), NodeSettingsContractName, ReducedBondSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeReducedBond(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", ReducedBondSettingPath), NodeSettingsContractName, ReducedBondSettingPath, value, blockNumber, treeNodes, opts)
}

// Get the `reduced_bond` variable used in bond requirements calculation as Wei
func GetReducedBondRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getReducedBond"); err != nil {
		return nil, fmt.Errorf("error getting reduced bond variable: %w", err)
	}
	return *value, nil
}

// The the period of time a node must wait before withdrawing RPL
func GetNodeUnstakingPeriod(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getUnstakingPeriod"); err != nil {
		return nil, fmt.Errorf("error getting the unstaking period: %w", err)
	}
	return *value, nil
}
func ProposeNodeUnstakingPeriod(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", NodeUnstakingPeriodSettingPath), NodeSettingsContractName, NodeUnstakingPeriodSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeNodeUnstakingPeriod(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", NodeUnstakingPeriodSettingPath), NodeSettingsContractName, NodeUnstakingPeriodSettingPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var nodeSettingsContractLock sync.Mutex

func getNodeSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	nodeSettingsContractLock.Lock()
	defer nodeSettingsContractLock.Unlock()
	return rp.GetContract(NodeSettingsContractName, opts)
}
