package minipool

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Estimate the gas required to vote to cancel a minipool's bond reduction
func EstimateVoteCancelReductionGas(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketMinipoolBondReducer, err := getRocketMinipoolBondReducer(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketMinipoolBondReducer.GetTransactionGasInfo(opts, "voteCancelReduction", minipoolAddress)
}

// Vote to cancel a minipool's bond reduction
func VoteCancelReduction(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	rocketMinipoolBondReducer, err := getRocketMinipoolBondReducer(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketMinipoolBondReducer.Transact(opts, "voteCancelReduction", minipoolAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not vote to cancel bond reduction for minipool %s: %w", minipoolAddress.Hex(), err)
	}
	return tx.Hash(), nil
}

// Gets whether or not the bond reduction process for this minipool has already been cancelled
func GetReduceBondCancelled(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketMinipoolBondReducer, err := getRocketMinipoolBondReducer(rp, nil)
	if err != nil {
		return false, err
	}
	isCancelled := new(bool)
	if err := rocketMinipoolBondReducer.Call(opts, isCancelled, "getReduceBondCancelled", minipoolAddress); err != nil {
		return false, fmt.Errorf("Could not get reduce bond cancelled status for minipool %s: %w", minipoolAddress.Hex(), err)
	}
	return *isCancelled, nil
}

// Gets the time at which the MP owner started the bond reduction process
func GetReduceBondTime(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (time.Time, error) {
	rocketMinipoolBondReducer, err := getRocketMinipoolBondReducer(rp, nil)
	if err != nil {
		return time.Time{}, err
	}
	reduceBondTime := new(*big.Int)
	if err := rocketMinipoolBondReducer.Call(opts, reduceBondTime, "getReduceBondTime", minipoolAddress); err != nil {
		return time.Time{}, fmt.Errorf("Could not get reduce bond time for minipool %s: %w", minipoolAddress.Hex(), err)
	}
	return time.Unix((*reduceBondTime).Int64(), 0), nil
}

// Estimate the gas required to begin a minipool bond reduction
func EstimateBeginReduceBondAmountGas(rp *rocketpool.RocketPool, minipoolAddress common.Address, newBondAmount *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketMinipoolBondReducer, err := getRocketMinipoolBondReducer(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketMinipoolBondReducer.GetTransactionGasInfo(opts, "beginReduceBondAmount", minipoolAddress, newBondAmount)
}

// Begin a minipool bond reduction
func BeginReduceBondAmount(rp *rocketpool.RocketPool, minipoolAddress common.Address, newBondAmount *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketMinipoolBondReducer, err := getRocketMinipoolBondReducer(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketMinipoolBondReducer.Transact(opts, "beginReduceBondAmount", minipoolAddress, newBondAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not begin bond reduction for minipool %s: %w", minipoolAddress.Hex(), err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketMinipoolBondReducerLock sync.Mutex

func getRocketMinipoolBondReducer(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMinipoolBondReducerLock.Lock()
	defer rocketMinipoolBondReducerLock.Unlock()
	return rp.GetContract("rocketMinipoolBondReducer", opts)
}
