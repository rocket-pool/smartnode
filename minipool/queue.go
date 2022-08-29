package minipool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/rocketpool-go/storage"
)

// Minipool queue lengths
type QueueLengths struct {
	Total        uint64
	FullDeposit  uint64
	HalfDeposit  uint64
	EmptyDeposit uint64
}

// Minipool queue capacity
type QueueCapacity struct {
	Total        *big.Int
	Effective    *big.Int
	NextMinipool *big.Int
}

// Get minipool queue lengths
func GetQueueLengths(rp *rocketpool.RocketPool, opts *bind.CallOpts) (QueueLengths, error) {

	// Data
	var wg errgroup.Group
	var total uint64
	var fullDeposit uint64
	var halfDeposit uint64
	var emptyDeposit uint64

	// Load data
	wg.Go(func() error {
		var err error
		total, err = GetQueueTotalLength(rp, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		fullDeposit, err = GetQueueLength(rp, rptypes.Full, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		halfDeposit, err = GetQueueLength(rp, rptypes.Half, opts)
		return err
	})
	wg.Go(func() error {
		var err error
		emptyDeposit, err = GetQueueLength(rp, rptypes.Empty, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return QueueLengths{}, err
	}

	// Return
	return QueueLengths{
		Total:        total,
		FullDeposit:  fullDeposit,
		HalfDeposit:  halfDeposit,
		EmptyDeposit: emptyDeposit,
	}, nil

}

// Get minipool queue capacity
func GetQueueCapacity(rp *rocketpool.RocketPool, opts *bind.CallOpts) (QueueCapacity, error) {

	// Data
	var wg errgroup.Group
	var total *big.Int
	var effective *big.Int
	var nextMinipool *big.Int

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
	wg.Go(func() error {
		var err error
		nextMinipool, err = GetQueueNextCapacity(rp, opts)
		return err
	})

	// Wait for data
	if err := wg.Wait(); err != nil {
		return QueueCapacity{}, err
	}

	// Return
	return QueueCapacity{
		Total:        total,
		Effective:    effective,
		NextMinipool: nextMinipool,
	}, nil

}

// Get the total length of the minipool queue
func GetQueueTotalLength(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp)
	if err != nil {
		return 0, err
	}
	length := new(*big.Int)
	if err := rocketMinipoolQueue.Call(opts, length, "getTotalLength"); err != nil {
		return 0, fmt.Errorf("Could not get minipool queue total length: %w", err)
	}
	return (*length).Uint64(), nil
}

// Get the length of a single minipool queue
func GetQueueLength(rp *rocketpool.RocketPool, depositType rptypes.MinipoolDeposit, opts *bind.CallOpts) (uint64, error) {
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp)
	if err != nil {
		return 0, err
	}
	length := new(*big.Int)
	if err := rocketMinipoolQueue.Call(opts, length, "getLength", depositType); err != nil {
		return 0, fmt.Errorf("Could not get minipool queue length for deposit type %d: %w", depositType, err)
	}
	return (*length).Uint64(), nil
}

// Get the total capacity of the minipool queue
func GetQueueTotalCapacity(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp)
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
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp)
	if err != nil {
		return nil, err
	}
	capacity := new(*big.Int)
	if err := rocketMinipoolQueue.Call(opts, capacity, "getEffectiveCapacity"); err != nil {
		return nil, fmt.Errorf("Could not get minipool queue effective capacity: %w", err)
	}
	return *capacity, nil
}

// Get the capacity of the next minipool in the queue
func GetQueueNextCapacity(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp)
	if err != nil {
		return nil, err
	}
	capacity := new(*big.Int)
	if err := rocketMinipoolQueue.Call(opts, capacity, "getNextCapacity"); err != nil {
		return nil, fmt.Errorf("Could not get minipool queue next item capacity: %w", err)
	}
	return *capacity, nil
}

// Get a minipools position in queue
func (mp *Minipool) GetMinipoolPositionInQueue(opts *bind.CallOpts) (uint64, error) {

	index, err := storage.GetAddressQueueIndexOf(mp.RocketPool, opts, [32]byte("minipools.available.half"), mp.Address);
	if err != nil {
		return 0, err
	}
	return index, nil
}

// Get contracts
var rocketMinipoolQueueLock sync.Mutex

func getRocketMinipoolQueue(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
	rocketMinipoolQueueLock.Lock()
	defer rocketMinipoolQueueLock.Unlock()
	return rp.GetContract("rocketMinipoolQueue")
}
