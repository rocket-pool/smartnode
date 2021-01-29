package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Required node deposit amounts
func GetMinipoolFullDepositNodeAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketMinipoolSettings, err := getRocketMinipoolSettings(rp)
    if err != nil {
        return nil, err
    }
    fullDepositNodeAmount := new(*big.Int)
    if err := rocketMinipoolSettings.Call(opts, fullDepositNodeAmount, "getFullDepositNodeAmount"); err != nil {
        return nil, fmt.Errorf("Could not get full minipool deposit node amount: %w", err)
    }
    return *fullDepositNodeAmount, nil
}
func GetMinipoolHalfDepositNodeAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketMinipoolSettings, err := getRocketMinipoolSettings(rp)
    if err != nil {
        return nil, err
    }
    halfDepositNodeAmount := new(*big.Int)
    if err := rocketMinipoolSettings.Call(opts, halfDepositNodeAmount, "getHalfDepositNodeAmount"); err != nil {
        return nil, fmt.Errorf("Could not get half minipool deposit node amount: %w", err)
    }
    return *halfDepositNodeAmount, nil
}
func GetMinipoolEmptyDepositNodeAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketMinipoolSettings, err := getRocketMinipoolSettings(rp)
    if err != nil {
        return nil, err
    }
    emptyDepositNodeAmount := new(*big.Int)
    if err := rocketMinipoolSettings.Call(opts, emptyDepositNodeAmount, "getEmptyDepositNodeAmount"); err != nil {
        return nil, fmt.Errorf("Could not get empty minipool deposit node amount: %w", err)
    }
    return *emptyDepositNodeAmount, nil
}


// Minipool withdrawable event submissions currently enabled
func GetMinipoolSubmitWithdrawableEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketMinipoolSettings, err := getRocketMinipoolSettings(rp)
    if err != nil {
        return false, err
    }
    submitWithdrawableEnabled := new(bool)
    if err := rocketMinipoolSettings.Call(opts, submitWithdrawableEnabled, "getSubmitWithdrawableEnabled"); err != nil {
        return false, fmt.Errorf("Could not get minipool withdrawable submissions enabled status: %w", err)
    }
    return *submitWithdrawableEnabled, nil
}


// Timeout period in blocks for prelaunch minipools to launch
func GetMinipoolLaunchTimeout(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketMinipoolSettings, err := getRocketMinipoolSettings(rp)
    if err != nil {
        return 0, err
    }
    launchTimeout := new(*big.Int)
    if err := rocketMinipoolSettings.Call(opts, launchTimeout, "getLaunchTimeout"); err != nil {
        return 0, fmt.Errorf("Could not get minipool launch timeout: %w", err)
    }
    return (*launchTimeout).Uint64(), nil
}


// Withdrawal delay in blocks before withdrawable minipools can be closed
func GetMinipoolWithdrawalDelay(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketMinipoolSettings, err := getRocketMinipoolSettings(rp)
    if err != nil {
        return 0, err
    }
    withdrawalDelay := new(*big.Int)
    if err := rocketMinipoolSettings.Call(opts, withdrawalDelay, "getWithdrawalDelay"); err != nil {
        return 0, fmt.Errorf("Could not get minipool withdrawal delay: %w", err)
    }
    return (*withdrawalDelay).Uint64(), nil
}
func SetMinipoolWithdrawalDelay(rp *rocketpool.RocketPool, withdrawalDelay uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketMinipoolSettings, err := getRocketMinipoolSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketMinipoolSettings.Transact(opts, "setWithdrawalDelay", big.NewInt(int64(withdrawalDelay)))
    if err != nil {
        return nil, fmt.Errorf("Could not set minipool withdrawal delay: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketMinipoolSettingsLock sync.Mutex
func getRocketMinipoolSettings(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketMinipoolSettingsLock.Lock()
    defer rocketMinipoolSettingsLock.Unlock()
    return rp.GetContract("rocketMinipoolSettings")
}

