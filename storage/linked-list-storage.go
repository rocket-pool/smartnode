package storage

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

type DepositQueueValue struct {
	Receiver       common.Address
	ValidatorID    uint32
	SuppliedValue  uint32
	RequestedValue uint32
}

// Return the number of items in queue
func GetListLength(rp *rocketpool.RocketPool, namespace [32]byte, opts *bind.CallOpts) (*big.Int, error) {
	linkedListStorage, err := getLinkedListStorage(rp, opts)
	if err != nil {
		return nil, err
	}
	length := new(*big.Int)
	if err := linkedListStorage.Call(opts, length, "getLength", namespace); err != nil {
		return nil, fmt.Errorf("error getting address queue length for namespace %s: %w", namespace, err)
	}
	return *length, nil
}

// Return the item in queue by index
func GetListItem(rp *rocketpool.RocketPool, namespace [32]byte, index *big.Int, opts *bind.CallOpts) (DepositQueueValue, error) {
	linkedListStorage, err := getLinkedListStorage(rp, opts)
	if err != nil {
		return DepositQueueValue{}, err
	}
	item := DepositQueueValue{}
	if err := linkedListStorage.Call(opts, item, "getItem", namespace, index); err != nil {
		return DepositQueueValue{}, fmt.Errorf("error getting item at index %s for namespace %s: %w", index, namespace, err)
	}
	return item, nil
}

// Returns the item from the start of the queue without removing it
func PeekListItem(rp *rocketpool.RocketPool, namespace [32]byte, opts *bind.CallOpts) (DepositQueueValue, error) {
	linkedListStorage, err := getLinkedListStorage(rp, opts)
	if err != nil {
		return DepositQueueValue{}, err
	}
	item := DepositQueueValue{}
	if err := linkedListStorage.Call(opts, item, "peekItem", namespace); err != nil {
		return DepositQueueValue{}, fmt.Errorf("error getting peeking the item for namespace %s: %w", namespace, err)
	}
	return item, nil
}

// Returns the index of an item in queue. Returns 0 if the value is not found
func GetListQueueIndexOf(rp *rocketpool.RocketPool, namespace [32]byte, value DepositQueueValue, opts *bind.CallOpts) (*big.Int, error) {
	linkedListStorage, err := getLinkedListStorage(rp, opts)
	if err != nil {
		return nil, err
	}
	queueIndex := new(*big.Int)
	if err := linkedListStorage.Call(opts, queueIndex, "getIndexOf", namespace, value); err != nil {
		return nil, fmt.Errorf("error getting linked list queue for namespace %s: %w", namespace, err)
	}
	return *queueIndex, nil
}

// Get contracts
var LinkedListStorageLock sync.Mutex

func getLinkedListStorage(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	LinkedListStorageLock.Lock()
	defer LinkedListStorageLock.Unlock()
	return rp.GetContract("linkedListStorage", opts)
}
