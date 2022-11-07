package minipool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/storage"
	rptypes "github.com/rocket-pool/rocketpool-go/types"
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

// Minipools queue status details
type QueueDetails struct {
	Position uint64
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
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp, opts)
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
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp, opts)
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

// Get the capacity of the next minipool in the queue
func GetQueueNextCapacity(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketMinipoolQueue, err := getRocketMinipoolQueue(rp, opts)
	if err != nil {
		return nil, err
	}
	capacity := new(*big.Int)
	if err := rocketMinipoolQueue.Call(opts, capacity, "getNextCapacity"); err != nil {
		return nil, fmt.Errorf("Could not get minipool queue next item capacity: %w", err)
	}
	return *capacity, nil
}

// Get Queue position details of a minipool
func GetQueueDetails(rp *rocketpool.RocketPool, mp *Minipool, opts *bind.CallOpts) (QueueDetails, error) {
	position, err := GetQueuePositionOfMinipool(mp, opts)
	if err != nil {
		return QueueDetails{}, err
	}

	// Return
	return QueueDetails{
		Position: position,
	}, nil
}

// Get a minipools position in queue (1-indexed). 0 means it is currently not queued.
func GetQueuePositionOfMinipool(mp *Minipool, opts *bind.CallOpts) (uint64, error) {
	depositType, err := mp.GetDepositType(opts)
	if err != nil {
		return 0, fmt.Errorf("Could not get deposit type: %w", err)
	}
	if depositType == rptypes.None {
		return 0, fmt.Errorf("Minipool address %s has no deposit type", mp.Address)
	}

	queryIndex := func(key string) (uint64, error) {
		index, err := storage.GetAddressQueueIndexOf(mp.RocketPool, opts, crypto.Keccak256Hash([]byte(key)), mp.Address)
		if err != nil {
			return 0, fmt.Errorf("Could not get queue index for address %s: %w", mp.Address, err)
		}
		return uint64(index + 1), nil
	}

	position := uint64(0)

	// half cleared first
	if depositType != rptypes.Half {
		position, err = GetQueueLength(mp.RocketPool, rptypes.Half, opts)
		if err != nil {
			return 0, fmt.Errorf("Could not get queue length of type %s: %w", rptypes.MinipoolDepositTypes[rptypes.Empty], err)
		}
	} else {
		return queryIndex("minipools.available.half")
	}

	// full deposits next
	if depositType != rptypes.Full {
		length, err := GetQueueLength(mp.RocketPool, rptypes.Full, opts)
		if err != nil {
			return 0, fmt.Errorf("Could not get queue length of type %s: %w", rptypes.MinipoolDepositTypes[rptypes.Empty], err)
		}
		position += length
	} else {
		index, err := queryIndex("minipools.available.full")
		if err != nil || index == 0 {
			return 0, err
		}
		return position + index, nil
	}

	// must be empty type now
	index, err := queryIndex("minipools.available.empty")
	if err != nil || index == 0 {
		return 0, err
	}
	return position + index, nil
}

// Get the minipool at the specified position in queue (0-indexed).
func GetQueueMinipoolAtPosition(rp *rocketpool.RocketPool, position uint64, opts *bind.CallOpts) (*Minipool, error) {
	totalLength, err := GetQueueTotalLength(rp, opts)
	if err != nil {
		return nil, fmt.Errorf("Could not get total queue length: %w", err)
	}
	if position >= totalLength {
		return nil, fmt.Errorf("Could not get index %d beyond queue length %d", position, totalLength)
	}
	lengths, err := GetQueueLengths(rp, opts)
	if err != nil {
		return nil, fmt.Errorf("Could not get queue lengths: %w", err)
	}

	getMinipool := func(key string) (*Minipool, error) {
		pos := big.NewInt(int64(position))
		address, err := storage.GetAddressQueueItem(rp, opts, crypto.Keccak256Hash([]byte(key)), pos)
		if err != nil {
			return nil, fmt.Errorf("Could not get address in queue at position %d: %w", position, err)
		}
		return NewMinipool(rp, address, opts)
	}

	if position < lengths.HalfDeposit {
		return getMinipool("minipools.available.half")
	}
	position -= lengths.HalfDeposit
	if position < lengths.FullDeposit {
		return getMinipool("minipools.available.full")
	}
	position -= lengths.FullDeposit
	return getMinipool("minipools.available.empty")
}

// Get contracts
var rocketMinipoolQueueLock sync.Mutex

func getRocketMinipoolQueue(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMinipoolQueueLock.Lock()
	defer rocketMinipoolQueueLock.Unlock()
	return rp.GetContract("rocketMinipoolQueue", opts)
}
