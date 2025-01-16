package deposit

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Estimate the gas required to exit the validator queue
func EstimateExitQueueGas(rp *rocketpool.RocketPool, validatorIndex uint64, expressQueue bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	validatorIndexBig := big.NewInt(int64(validatorIndex))
	return rocketDepositPool.GetTransactionGasInfo(opts, "exitQueue", validatorIndexBig, expressQueue)
}

// Exit the validator queue
func ExitQueue(rp *rocketpool.RocketPool, validatorIndex uint64, expressQueue bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	validatorIndexBig := big.NewInt(int64(validatorIndex))
	tx, err := rocketDepositPool.Transact(opts, "exitQueue", validatorIndexBig, expressQueue)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error exiting validator queue: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas required to assign megapools
func EstimateAssignMegapoolsGas(rp *rocketpool.RocketPool, count uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDepositPool.GetTransactionGasInfo(opts, "assignMegapools", count)
}

// Assign megapools
func AssignMegapools(rp *rocketpool.RocketPool, count uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDepositPool.Transact(opts, "assignMegapools", count)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error assigning megapools: %w", err)
	}
	return tx.Hash(), nil
}

// Struct to hold queue top (address of the validator at the top of the queue and a boolean indicating if the assignment is possible)
type QueueTop struct {
	MegapoolAddress common.Address
	CanAssign       bool
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
