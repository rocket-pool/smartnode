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
	// Exit settings are stored in the network settings contract.
	CooperativeExitPhaseSettingPath  string = "network.cooperative.exit.phase"
	DidNotExitPenaltyBaseSettingPath string = "network.did.not.exit.penalty.base"
	DidNotExitBaseSettingPath        string = "network.did.not.exit.base"
	DidNotExitBackoffSettingPath     string = "network.did.not.exit.backoff"
)

// Minimum time a validator must remain exit-requested before triggered exit or penalty
func GetCooperativeExitPhase(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getCooperativeExitPhase"); err != nil {
		return 0, fmt.Errorf("error getting cooperative exit phase: %w", err)
	}
	return time.Duration((*value).Int64()) * time.Second, nil
}

// Propose the cooperative exit phase in seconds
func ProposeCooperativeExitPhase(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", CooperativeExitPhaseSettingPath), NetworkSettingsContractName, CooperativeExitPhaseSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeCooperativeExitPhaseGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", CooperativeExitPhaseSettingPath), NetworkSettingsContractName, CooperativeExitPhaseSettingPath, value, blockNumber, treeNodes, opts)
}

// Base penalty applied to a minipool that fails to exit when requested
func GetDidNotExitPenaltyBase(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getDidNotExitPenaltyBase"); err != nil {
		return nil, fmt.Errorf("error getting did not exit penalty base: %w", err)
	}
	return *value, nil
}
func ProposeDidNotExitPenaltyBase(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", DidNotExitPenaltyBaseSettingPath), NetworkSettingsContractName, DidNotExitPenaltyBaseSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeDidNotExitPenaltyBaseGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", DidNotExitPenaltyBaseSettingPath), NetworkSettingsContractName, DidNotExitPenaltyBaseSettingPath, value, blockNumber, treeNodes, opts)
}

// Base delay after a failed cooperative exit before the protocol can try again
func GetDidNotExitBase(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getDidNotExitBase"); err != nil {
		return 0, fmt.Errorf("error getting did not exit base: %w", err)
	}
	return time.Duration((*value).Int64()) * time.Second, nil
}

// Propose the base delay in seconds
func ProposeDidNotExitBase(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", DidNotExitBaseSettingPath), NetworkSettingsContractName, DidNotExitBaseSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeDidNotExitBaseGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", DidNotExitBaseSettingPath), NetworkSettingsContractName, DidNotExitBaseSettingPath, value, blockNumber, treeNodes, opts)
}

// Multiplier applied to the penalty window on each iteration (18-decimal fixed point):
// the i-th window is did_not_exit_base * did_not_exit_backoff ** i
func GetDidNotExitBackoff(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getDidNotExitBackoff"); err != nil {
		return nil, fmt.Errorf("error getting did not exit backoff: %w", err)
	}
	return *value, nil
}
func ProposeDidNotExitBackoff(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", DidNotExitBackoffSettingPath), NetworkSettingsContractName, DidNotExitBackoffSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeDidNotExitBackoffGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (gaslimit.Limits, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", DidNotExitBackoffSettingPath), NetworkSettingsContractName, DidNotExitBackoffSettingPath, value, blockNumber, treeNodes, opts)
}
