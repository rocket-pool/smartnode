package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Deposit assignments currently enabled
func GetAssignDepositsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return false, err
    }
    assignDepositsEnabled := new(bool)
    if err := rocketDepositSettings.Call(opts, assignDepositsEnabled, "getAssignDepositsEnabled"); err != nil {
        return false, fmt.Errorf("Could not get deposit assignments enabled status: %w", err)
    }
    return *assignDepositsEnabled, nil
}


// Maximum deposit assignments per transaction
func GetMaximumDepositAssignments(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketDepositSettings, err := getRocketDepositSettings(rp)
    if err != nil {
        return 0, err
    }
    maximumDepositAssignments := new(*big.Int)
    if err := rocketDepositSettings.Call(opts, maximumDepositAssignments, "getMaximumDepositAssignments"); err != nil {
        return 0, fmt.Errorf("Could not get maximum deposit assignments: %w", err)
    }
    return (*maximumDepositAssignments).Uint64(), nil
}


// Get contracts
var rocketDepositSettingsLock sync.Mutex
func getRocketDepositSettings(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketDepositSettingsLock.Lock()
    defer rocketDepositSettingsLock.Unlock()
    return rp.GetContract("rocketDepositSettings")
}

