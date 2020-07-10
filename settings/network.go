package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


func GetSubmitBalancesEnabled(rp *rocketpool.RocketPool) (bool, error) {

}


func GetProcessWithdrawalsEnabled(rp *rocketpool.RocketPool) (bool, error) {

}


func GetMinimumNodeFee(rp *rocketpool.RocketPool) (float64, error) {

}


func GetTargetNodeFee(rp *rocketpool.RocketPool) (float64, error) {

}


func GetMaximumNodeFee(rp *rocketpool.RocketPool) (float64, error) {

}


// Get contracts
var rocketNetworkSettingsLock sync.Mutex
func getRocketNetworkSettings(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNetworkSettingsLock.Lock()
    defer rocketNetworkSettingsLock.Unlock()
    return rp.GetContract("rocketNetworkSettings")
}

