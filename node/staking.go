package node

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
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
func GetTotalStakedRPL(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	totalRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, totalRplStake, "getTotalStakedRPL"); err != nil {
		return nil, fmt.Errorf("error getting total network RPL stake: %w", err)
	}
	return *totalRplStake, nil
}

// Get the total RPL staked in the network on megapools
func GetTotalMegapoolStakedRPL(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	totalRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, totalRplStake, "getTotalMegapoolStakedRPL"); err != nil {
		return nil, fmt.Errorf("error getting total network megapool RPL stake: %w", err)
	}
	return *totalRplStake, nil
}

// Get the total RPL staked in the network on megapools
func GetTotalLegacyStakedRPL(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	totalRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, totalRplStake, "getTotalLegacyStakedRPL"); err != nil {
		return nil, fmt.Errorf("error getting total network legacy RPL stake: %w", err)
	}
	return *totalRplStake, nil
}

// Get a node's total RPL staked
func GetNodeStakedRPL(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeRplStake, "getNodeStakedRPL", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting total node RPL stake: %w", err)
	}
	return *nodeRplStake, nil
}

// Get a node's megapool RPL staked
func GetNodeMegapoolStakedRPL(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeRplStake, "getNodeMegapoolStakedRPL", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting megapool node RPL stake: %w", err)
	}
	return *nodeRplStake, nil
}

// Get a node's legacy RPL staked
func GetNodeLegacyStakedRPL(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeRplStake, "getNodeLegacyStakedRPL", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting megapool node RPL stake: %w", err)
	}
	return *nodeRplStake, nil
}

// Get the amount of unstaking RPL for a node
func GetNodeUnstakingRPL(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	unstakingRpl := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, unstakingRpl, "getNodeUnstakingRPL", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node unstaking RPL: %w", err)
	}
	return *unstakingRpl, nil
}

// Get a node's maximum RPL stake to collateralize their minipools
func GetNodeMaximumRPLStakeForMinipools(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeMaximumRplStake := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeMaximumRplStake, "getNodeMaximumRPLStakeForMinipools", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting maximum node RPL stake for minipools: %w", err)
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

// Get the time a node last unstaked RPL
func GetNodeLastUnstakeTime(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return 0, err
	}
	nodeRplStakedTime := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeRplStakedTime, "getNodeLastUnstakeTime", nodeAddress); err != nil {
		return 0, fmt.Errorf("error getting node last unstaked RPL time: %w", err)
	}
	return (*nodeRplStakedTime).Uint64(), nil
}

// Get the ratio between capital taken from users and provided by a node operator for minipools
func GetNodeETHCollateralisationRatio(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeEthCol := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeEthCol, "getNodeETHCollateralisationRatio", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting NodeETHCollateralisationRatio: %w", err)
	}
	return *nodeEthCol, nil
}

// Get the amount of ETH the node has borrowed from the deposit pool
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

// Get the amount of ETH the node has borrowed from the deposit pool for its megapool
func GetNodeMegapoolEthMatched(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeEthMatched := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeEthMatched, "getNodeMegapoolETHMatched", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node ETH matched: %w", err)
	}
	return *nodeEthMatched, nil
}

// Get the amount of ETH the node has borrowed from the deposit pool for its minipools
func GetNodeMinipoolEthMatched(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeEthMatched := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeEthMatched, "getNodeMinipoolETHMatched", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node ETH matched: %w", err)
	}
	return *nodeEthMatched, nil
}

// Get the amount of ETH the node has provided
func GetNodeEthProvided(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeEthProvided := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeEthProvided, "getNodeETHProvided", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node ETH matched: %w", err)
	}
	return *nodeEthProvided, nil
}

// Get the amount of ETH the node has provided for its megapool
func GetNodeMegapoolEthProvided(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeEthProvided := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeEthProvided, "getNodeMegapoolETHProvided", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node ETH matched: %w", err)
	}
	return *nodeEthProvided, nil
}

// Get the amount of ETH the node has provided for its minipools
func GetNodeMinipoolEthProvided(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeEthProvided := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, nodeEthProvided, "getNodeMinipoolETHProvided", nodeAddress); err != nil {
		return nil, fmt.Errorf("error getting node ETH matched: %w", err)
	}
	return *nodeEthProvided, nil
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

// Estimate the gas of UnstakeRPL
func EstimateUnstakeGas(rp *rocketpool.RocketPool, rplAmount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeStaking.GetTransactionGasInfo(opts, "unstakeRPL", rplAmount)
}

// Unstake RPL
func UnstakeRPL(rp *rocketpool.RocketPool, rplAmount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeStaking.Transact(opts, "unstakeRPL", rplAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error unstaking RPL: %w", err)
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

// Set stake RPL for allowed for a certain node
func SetNodeStakeRPLForAllowed(rp *rocketpool.RocketPool, nodeAddress common.Address, caller common.Address, allowed bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeStaking.Transact(opts, "setStakeRPLForAllowed", nodeAddress, caller, allowed)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error setting node stake RPL for allowed: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of WithdrawRPL
func EstimateWithdrawRPLGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeStaking.GetTransactionGasInfo(opts, "withdrawRPL")
}

// Withdraw staked RPL
func WithdrawRPL(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeStaking.Transact(opts, "withdrawRPL")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error withdrawing staked RPL: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of WithdrawLegacyRPL
func EstimateWithdrawLegacyRPLGas(rp *rocketpool.RocketPool, rplAmount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNodeStaking.GetTransactionGasInfo(opts, "withdrawLegacyRPL", rplAmount)
}

// Withdraw legacy RPL
func WithdrawLegacyRPL(rp *rocketpool.RocketPool, rplAmount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNodeStaking.Transact(opts, "withdrawLegacyRPL", rplAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error withdrawing legacy RPL: %w", err)
	}
	return tx.Hash(), nil
}

// Get the amount of RPL locked as part of active PDAO proposals or challenges
func GetNodeLockedRPL(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketNodeStaking, err := getRocketNodeStaking(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketNodeStaking.Call(opts, value, "getNodeLockedRPL", nodeAddress); err != nil {
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
