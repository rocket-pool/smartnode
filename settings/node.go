package settings

import (
    "fmt"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Node registrations currently enabled
func GetNodeRegistrationEnabled(rp *rocketpool.RocketPool) (bool, error) {
    rocketNodeSettings, err := getRocketNodeSettings(rp)
    if err != nil {
        return false, err
    }
    registrationEnabled := new(bool)
    if err := rocketNodeSettings.Call(nil, registrationEnabled, "getRegistrationEnabled"); err != nil {
        return false, fmt.Errorf("Could not get node registrations enabled status: %w", err)
    }
    return *registrationEnabled, nil
}


// Node deposits currently enabled
func GetNodeDepositEnabled(rp *rocketpool.RocketPool) (bool, error) {
    rocketNodeSettings, err := getRocketNodeSettings(rp)
    if err != nil {
        return false, err
    }
    depositEnabled := new(bool)
    if err := rocketNodeSettings.Call(nil, depositEnabled, "getDepositEnabled"); err != nil {
        return false, fmt.Errorf("Could not get node deposits enabled status: %w", err)
    }
    return *depositEnabled, nil
}


// Get contracts
var rocketNodeSettingsLock sync.Mutex
func getRocketNodeSettings(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNodeSettingsLock.Lock()
    defer rocketNodeSettingsLock.Unlock()
    return rp.GetContract("rocketNodeSettings")
}

