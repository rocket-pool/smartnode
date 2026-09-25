package deposit

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Struct to hold queue top (address of the validator at the top of the queue and a boolean indicating if the assignment is possible)
type QueueTop struct {
	Receiver           common.Address `abi:"receiver"`
	AssignmentPossible bool           `abi:"assignmentPossible"`
	HeadMovedBlock     *big.Int       `abi:"headMovedBlock"`
}

func GetQueueTop(rp *rocketpool.RocketPool, opts *bind.CallOpts) (QueueTop, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, opts)
	if err != nil {
		return QueueTop{}, err
	}
	queueTop := new(QueueTop)
	if err := rocketDepositPool.Call(opts, queueTop, "getQueueTop"); err != nil {
		return QueueTop{}, fmt.Errorf("error getting queue top: %w", err)
	}
	return *queueTop, nil
}

func GetTotalQueueLength(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint32, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, opts)
	if err != nil {
		return 0, err
	}
	totalLength := new(*big.Int)
	if err := rocketDepositPool.Call(opts, totalLength, "getTotalQueueLength"); err != nil {
		return 0, fmt.Errorf("error getting total queue length: %w", err)
	}
	return uint32((*totalLength).Uint64()), nil
}

func GetQueueIndex(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint32, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, opts)
	if err != nil {
		return 0, err
	}
	queueIndexc := new(*big.Int)
	if err := rocketDepositPool.Call(opts, queueIndexc, "getQueueIndex"); err != nil {
		return 0, fmt.Errorf("error getting total queue length: %w", err)
	}
	return uint32((*queueIndexc).Uint64()), nil
}

func GetExpressQueueLength(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint32, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, opts)
	if err != nil {
		return 0, err
	}
	length := new(*big.Int)
	if err := rocketDepositPool.Call(opts, length, "getExpressQueueLength"); err != nil {
		return 0, fmt.Errorf("error getting express queue length: %w", err)
	}
	return uint32((*length).Uint64()), nil
}

func GetStandardQueueLength(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint32, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, opts)
	if err != nil {
		return 0, err
	}
	length := new(*big.Int)
	if err := rocketDepositPool.Call(opts, length, "getStandardQueueLength"); err != nil {
		return 0, fmt.Errorf("error getting standard queue length: %w", err)
	}
	return uint32((*length).Uint64()), nil
}
