package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/dao/protocol"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Config
const MinipoolSettingsContractName = "rocketDAOProtocolSettingsMinipool"


// Get the minipool launch balance
func GetMinipoolLaunchBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    minipoolSettingsContract, err := getMinipoolSettingsContract(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := minipoolSettingsContract.Call(opts, value, "getLaunchBalance"); err != nil {
        return nil, fmt.Errorf("Could not get minipool launch balance: %w", err)
    }
    return *value, nil
}


// Required node deposit amounts
func GetMinipoolFullDepositNodeAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    minipoolSettingsContract, err := getMinipoolSettingsContract(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := minipoolSettingsContract.Call(opts, value, "getFullDepositNodeAmount"); err != nil {
        return nil, fmt.Errorf("Could not get full minipool deposit node amount: %w", err)
    }
    return *value, nil
}
func GetMinipoolHalfDepositNodeAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    minipoolSettingsContract, err := getMinipoolSettingsContract(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := minipoolSettingsContract.Call(opts, value, "getHalfDepositNodeAmount"); err != nil {
        return nil, fmt.Errorf("Could not get half minipool deposit node amount: %w", err)
    }
    return *value, nil
}
func GetMinipoolEmptyDepositNodeAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    minipoolSettingsContract, err := getMinipoolSettingsContract(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := minipoolSettingsContract.Call(opts, value, "getEmptyDepositNodeAmount"); err != nil {
        return nil, fmt.Errorf("Could not get empty minipool deposit node amount: %w", err)
    }
    return *value, nil
}


// Required user deposit amounts
func GetMinipoolFullDepositUserAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    minipoolSettingsContract, err := getMinipoolSettingsContract(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := minipoolSettingsContract.Call(opts, value, "getFullDepositUserAmount"); err != nil {
        return nil, fmt.Errorf("Could not get full minipool deposit user amount: %w", err)
    }
    return *value, nil
}
func GetMinipoolHalfDepositUserAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    minipoolSettingsContract, err := getMinipoolSettingsContract(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := minipoolSettingsContract.Call(opts, value, "getHalfDepositUserAmount"); err != nil {
        return nil, fmt.Errorf("Could not get half minipool deposit user amount: %w", err)
    }
    return *value, nil
}
func GetMinipoolEmptyDepositUserAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    minipoolSettingsContract, err := getMinipoolSettingsContract(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := minipoolSettingsContract.Call(opts, value, "getEmptyDepositUserAmount"); err != nil {
        return nil, fmt.Errorf("Could not get empty minipool deposit user amount: %w", err)
    }
    return *value, nil
}


// Minipool withdrawable event submissions currently enabled
func GetMinipoolSubmitWithdrawableEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    minipoolSettingsContract, err := getMinipoolSettingsContract(rp)
    if err != nil {
        return false, err
    }
    value := new(bool)
    if err := minipoolSettingsContract.Call(opts, value, "getSubmitWithdrawableEnabled"); err != nil {
        return false, fmt.Errorf("Could not get minipool withdrawable submissions enabled status: %w", err)
    }
    return *value, nil
}
func BootstrapMinipoolSubmitWithdrawableEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    return protocol.BootstrapBool(rp, MinipoolSettingsContractName, "minipool.submit.withdrawable.enabled", value, opts)
}


// Timeout period in blocks for prelaunch minipools to launch
func GetMinipoolLaunchTimeout(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    minipoolSettingsContract, err := getMinipoolSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := minipoolSettingsContract.Call(opts, value, "getLaunchTimeout"); err != nil {
        return 0, fmt.Errorf("Could not get minipool launch timeout: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapMinipoolLaunchTimeout(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return protocol.BootstrapUint(rp, MinipoolSettingsContractName, "minipool.launch.timeout", big.NewInt(int64(value)), opts)
}


// Withdrawal delay in blocks before withdrawable minipools can be closed
func GetMinipoolWithdrawalDelay(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    minipoolSettingsContract, err := getMinipoolSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := minipoolSettingsContract.Call(opts, value, "getWithdrawalDelay"); err != nil {
        return 0, fmt.Errorf("Could not get minipool withdrawal delay: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapMinipoolWithdrawalDelay(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    return protocol.BootstrapUint(rp, MinipoolSettingsContractName, "minipool.withdrawal.delay", big.NewInt(int64(value)), opts)
}


// Get contracts
var minipoolSettingsContractLock sync.Mutex
func getMinipoolSettingsContract(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    minipoolSettingsContractLock.Lock()
    defer minipoolSettingsContractLock.Unlock()
    return rp.GetContract(MinipoolSettingsContractName)
}

