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
	NodeSettingsContractName           string = "rocketDAOProtocolSettingsNode"
	MinimumPerMinipoolStakeSettingPath string = "node.per.minipool.stake.minimum"
	MaximumPerMinipoolStakeSettingPath string = "node.per.minipool.stake.maximum"
)

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
