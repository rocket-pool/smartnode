package deposit

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get the deposit pool balance
func GetBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, opts)
	if err != nil {
		return nil, err
	}
	balance := new(*big.Int)
	if err := rocketDepositPool.Call(opts, balance, "getBalance"); err != nil {
		return nil, fmt.Errorf("Could not get deposit pool balance: %w", err)
	}
	return *balance, nil
}

// Get the excess deposit pool balance
func GetExcessBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, opts)
	if err != nil {
		return nil, err
	}
	excessBalance := new(*big.Int)
	if err := rocketDepositPool.Call(opts, excessBalance, "getExcessBalance"); err != nil {
		return nil, fmt.Errorf("Could not get deposit pool excess balance: %w", err)
	}
	return *excessBalance, nil
}

// Estimate the gas of Deposit
func EstimateDepositGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDepositPool.GetTransactionGasInfo(opts, "deposit")
}

// Make a deposit
func Deposit(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDepositPool.Transact(opts, "deposit")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not deposit: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of AssignDeposits
func EstimateAssignDepositsGas(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketDepositPool.GetTransactionGasInfo(opts, "assignDeposits")
}

// Assign deposits
func AssignDeposits(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
	rocketDepositPool, err := getRocketDepositPool(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketDepositPool.Transact(opts, "assignDeposits")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not assign deposits: %w", err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketDepositPoolLock sync.Mutex

func getRocketDepositPool(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDepositPoolLock.Lock()
	defer rocketDepositPoolLock.Unlock()
	return rp.GetContract("rocketDepositPool", opts)
}
