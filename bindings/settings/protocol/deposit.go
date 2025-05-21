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
)

// Config
const (
	DepositSettingsContractName                    string = "rocketDAOProtocolSettingsDeposit"
	DepositEnabledSettingPath                      string = "deposit.enabled"
	AssignDepositsEnabledSettingPath               string = "deposit.assign.enabled"
	MinimumDepositSettingPath                      string = "deposit.minimum"
	MaximumDepositPoolSizeSettingPath              string = "deposit.pool.maximum"
	MaximumDepositAssignmentsSettingPath           string = "deposit.assign.maximum"
	MaximumSocializedDepositAssignmentsSettingPath string = "deposit.assign.socialised.maximum"
	DepositFeeSettingPath                          string = "deposit.fee"
)

// Deposits currently enabled
func GetDepositEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := depositSettingsContract.Call(opts, value, "getDepositEnabled"); err != nil {
		return false, fmt.Errorf("error getting deposits enabled status: %w", err)
	}
	return *value, nil
}
func ProposeDepositEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", DepositEnabledSettingPath), DepositSettingsContractName, DepositEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeDepositEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", DepositEnabledSettingPath), DepositSettingsContractName, DepositEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// Deposit assignments currently enabled
func GetAssignDepositsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := depositSettingsContract.Call(opts, value, "getAssignDepositsEnabled"); err != nil {
		return false, fmt.Errorf("error getting deposit assignments enabled status: %w", err)
	}
	return *value, nil
}
func ProposeAssignDepositsEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", AssignDepositsEnabledSettingPath), DepositSettingsContractName, AssignDepositsEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeAssignDepositsEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", AssignDepositsEnabledSettingPath), DepositSettingsContractName, AssignDepositsEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// Minimum deposit amount
func GetMinimumDeposit(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := depositSettingsContract.Call(opts, value, "getMinimumDeposit"); err != nil {
		return nil, fmt.Errorf("error getting minimum deposit amount: %w", err)
	}
	return *value, nil
}
func ProposeMinimumDeposit(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MinimumDepositSettingPath), DepositSettingsContractName, MinimumDepositSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMinimumDepositGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MinimumDepositSettingPath), DepositSettingsContractName, MinimumDepositSettingPath, value, blockNumber, treeNodes, opts)
}

// Maximum deposit pool size
func GetMaximumDepositPoolSize(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := depositSettingsContract.Call(opts, value, "getMaximumDepositPoolSize"); err != nil {
		return nil, fmt.Errorf("error getting maximum deposit pool size: %w", err)
	}
	return *value, nil
}
func ProposeMaximumDepositPoolSize(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MaximumDepositPoolSizeSettingPath), DepositSettingsContractName, MaximumDepositPoolSizeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMaximumDepositPoolSizeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MaximumDepositPoolSizeSettingPath), DepositSettingsContractName, MaximumDepositPoolSizeSettingPath, value, blockNumber, treeNodes, opts)
}

// Maximum deposit assignments per transaction
func GetMaximumDepositAssignments(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := depositSettingsContract.Call(opts, value, "getMaximumDepositAssignments"); err != nil {
		return 0, fmt.Errorf("error getting maximum deposit assignments: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeMaximumDepositAssignments(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MaximumDepositAssignmentsSettingPath), DepositSettingsContractName, MaximumDepositAssignmentsSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMaximumDepositAssignmentsGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MaximumDepositAssignmentsSettingPath), DepositSettingsContractName, MaximumDepositAssignmentsSettingPath, value, blockNumber, treeNodes, opts)
}

// Maximum socialized deposit assignments per transaction
func GetMaximumSocializedDepositAssignments(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := depositSettingsContract.Call(opts, value, "getMaximumDepositSocialisedAssignments"); err != nil {
		return 0, fmt.Errorf("error getting maximum socialized deposit assignments: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeMaximumSocializedDepositAssignments(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MaximumSocializedDepositAssignmentsSettingPath), DepositSettingsContractName, MaximumSocializedDepositAssignmentsSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMaximumSocializedDepositAssignmentsGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MaximumSocializedDepositAssignmentsSettingPath), DepositSettingsContractName, MaximumSocializedDepositAssignmentsSettingPath, value, blockNumber, treeNodes, opts)
}

// Current fee taken from user deposits
func GetDepositFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := depositSettingsContract.Call(opts, value, "getDepositFee"); err != nil {
		return nil, fmt.Errorf("error getting deposit fee: %w", err)
	}
	return *value, nil
}
func ProposeDepositFee(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", DepositFeeSettingPath), DepositSettingsContractName, DepositFeeSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeDepositFeeGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", DepositFeeSettingPath), DepositSettingsContractName, DepositFeeSettingPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var depositSettingsContractLock sync.Mutex

func getDepositSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	depositSettingsContractLock.Lock()
	defer depositSettingsContractLock.Unlock()
	return rp.GetContract(DepositSettingsContractName, opts)
}
