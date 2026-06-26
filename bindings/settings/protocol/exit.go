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
	ExitSettingsContractName        string = "rocketDAOProtocolSettingsExit"
	CooperativeExitPhaseSettingPath string = "cooperative.exit.phase"
	DidNotExitPenaltySettingPath    string = "did.not.exit.penalty"
	DidNotExitCooldownSettingPath   string = "did.not.exit.cooldown"
)

// Minimum time a validator must remain exit-requested before triggered exit or penalty (hours)
func GetCooperativeExitPhase(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	exitSettingsContract, err := getExitSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := exitSettingsContract.Call(opts, value, "getCooperativeExitPhase"); err != nil {
		return 0, fmt.Errorf("error getting cooperative exit phase: %w", err)
	}
	return time.Duration((*value).Int64()) * time.Hour, nil
}
func ProposeCooperativeExitPhase(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", CooperativeExitPhaseSettingPath), ExitSettingsContractName, CooperativeExitPhaseSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeCooperativeExitPhaseGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", CooperativeExitPhaseSettingPath), ExitSettingsContractName, CooperativeExitPhaseSettingPath, value, blockNumber, treeNodes, opts)
}

// Penalty applied to a minipool that fails to exit when requested
func GetDidNotExitPenalty(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	exitSettingsContract, err := getExitSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := exitSettingsContract.Call(opts, value, "getDidNotExitPenalty"); err != nil {
		return nil, fmt.Errorf("error getting did not exit penalty: %w", err)
	}
	return *value, nil
}
func ProposeDidNotExitPenalty(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", DidNotExitPenaltySettingPath), ExitSettingsContractName, DidNotExitPenaltySettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeDidNotExitPenaltyGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", DidNotExitPenaltySettingPath), ExitSettingsContractName, DidNotExitPenaltySettingPath, value, blockNumber, treeNodes, opts)
}

// Minimum time before a validator can be exit-requested again after a failed exit penalty (days)
func GetDidNotExitCooldown(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	exitSettingsContract, err := getExitSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := exitSettingsContract.Call(opts, value, "getDidNotExitCooldown"); err != nil {
		return 0, fmt.Errorf("error getting did not exit cooldown: %w", err)
	}
	return time.Duration((*value).Int64()) * 24 * time.Hour, nil
}
func ProposeDidNotExitCooldown(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", DidNotExitCooldownSettingPath), ExitSettingsContractName, DidNotExitCooldownSettingPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeDidNotExitCooldownGas(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", DidNotExitCooldownSettingPath), ExitSettingsContractName, DidNotExitCooldownSettingPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var exitSettingsContractLock sync.Mutex

func getExitSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	exitSettingsContractLock.Lock()
	defer exitSettingsContractLock.Unlock()
	return rp.GetContract(ExitSettingsContractName, opts)
}
