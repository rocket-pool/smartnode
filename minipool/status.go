package minipool

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Estimate the gas of SubmitMinipoolWithdrawable
func EstimateSubmitMinipoolWithdrawableGas(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
    rocketMinipoolStatus, err := getRocketMinipoolStatus(rp)
    if err != nil {
        return rocketpool.GasInfo{}, err
    }
    return rocketMinipoolStatus.GetTransactionGasInfo(opts, "submitMinipoolWithdrawable", minipoolAddress)
}


// Submit a minipool withdrawable event
func SubmitMinipoolWithdrawable(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
    rocketMinipoolStatus, err := getRocketMinipoolStatus(rp)
    if err != nil {
        return common.Hash{}, err
    }
    hash, err := rocketMinipoolStatus.Transact(opts, "submitMinipoolWithdrawable", minipoolAddress)
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not submit minipool withdrawable event: %w", err)
    }
    return hash, nil
}


// Get contracts
var rocketMinipoolStatusLock sync.Mutex
func getRocketMinipoolStatus(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketMinipoolStatusLock.Lock()
    defer rocketMinipoolStatusLock.Unlock()
    return rp.GetContract("rocketMinipoolStatus")
}

