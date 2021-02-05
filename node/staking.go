package node

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Get the total RPL staked in the network
func GetTotalRPLStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketNodeStaking, err := getRocketNodeStaking(rp)
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
    rocketNodeStaking, err := getRocketNodeStaking(rp)
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
    rocketNodeStaking, err := getRocketNodeStaking(rp)
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
    rocketNodeStaking, err := getRocketNodeStaking(rp)
    if err != nil {
        return nil, err
    }
    nodeEffectiveRplStake := new(*big.Int)
    if err := rocketNodeStaking.Call(opts, nodeEffectiveRplStake, "getNodeEffectiveRPLStake", nodeAddress); err != nil {
        return nil, fmt.Errorf("Could not get effective node RPL stake: %w", err)
    }
    return *nodeEffectiveRplStake, nil
}


// Get a node's minimum RPL stake to collateralize their minipools
func GetNodeMinimumRPLStake(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
    rocketNodeStaking, err := getRocketNodeStaking(rp)
    if err != nil {
        return nil, err
    }
    nodeMinimumRplStake := new(*big.Int)
    if err := rocketNodeStaking.Call(opts, nodeMinimumRplStake, "getNodeMinimumRPLStake", nodeAddress); err != nil {
        return nil, fmt.Errorf("Could not get minimum node RPL stake: %w", err)
    }
    return *nodeMinimumRplStake, nil
}


// Get the block a node last staked RPL at
func GetNodeRPLStakedBlock(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (uint64, error) {
    rocketNodeStaking, err := getRocketNodeStaking(rp)
    if err != nil {
        return 0, err
    }
    nodeRplStakedBlock := new(*big.Int)
    if err := rocketNodeStaking.Call(opts, nodeRplStakedBlock, "getNodeRPLStakedBlock", nodeAddress); err != nil {
        return 0, fmt.Errorf("Could not get node RPL staked block: %w", err)
    }
    return (*nodeRplStakedBlock).Uint64(), nil
}


// Get a node's minipool limit based on RPL stake
func GetNodeMinipoolLimit(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (uint64, error) {
    rocketNodeStaking, err := getRocketNodeStaking(rp)
    if err != nil {
        return 0, err
    }
    minipoolLimit := new(*big.Int)
    if err := rocketNodeStaking.Call(opts, minipoolLimit, "getNodeMinipoolLimit", nodeAddress); err != nil {
        return 0, fmt.Errorf("Could not get node minipool limit: %w", err)
    }
    return (*minipoolLimit).Uint64(), nil
}


// Stake RPL
func StakeRPL(rp *rocketpool.RocketPool, rplAmount *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNodeStaking, err := getRocketNodeStaking(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNodeStaking.Transact(opts, "stakeRPL", rplAmount)
    if err != nil {
        return nil, fmt.Errorf("Could not stake RPL: %w", err)
    }
    return txReceipt, nil
}


// Withdraw staked RPL
func WithdrawRPL(rp *rocketpool.RocketPool, rplAmount *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNodeStaking, err := getRocketNodeStaking(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNodeStaking.Transact(opts, "withdrawRPL", rplAmount)
    if err != nil {
        return nil, fmt.Errorf("Could not withdraw staked RPL: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketNodeStakingLock sync.Mutex
func getRocketNodeStaking(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketNodeStakingLock.Lock()
    defer rocketNodeStakingLock.Unlock()
    return rp.GetContract("rocketNodeStaking")
}

