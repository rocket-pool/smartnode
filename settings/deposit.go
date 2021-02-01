package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Deposits currently enabled
func GetDepositEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return false, err
    }
    value := new(bool)
    if err := rocketDepositSettings.Call(opts, value, "getDepositEnabled"); err != nil {
        return false, fmt.Errorf("Could not get deposits enabled status: %w", err)
    }
    return *value, nil
}
func SetDepositEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDepositSettings.Transact(opts, "setDepositEnabled", value)
    if err != nil {
        return nil, fmt.Errorf("Could not set deposits enabled status: %w", err)
    }
    return txReceipt, nil
}


// Deposit assignments currently enabled
func GetAssignDepositsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return false, err
    }
    value := new(bool)
    if err := rocketDepositSettings.Call(opts, value, "getAssignDepositsEnabled"); err != nil {
        return false, fmt.Errorf("Could not get deposit assignments enabled status: %w", err)
    }
    return *value, nil
}
func SetAssignDepositsEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDepositSettings.Transact(opts, "setAssignDepositsEnabled", value)
    if err != nil {
        return nil, fmt.Errorf("Could not set deposit assignments enabled status: %w", err)
    }
    return txReceipt, nil
}


// Minimum deposit amount
func GetMinimumDeposit(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := rocketDepositSettings.Call(opts, value, "getMinimumDeposit"); err != nil {
        return nil, fmt.Errorf("Could not get minimum deposit amount: %w", err)
    }
    return *value, nil
}
func SetMinimumDeposit(rp *rocketpool.RocketPool, value *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDepositSettings.Transact(opts, "setMinimumDeposit", value)
    if err != nil {
        return nil, fmt.Errorf("Could not set minimum deposit amount: %w", err)
    }
    return txReceipt, nil
}


// Maximum deposit pool size
func GetMaximumDepositPoolSize(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := rocketDepositSettings.Call(opts, value, "getMaximumDepositPoolSize"); err != nil {
        return nil, fmt.Errorf("Could not get maximum deposit pool size: %w", err)
    }
    return *value, nil
}
func SetMaximumDepositPoolSize(rp *rocketpool.RocketPool, value *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDepositSettings.Transact(opts, "setMaximumDepositPoolSize", value)
    if err != nil {
        return nil, fmt.Errorf("Could not set maximum deposit pool size: %w", err)
    }
    return txReceipt, nil
}


// Maximum deposit assignments per transaction
func GetMaximumDepositAssignments(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rocketDepositSettings.Call(opts, value, "getMaximumDepositAssignments"); err != nil {
        return 0, fmt.Errorf("Could not get maximum deposit assignments: %w", err)
    }
    return (*value).Uint64(), nil
}
func SetMaximumDepositAssignments(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDepositSettings.Transact(opts, "setMaximumDepositAssignments", big.NewInt(int64(value)))
    if err != nil {
        return nil, fmt.Errorf("Could not set maximum deposit assignments: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketDepositSettingsLock sync.Mutex
func getRocketDepositSettings(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketDepositSettingsLock.Lock()
    defer rocketDepositSettingsLock.Unlock()
    return rp.GetContract("rocketDepositSettings")
}

