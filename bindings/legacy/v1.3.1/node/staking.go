package node

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Get the version of the Node Staking contract
func GetNodeStakingVersion(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint8, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return 0, err
	}
	return rocketpool.GetContractVersion(rp, *rocketNodeStaking.Address, opts)
}

// Get the total RPL staked in the network
func GetTotalRPLStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	totalRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, totalRplStake, "getTotalRPLStake"); err != nil {
		return nil, fmt.Errorf("error getting total network RPL stake: %w", err)
	}
	return *totalRplStake, nil
}

// Get a node's RPL stake
func GetNodeRPLStake(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeRplStake, "getNodeRPLStake", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting total node RPL stake: %w", err)
	}
	return *nodeRplStake, nil
}

// Get a node's effective RPL stake
func GetNodeEffectiveRPLStake(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeEffectiveRplStakeWrapper := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeEffectiveRplStakeWrapper, "getNodeEffectiveRPLStake", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting effective node RPL stake: %w", err)
	}

	minimumStake, err := GetNodeMinimumRPLStake(rp, nodeAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minimum node RPL stake to verify effective stake: %w", err)
	}

	nodeEffectiveRplStake := *nodeEffectiveRplStakeWrapper
	if nodeEffectiveRplStake.Cmp(minimumStake) == -1 {
		// Effective stake should be zero if it's less than the minimum RPL stake
		return big.NewInt(0), nil
	}

	return nodeEffectiveRplStake, nil
}

// Get a node's minimum RPL stake to collateralize their minipools
func GetNodeMinimumRPLStake(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeMinimumRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeMinimumRplStake, "getNodeMinimumRPLStake", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting minimum node RPL stake: %w", err)
	}
	return *nodeMinimumRplStake, nil
}

// Get a node's maximum RPL stake to collateralize their minipools
func GetNodeMaximumRPLStake(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeMaximumRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeMaximumRplStake, "getNodeMaximumRPLStake", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting maximum node RPL stake: %w", err)
	}
	return *nodeMaximumRplStake, nil
}

// Get the time a node last staked RPL
func GetNodeRPLStakedTime(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return 0, err
	}
	nodeRplStakedTime := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeRplStakedTime, "getNodeRPLStakedTime", nodeAddress); err != nil {
		return 0, fmt.Errorf("error getting node RPL staked time: %w", err)
	}
	return (*nodeRplStakedTime).Uint64(), nil
}

// Get the amount of ETH the node has borrowed from the deposit pool to create its minipools
func GetNodeEthMatched(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeEthMatched := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeEthMatched, "getNodeETHMatched", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node ETH matched: %w", err)
	}
	return *nodeEthMatched, nil
}

// Get the amount of ETH the node can borrow from the deposit pool to create its minipools
func GetNodeEthMatchedLimit(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeEthMatchedLimit := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeEthMatchedLimit, "getNodeETHMatchedLimit", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node ETH matched limit: %w", err)
	}
	return *nodeEthMatchedLimit, nil
}

// Estimate the gas of Stake
func EstimateStakeGas(rp *rocketpool.RocketPool, rplAmount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeStaking.GetTransactionGasInfo(opts, "stakeRPL", rplAmount)
}

// Stake RPL
func StakeRPL(rp *rocketpool.RocketPool, rplAmount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeStaking.Transact(opts, "stakeRPL", rplAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error staking RPL: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of Burn RPL
func EstimateBurnRpl(rp *rocketpool.RocketPool, from common.Address, rplAmount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeStaking.GetTransactionGasInfo(opts, "burnRPL", from, rplAmount)
}

// Burn RPL
func BurnRPL(rp *rocketpool.RocketPool, from common.Address, rplAmount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeStaking.Transact(opts, "burnRPL", from, rplAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error burning RPL: %w", err)
	}
	return tx.Hash(), nil

}

// Estimate the gas of set RPL locking allowed
func EstimateSetRPLLockingAllowedGas(rp *rocketpool.RocketPool, caller common.Address, allowed bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeStaking.GetTransactionGasInfo(opts, "setRPLLockingAllowed", caller, allowed)
}

// Set RPL locking allowed
func SetRPLLockingAllowed(rp *rocketpool.RocketPool, caller common.Address, allowed bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeStaking.Transact(opts, "setRPLLockingAllowed", caller, allowed)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error setting RPL locking allowed: %w", err)
	}
	return tx.Hash(), nil
}

// Get RPL locking allowed state for a node
func GetRPLLockedAllowed(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketNodeStaking.Call(opts, value, "getRPLLockingAllowed", nodeAddress); err != nil {
		return false, fmt.Errorf("error getting node RPL locked: %w", err)
	}
	return *value, nil
}

// Estimate the gas of set stake RPL for allowed
func EstimateSetStakeRPLForAllowedGas(rp *rocketpool.RocketPool, caller common.Address, allowed bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeStaking.GetTransactionGasInfo(opts, "setStakeRPLForAllowed", caller, allowed)
}

// Set stake RPL for allowed
func SetStakeRPLForAllowed(rp *rocketpool.RocketPool, caller common.Address, allowed bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeStaking.Transact(opts, "setStakeRPLForAllowed", caller, allowed)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error setting stake RPL for allowed: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of WithdrawRPL
func EstimateWithdrawRPLGas(rp *rocketpool.RocketPool, nodeAddress common.Address, rplAmount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeStaking.GetTransactionGasInfo(opts, "withdrawRPL", nodeAddress, rplAmount)
}

// Withdraw staked RPL
func WithdrawRPL(rp *rocketpool.RocketPool, nodeAddress common.Address, rplAmount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeStaking.Transact(opts, "withdrawRPL", nodeAddress, rplAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error withdrawing staked RPL: %w", err)
	}
	return tx.Hash(), nil
}

// Calculate total effective RPL stake
func CalculateTotalEffectiveRPLStake(rp *rocketpool.RocketPool, offset, limit, rplPrice *big.Int, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	totalEffectiveRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, totalEffectiveRplStake, "calculateTotalEffectiveRPLStake", offset, limit, rplPrice); err != nil {
		return nil, fmt.Errorf("error getting total effective RPL stake: %w", err)
	}
	return *totalEffectiveRplStake, nil
}

// Get the amount of RPL locked as part of active PDAO proposals or challenges
func GetNodeRPLLocked(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, value, "getNodeRPLLocked", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node RPL locked: %w", err)
	}
	return *value, nil
}

// Get contracts
var rocketNodeStakingLock sync.Mutex

func getRocketNodeStaking(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNodeStakingLock.Lock()
	defer rocketNodeStakingLock.Unlock()
	return rp.GetContract("rocketNodeStaking", opts)
}
