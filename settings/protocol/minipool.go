package protocol

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
)

// Config
const (
	MinipoolSettingsContractName                  string = "rocketDAOProtocolSettingsMinipool"
	MinipoolSubmitWithdrawableEnabledSettingPath  string = "minipool.submit.withdrawable.enabled"
	MinipoolLaunchTimeoutSettingPath              string = "minipool.launch.timeout"
	BondReductionEnabledSettingPath               string = "minipool.bond.reduction.enabled"
	MaximumMinipoolCountSettingPath               string = "minipool.maximum.count"
	MinipoolUserDistributeWindowStartSettingPath  string = "minipool.user.distribute.window.start"
	MinipoolUserDistributeWindowLengthSettingPath string = "minipool.user.distribute.window.length"
)

// Minipool withdrawable event submissions currently enabled
func GetMinipoolSubmitWithdrawableEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := minipoolSettingsContract.Call(opts, value, "getSubmitWithdrawableEnabled"); err != nil {
		return false, fmt.Errorf("error getting minipool withdrawable submissions enabled status: %w", err)
	}
	return *value, nil
}
func ProposeMinipoolSubmitWithdrawableEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", MinipoolSubmitWithdrawableEnabledSettingPath), MinipoolSettingsContractName, MinipoolSubmitWithdrawableEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMinipoolSubmitWithdrawableEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", MinipoolSubmitWithdrawableEnabledSettingPath), MinipoolSettingsContractName, MinipoolSubmitWithdrawableEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// Timeout period in seconds for prelaunch minipools to launch
func GetMinipoolLaunchTimeout(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getLaunchTimeout"); err != nil {
		return 0, fmt.Errorf("error getting minipool launch timeout: %w", err)
	}
	seconds := time.Duration((*value).Int64()) * time.Second
	return seconds, nil
}
func ProposeMinipoolLaunchTimeout(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MinipoolLaunchTimeoutSettingPath), MinipoolSettingsContractName, MinipoolLaunchTimeoutSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMinipoolLaunchTimeoutGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MinipoolLaunchTimeoutSettingPath), MinipoolSettingsContractName, MinipoolLaunchTimeoutSettingPath, value, blockNumber, treeNodes, opts)
}

// Timeout period in seconds for prelaunch minipools to launch
func GetMinipoolLaunchTimeoutRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getLaunchTimeout"); err != nil {
		return nil, fmt.Errorf("error getting minipool launch timeout: %w", err)
	}
	return *value, nil
}

// Minipool bond reductions currently enabled
func GetBondReductionEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := minipoolSettingsContract.Call(opts, value, "getBondReductionEnabled"); err != nil {
		return false, fmt.Errorf("error getting bond reduction enabled status: %w", err)
	}
	return *value, nil
}
func ProposeBondReductionEnabled(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetBool(rp, fmt.Sprintf("set %s", BondReductionEnabledSettingPath), MinipoolSettingsContractName, BondReductionEnabledSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeBondReductionEnabledGas(rp *rocketpool.RocketPool, value bool, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetBoolGas(rp, fmt.Sprintf("set %s", BondReductionEnabledSettingPath), MinipoolSettingsContractName, BondReductionEnabledSettingPath, value, blockNumber, treeNodes, opts)
}

// The maximum number of minipools allowed
func GetMaximumMinipoolCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getMaximumCount"); err != nil {
		return 0, fmt.Errorf("error getting maximum minipool count: %w", err)
	}
	return (*value).Uint64(), nil
}
func ProposeMaximumMinipoolCount(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MaximumMinipoolCountSettingPath), MinipoolSettingsContractName, MaximumMinipoolCountSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMaximumMinipoolCountGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MaximumMinipoolCountSettingPath), MinipoolSettingsContractName, MaximumMinipoolCountSettingPath, value, blockNumber, treeNodes, opts)
}

// The time a user must wait before being able to distribute a minipool
func GetMinipoolUserDistributeWindowStart(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getUserDistributeWindowStart"); err != nil {
		return 0, fmt.Errorf("error getting user distribute window start: %w", err)
	}
	return time.Duration((*value).Uint64()) * time.Second, nil
}
func ProposeMinipoolUserDistributeWindowStart(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MinipoolUserDistributeWindowStartSettingPath), MinipoolSettingsContractName, MinipoolUserDistributeWindowStartSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMinipoolUserDistributeWindowStartGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MinipoolUserDistributeWindowStartSettingPath), MinipoolSettingsContractName, MinipoolUserDistributeWindowStartSettingPath, value, blockNumber, treeNodes, opts)
}

// The time a user has to distribute a minipool after waiting the start length
func GetMinipoolUserDistributeWindowLength(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getUserDistributeWindowLength"); err != nil {
		return 0, fmt.Errorf("error getting user distribute window length: %w", err)
	}
	return time.Duration((*value).Uint64()) * time.Second, nil
}
func ProposeMinipoolUserDistributeWindowLength(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MinipoolUserDistributeWindowLengthSettingPath), MinipoolSettingsContractName, MinipoolUserDistributeWindowLengthSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMinipoolUserDistributeWindowLengthGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MinipoolUserDistributeWindowLengthSettingPath), MinipoolSettingsContractName, MinipoolUserDistributeWindowLengthSettingPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var minipoolSettingsContractLock sync.Mutex

func getMinipoolSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	minipoolSettingsContractLock.Lock()
	defer minipoolSettingsContractLock.Unlock()
	return rp.GetContract(MinipoolSettingsContractName, opts)
}
