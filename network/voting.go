package network

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Check whether or not on-chain voting has been initialized for the given node
func GetVotingInitialized(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (bool, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := rocketNetworkVoting.Call(opts, value, "getVotingInitialised", address); err != nil {
		return false, fmt.Errorf("error getting voting initialized status: %w", err)
	}
	return *value, nil
}

// Estimate the gas of InitializeVoting
func EstimateInitializeVotingGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkVoting.GetTransactionGasInfo(opts, "initialiseVoting")
}

// Initialize on-chain voting for the node
func InitializeVoting(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkVoting.Transact(opts, "initialiseVoting")
	if err != nil {
		return common.Hash{}, fmt.Errorf("error initializing voting: %w", err)
	}
	return tx.Hash(), nil
}

// Get the number of nodes that were present in the network at the provided block
func GetVotingNodeCount(rp *rocketpool.RocketPool, blockNumber uint32, opts *bind.CallOpts) (*big.Int, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketNetworkVoting.Call(opts, value, "getNodeCount", blockNumber); err != nil {
		return nil, fmt.Errorf("error getting node count for block %d: %w", blockNumber, err)
	}
	return *value, nil
}

// Get the voting power of the given node on the provided block
func GetVotingPower(rp *rocketpool.RocketPool, address common.Address, blockNumber uint32, opts *bind.CallOpts) (*big.Int, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := rocketNetworkVoting.Call(opts, value, "getVotingPower", address, blockNumber); err != nil {
		return nil, fmt.Errorf("error getting voting power for node %s on block %d: %w", address.Hex(), blockNumber, err)
	}
	return *value, nil
}

// Get the address that the provided node has delegated voting power to on the given block
func GetVotingDelegate(rp *rocketpool.RocketPool, address common.Address, blockNumber uint32, opts *bind.CallOpts) (common.Address, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return common.Address{}, err
	}
	value := new(common.Address)
	if err := rocketNetworkVoting.Call(opts, value, "getDelegate", address, blockNumber); err != nil {
		return common.Address{}, fmt.Errorf("error getting delegate for node %s on block %d: %w", address.Hex(), blockNumber, err)
	}
	return *value, nil
}

// Get the address that the provided node has currently delegated voting power to
func GetCurrentVotingDelegate(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (common.Address, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return common.Address{}, err
	}
	value := new(common.Address)
	if err := rocketNetworkVoting.Call(opts, value, "getCurrentDelegate", address); err != nil {
		return common.Address{}, fmt.Errorf("error getting current delegate for node %s: %w", address.Hex(), err)
	}
	return *value, nil
}

// Estimate the gas of SetVotingDelegate
func EstimateSetVotingDelegateGas(rp *rocketpool.RocketPool, newDelegate common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkVoting.GetTransactionGasInfo(opts, "setDelegate", newDelegate)
}

// Set the voting delegate for the node
func SetVotingDelegate(rp *rocketpool.RocketPool, newDelegate common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkVoting.Transact(opts, "setDelegate", newDelegate)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error setting voting delegate: %w", err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketNetworkVotingLock sync.Mutex

func getRocketNetworkVoting(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkVotingLock.Lock()
	defer rocketNetworkVotingLock.Unlock()
	return rp.GetContract("rocketNetworkVoting", opts)
}
