package minipool

import (
	"fmt"
	"math/big"
	"sync"

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

// Gets the time at which the MP owner started the bond reduction process
func GetReduceBondTime(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (*big.Int, error) {
	rocketMinipoolBondReducer, err := getRocketMinipoolBondReducer(rp, nil)
	if err != nil {
		return nil, err
	}
	reduceBondTime := new(*big.Int)
	if err := rocketMinipoolBondReducer.Call(opts, reduceBondTime, "getReduceBondTime", minipoolAddress); err != nil {
		return nil, fmt.Errorf("Could not get reduce bond start time for minipool %s: %w", minipoolAddress.Hex(), err)
	}
	return *reduceBondTime, nil
}

// Get contracts
var rocketMinipoolBondReducerLock sync.Mutex

func getRocketMinipoolBondReducer(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMinipoolBondReducerLock.Lock()
	defer rocketMinipoolBondReducerLock.Unlock()
	return rp.GetContract("rocketMinipoolBondReducer", opts)
}
