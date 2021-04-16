package deposit

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get the deposit pool balance
func GetBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketDepositPool, err := getRocketDepositPool(rp)
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
    rocketDepositPool, err := getRocketDepositPool(rp)
    if err != nil {
        return nil, err
    }
    excessBalance := new(*big.Int)
    if err := rocketDepositPool.Call(opts, excessBalance, "getExcessBalance"); err != nil {
        return nil, fmt.Errorf("Could not get deposit pool excess balance: %w", err)
    }
    return *excessBalance, nil
}


// Make a deposit
func Deposit(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
    rocketDepositPool, err := getRocketDepositPool(rp)
    if err != nil {
        return common.Hash{}, err
    }
    hash, err := rocketDepositPool.Transact(opts, "deposit")
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not deposit: %w", err)
    }
    return hash, nil
}


// Assign deposits
func AssignDeposits(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (common.Hash, error) {
    rocketDepositPool, err := getRocketDepositPool(rp)
    if err != nil {
        return common.Hash{}, err
    }
    hash, err := rocketDepositPool.Transact(opts, "assignDeposits")
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not assign deposits: %w", err)
    }
    return hash, nil
}


// Wait for a transaction to get mined
func WaitForTransaction(rp *rocketpool.RocketPool, hash common.Hash) (*types.Receipt, error) {
    rocketDepositPool, err := getRocketDepositPool(rp)
    if err != nil {
        return nil, err
    }
    return rocketDepositPool.WaitForTransaction(hash)
}


// Get contracts
var rocketDepositPoolLock sync.Mutex
func getRocketDepositPool(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketDepositPoolLock.Lock()
    defer rocketDepositPoolLock.Unlock()
    return rp.GetContract("rocketDepositPool")
}

