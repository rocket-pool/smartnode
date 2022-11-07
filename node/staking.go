package node

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get the total RPL staked in the network
func GetTotalRPLStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	totalRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, totalRplStake, "getTotalRPLStake"); err != nil {
		return nil, fmt.Errorf("Could not get total network RPL stake: %w", err)
	}
	return *totalRplStake, nil
}

// Get the effective RPL staked in the network
func GetTotalEffectiveRPLStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	totalEffectiveRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, totalEffectiveRplStake, "getTotalEffectiveRPLStake"); err != nil {
		return nil, fmt.Errorf("Could not get effective network RPL stake: %w", err)
	}
	return *totalEffectiveRplStake, nil
}

// Get a node's RPL stake
func GetNodeRPLStake(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeRplStake, "getNodeRPLStake", nodeAddress); err != nil {
		return nil, fmt.Errorf("Could not get total node RPL stake: %w", err)
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
		return nil, fmt.Errorf("Could not get effective node RPL stake: %w", err)
	}

	minimumStake, err := GetNodeMinimumRPLStake(rp, nodeAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("Could not get minimum node RPL stake to verify effective stake: %w", err)
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
		return nil, fmt.Errorf("Could not get minimum node RPL stake: %w", err)
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
		return nil, fmt.Errorf("Could not get maximum node RPL stake: %w", err)
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
		return 0, fmt.Errorf("Could not get node RPL staked time: %w", err)
	}
	return (*nodeRplStakedTime).Uint64(), nil
}

// Get a node's minipool limit based on RPL stake
func GetNodeMinipoolLimit(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return 0, err
	}
	minipoolLimit := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, minipoolLimit, "getNodeMinipoolLimit", nodeAddress); err != nil {
		return 0, fmt.Errorf("Could not get node minipool limit: %w", err)
	}
	return (*minipoolLimit).Uint64(), nil
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
		return common.Hash{}, fmt.Errorf("Could not stake RPL: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of WithdrawRPL
func EstimateWithdrawRPLGas(rp *rocketpool.RocketPool, rplAmount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeStaking.GetTransactionGasInfo(opts, "withdrawRPL", rplAmount)
}

// Withdraw staked RPL
func WithdrawRPL(rp *rocketpool.RocketPool, rplAmount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeStaking.Transact(opts, "withdrawRPL", rplAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not withdraw staked RPL: %w", err)
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
		return nil, fmt.Errorf("Could not get total effective RPL stake: %w", err)
	}
	return *totalEffectiveRplStake, nil
}

// Get contracts
var rocketNodeStakingLock sync.Mutex

func getRocketNodeStaking(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNodeStakingLock.Lock()
	defer rocketNodeStakingLock.Unlock()
	return rp.GetContract("rocketNodeStaking", opts)
}
