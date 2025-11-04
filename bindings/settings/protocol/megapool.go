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
)

// Config
const (
	MegapoolSettingsContractName           string = "rocketDAOProtocolSettingsMegapool"
	MegapoolTimeBeforeDissolveSettingsPath string = "megapool.time.before.dissolve"
	MegapoolMaximumMegapoolEthPenaltyPath  string = "maximum.megapool.eth.penalty"
	MegapoolNotifyThresholdPath            string = "notify.threshold"
	MegapoolLateNotifyFinePath             string = "late.notify.fine"
	MegapoolDissolvePenaltyPath            string = "megapool.dissolve.penalty"
)

// How long after an assignment a watcher must wait to dissolve a megapool validator
func GetMegapoolTimeBeforeDissolve(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	megapoolSettingsContract, err := getMegapoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := megapoolSettingsContract.Call(opts, value, "getTimeBeforeDissolve"); err != nil {
		return 0, fmt.Errorf("error getting megapool time before dissolve value: %w", err)
	}
	return (*value).Uint64(), nil
}

func ProposeMegapoolTimeBeforeDissolve(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MegapoolTimeBeforeDissolveSettingsPath), MegapoolSettingsContractName, MegapoolTimeBeforeDissolveSettingsPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMegapoolTimeBeforeDissolve(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MegapoolTimeBeforeDissolveSettingsPath), MegapoolSettingsContractName, MegapoolTimeBeforeDissolveSettingsPath, value, blockNumber, treeNodes, opts)
}

// The maximum amount a megapool can be penalised in 50,400 consecutive slots (~7 days)
func GetMaximumEthPenalty(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	megapoolSettingsContract, err := getMegapoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := megapoolSettingsContract.Call(opts, value, "getMaximumEthPenalty"); err != nil {
		return nil, fmt.Errorf("error getting megapool eth penalty value: %w", err)
	}
	return *value, nil
}

func ProposeMaximumEthPenalty(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MegapoolMaximumMegapoolEthPenaltyPath), MegapoolSettingsContractName, MegapoolMaximumMegapoolEthPenaltyPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeMaximumEthPenalty(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MegapoolMaximumMegapoolEthPenaltyPath), MegapoolSettingsContractName, MegapoolMaximumMegapoolEthPenaltyPath, value, blockNumber, treeNodes, opts)
}

// The amount of time before `withdrawable_epoch` a node operator must notify their exit
func GetNotifyThreshold(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	megapoolSettingsContract, err := getMegapoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := megapoolSettingsContract.Call(opts, value, "getNotifyThreshold"); err != nil {
		return 0, fmt.Errorf("error getting megapool notify threshold value: %w", err)
	}
	return (*value).Uint64(), nil
}

func ProposeNotifyThreshold(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MegapoolNotifyThresholdPath), MegapoolSettingsContractName, MegapoolNotifyThresholdPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeNotifyThreshold(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MegapoolNotifyThresholdPath), MegapoolSettingsContractName, MegapoolNotifyThresholdPath, value, blockNumber, treeNodes, opts)
}

// The amount a node operator is fined for notifying their exit late
func GetLateNotifyFine(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	megapoolSettingsContract, err := getMegapoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := megapoolSettingsContract.Call(opts, value, "getLateNotifyFine"); err != nil {
		return nil, fmt.Errorf("error getting megapool late notify fine value: %w", err)
	}
	return *value, nil
}

func ProposeLateNotifyFine(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MegapoolLateNotifyFinePath), MegapoolSettingsContractName, MegapoolLateNotifyFinePath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeLateNotifyFine(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MegapoolLateNotifyFinePath), MegapoolSettingsContractName, MegapoolLateNotifyFinePath, value, blockNumber, treeNodes, opts)
}

// The penalty applied to a NO for having a validator dissolved
func GetMegapoolDissolvePenalty(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	megapoolSettingsContract, err := getMegapoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := megapoolSettingsContract.Call(opts, value, "getDissolvePenalty"); err != nil {
		return nil, fmt.Errorf("error getting megapool dissolve penalty value: %w", err)
	}
	return *value, nil
}

func ProposeDissolvePenalty(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (uint64, common.Hash, error) {
	return protocol.ProposeSetUint(rp, fmt.Sprintf("set %s", MegapoolDissolvePenaltyPath), MegapoolSettingsContractName, MegapoolDissolvePenaltyPath, value, blockNumber, treeNodes, opts)
}
func EstimateProposeDissolvePenalty(rp *rocketpool.RocketPool, value *big.Int, blockNumber uint32, treeNodes []types.VotingTreeNode, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return protocol.EstimateProposeSetUintGas(rp, fmt.Sprintf("set %s", MegapoolDissolvePenaltyPath), MegapoolSettingsContractName, MegapoolDissolvePenaltyPath, value, blockNumber, treeNodes, opts)
}

// Get contracts
var megapoolSettingsContractLock sync.Mutex

func getMegapoolSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	megapoolSettingsContractLock.Lock()
	defer megapoolSettingsContractLock.Unlock()
	return rp.GetContract(MegapoolSettingsContractName, opts)
}
