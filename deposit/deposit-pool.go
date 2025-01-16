package deposit

import (
	"fmt"

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
	return rocketDepositPool.GetTransactionGasInfo(opts, "exitQueue", validatorIndex, expressQueue)
}

// Exit the validator queue
func ExitQueue(rp *rocketpool.RocketPool, validatorIndex uint64, expressQueue bool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDepositPool.Transact(opts, "exitQueue", validatorIndex, expressQueue)
	if err != nil {
		return common.Hash{}, fmt.Errorf("error exiting validator queue: %w", err)
	}
	return tx.Hash(), nil
}

func GetQueueTop(rp *rocketpool.RocketPool, opts *bind.CallOpts) (common.Address, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	queueTop := new(common.Address)
	if err := rocketDepositPool.Call(opts, queueTop, "getQueueTop"); err != nil {
		return common.Address{}, fmt.Errorf("error getting queue top: %w", err)
	}
	return *queueTop, nil
}
