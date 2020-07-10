package settings

import (
    "fmt"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


func GetNodeRegistrationEnabled(rp *rocketpool.RocketPool) (bool, error) {

}


func GetNodeDepositEnabled(rp *rocketpool.RocketPool) (bool, error) {

}


// Get contracts
var rocketNodeSettingsLock sync.Mutex
func getRocketNodeSettings(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNodeSettingsLock.Lock()
    defer rocketNodeSettingsLock.Unlock()
    return rp.GetContract("rocketNodeSettings")
}

