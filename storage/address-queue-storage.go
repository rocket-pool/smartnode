package storage

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// low-level address queue storage interface. Currently only used for the minipool queue.

// Return the length of all addresses matching the given key in the queue
func GetAddressQueueLength(rp *rocketpool.RocketPool, opts *bind.CallOpts, key [32]byte) (uint64, error) {
	addressQueueStorage, err := getAddressQueueStorage(rp, opts)
	if err != nil {
		return 0, err
	}
	length := new(*big.Int)
	if err := addressQueueStorage.Call(opts, length, "getIndexOf", key); err != nil {
		return 0, fmt.Errorf("Could not get address queue length for key: %w", key, err)
	}
	return (*length).Uint64(), nil
}

// Return address item at index for the given key
func GetAddressQueueItem(rp *rocketpool.RocketPool, opts *bind.CallOpts, key [32]byte, index *big.Int) (common.Address, error) {
	addressQueueStorage, err := getAddressQueueStorage(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	address := new(common.Address)
	if err := addressQueueStorage.Call(opts, address, "getItem", key, index); err != nil {
		return common.Address{}, fmt.Errorf("Could not get address item at index %d: %w", index, key, err)
	}
	return *address, nil
}

// Return index of the input address for the given key. -1 if not present.
func GetAddressQueueIndexOf(rp *rocketpool.RocketPool, opts *bind.CallOpts, key [32]byte, address common.Address) (int64, error) {
	addressQueueStorage, err := getAddressQueueStorage(rp, opts)
	if err != nil {
		return 0, err
	}
	index := new(*big.Int)
	if err := addressQueueStorage.Call(opts, index, "getIndexOf", key, address); err != nil {
		return 0, fmt.Errorf("Could not get index for address %s: %w", address.String(), err)
	}
	return (*index).Int64(), nil
}

// Get contracts
var AddressQueueStorageLock sync.Mutex

func getAddressQueueStorage(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	AddressQueueStorageLock.Lock()
	defer AddressQueueStorageLock.Unlock()
	return rp.GetContract("addressQueueStorage", opts)
}
