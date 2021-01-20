package deposit

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/contract"
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
func Deposit(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDepositPool, err := getRocketDepositPool(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := contract.Transact(rp.Client, rocketDepositPool, opts, "deposit")
    if err != nil {
        return nil, fmt.Errorf("Could not deposit: %w", err)
    }
    return txReceipt, nil
}


// Assign deposits
func AssignDeposits(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDepositPool, err := getRocketDepositPool(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := contract.Transact(rp.Client, rocketDepositPool, opts, "assignDeposits")
    if err != nil {
        return nil, fmt.Errorf("Could not assign deposits: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketDepositPoolLock sync.Mutex
func getRocketDepositPool(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketDepositPoolLock.Lock()
    defer rocketDepositPoolLock.Unlock()
    return rp.GetContract("rocketDepositPool")
}

