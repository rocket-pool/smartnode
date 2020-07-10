package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


func GetMinipoolFullDepositNodeAmount(rp *rocketpool.RocketPool) (*big.Int, error) {

}


func GetMinipoolHalfDepositNodeAmount(rp *rocketpool.RocketPool) (*big.Int, error) {

}


func GetMinipoolEmptyDepositNodeAmount(rp *rocketpool.RocketPool) (*big.Int, error) {

}


func GetMinipoolSubmitExitedEnabled(rp *rocketpool.RocketPool) (bool, error) {

}


func GetMinipoolSubmitWithdrawableEnabled(rp *rocketpool.RocketPool) (bool, error) {

}


func GetMinipoolLaunchTimeout(rp *rocketpool.RocketPool) (int64, error) {

}


func GetMinipoolWithdrawalDelay(rp *rocketpool.RocketPool) (int64, error) {

}


// Get contracts
var rocketMinipoolSettingsLock sync.Mutex
func getRocketMinipoolSettings(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketMinipoolSettingsLock.Lock()
    defer rocketMinipoolSettingsLock.Unlock()
    return rp.GetContract("rocketMinipoolSettings")
}

