package network

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/utils/multicall"
	"golang.org/x/sync/errgroup"
)

const (
	nodeVotingDetailsBatchSize uint64 = 250
)

// Gets the voting power and delegation info for every node at the specified block using multicall
func GetNodeInfoSnapshotFast(rp *rocketpool.RocketPool, blockNumber uint32, multicallAddress common.Address, opts *bind.CallOpts) ([]types.NodeVotingInfo, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, opts)
	if err != nil {
		return nil, err
	}

	// Get the number of voting nodes
	nodeCountBig, err := GetVotingNodeCount(rp, blockNumber, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting voting node count: %w", err)
	}
	nodeCount := nodeCountBig.Uint64()

	// Get the node addresses
	nodeAddresses, err := node.GetNodeAddressesFast(rp, multicallAddress, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting node addresses: %w", err)
	}

	// Sync
	var wg errgroup.Group

	// Run the getters in batches
	votingInfos := make([]types.NodeVotingInfo, nodeCount)
	for i := uint64(0); i < nodeCount; i += nodeVotingDetailsBatchSize {
		i := i
		max := i + nodeVotingDetailsBatchSize
		if max > nodeCount {
			max = nodeCount
		}

		// Load details
		wg.Go(func() error {
			var err error
			mc, err := multicall.NewMultiCaller(rp.Client, multicallAddress)
			if err != nil {
				return err
			}
			for j := i; j < max; j++ {
				nodeAddress := nodeAddresses[j]
				votingInfos[j].NodeAddress = nodeAddress
				mc.AddCall(rocketNetworkVoting, &votingInfos[j].VotingPower, "getVotingPower", nodeAddress, blockNumber)
				mc.AddCall(rocketNetworkVoting, &votingInfos[j].Delegate, "getDelegate", nodeAddress, blockNumber)
			}
			_, err = mc.FlexibleCall(true, opts)
			if err != nil {
				return fmt.Errorf("error executing multicall: %w", err)
			}
			return nil
		})
	}

	// Wait for data
	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return votingInfos, nil
}

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

// Estimate the gas of InitializeVotingWithDelegate
func EstimateInitializeVotingWithDelegateGas(rp *rocketpool.RocketPool, delegateAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkVoting.GetTransactionGasInfo(opts, "initialiseVotingWithDelegate", delegateAddress)
}

// Initialize on-chain voting for the node and delegate voting power at the same transaction
func InitializeVotingWithDelegate(rp *rocketpool.RocketPool, delegateAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkVoting, err := getRocketNetworkVoting(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkVoting.Transact(opts, "initialiseVotingWithDelegate", delegateAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error initializing voting with delegate: %w", err)
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
