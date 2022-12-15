package minipool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Minipool queue capacity
type QueueCapacity struct {
	Total     *big.Int
	Effective *big.Int
}

// Minipools queue status details
type QueueDetails struct {
	Position int64
}

// Get minipool queue capacity
func GetQueueCapacity(rp *rocketpool.RocketPool, opts *bind.CallOpts) (QueueCapacity, error) {

	// Data
	var wg errgroup.Group
	var total *big.Int
	var effective *big.Int

	// Load data
	wg.Go(func() error {
		var err error
		total, err = GetQueueTotalCapacity(rp, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		effective, err = GetQueueEffectiveCapacity(rp, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return QueueCapacity{}, err
	}

	// Return
	return QueueCapacity{
		Total:     total,
		Effective: effective,
	}, nil

}

// Get the total length of the minipool queue
func GetQueueTotalLength(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp, opts)
	if err != nil {
		return 0, err
	}
	length := new(*big.Int)
	if err := rocketMinipoolQueue.Call(opts, length, "getTotalLength"); err != nil {
		return 0, fmt.Errorf("Could not get total minipool queue length: %w", err)
	}
	return (*length).Uint64(), nil
}

// Get the total capacity of the minipool queue
func GetQueueTotalCapacity(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp, opts)
	if err != nil {
		return nil, err
	}
	capacity := new(*big.Int)
	if err := rocketMinipoolQueue.Call(opts, capacity, "getTotalCapacity"); err != nil {
		return nil, fmt.Errorf("Could not get minipool queue total capacity: %w", err)
	}
	return *capacity, nil
}

// Get the total effective capacity of the minipool queue (used in node demand calculation)
func GetQueueEffectiveCapacity(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp, opts)
	if err != nil {
		return nil, err
	}
	capacity := new(*big.Int)
	if err := rocketMinipoolQueue.Call(opts, capacity, "getEffectiveCapacity"); err != nil {
		return nil, fmt.Errorf("Could not get minipool queue effective capacity: %w", err)
	}
	return *capacity, nil
}

// Get Queue position details of a minipool
func GetQueueDetails(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (QueueDetails, error) {
	position, err := GetQueuePositionOfMinipool(rp, minipoolAddress, opts)
	if err != nil {
		return QueueDetails{}, err
	}

	// Return
	return QueueDetails{
		Position: position,
	}, nil
}

// Get a minipools position in queue (1-indexed). 0 means it is currently not queued.
func GetQueuePositionOfMinipool(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (int64, error) {
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp, opts)
	if err != nil {
		return 0, err
	}
	position := new(*big.Int)
	if err := rocketMinipoolQueue.Call(opts, position, "getMinipoolPosition", minipoolAddress); err != nil {
		return 0, fmt.Errorf("Could not get queue position for minipool %s: %w", minipoolAddress.Hex(), err)
	}
	return (*position).Int64() + 1, nil
}

// Get the minipool at the specified position in queue (0-indexed).
func GetQueueMinipoolAtPosition(rp *rocketpool.RocketPool, position uint64, opts *bind.CallOpts) (common.Address, error) {
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	address := new(common.Address)
	if err := rocketMinipoolQueue.Call(opts, address, "getMinipoolAt", big.NewInt(int64(position))); err != nil {
		return common.Address{}, fmt.Errorf("Could not get minipool at queue position %d: %w", position, err)
	}
	return *address, nil
}

// Get contracts
var rocketMinipoolQueueLock sync.Mutex

func getRocketMinipoolQueue(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMinipoolQueueLock.Lock()
	defer rocketMinipoolQueueLock.Unlock()
	return rp.GetContract("rocketMinipoolQueue", opts)
}
