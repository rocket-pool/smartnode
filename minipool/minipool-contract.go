package minipool

import (
    "math/big"
    "sync"
    "time"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Contract access locks
var rocketMinipoolLock sync.Mutex


// Minipool contract
type Minipool struct {
    Address common.Address
    Contract *bind.BoundContract
    rp *rocketpool.RocketPool
}


// Create new minipool contract
func NewMinipool(rp *rocketpool.RocketPool, address common.Address) (*Minipool, error) {

    // Get contract
    contract, err := getMinipoolContract(rp, address)
    if err != nil {
        return nil, err
    }

    // Create and return
    return &Minipool{
        Address: address,
        Contract: contract,
        rp: rp,
    }, nil
}


// Get status details
func (mp *Minipool) GetStatus() (MinipoolStatus, error) {
    return 0, nil
}
func (mp *Minipool) GetStatusBlock() (int64, error) {
    return 0, nil
}
func (mp *Minipool) GetStatusTime() (time.Time, error) {
    return time.Time{}, nil
}


// Get deposit type
func (mp *Minipool) GetDepositType() (MinipoolDeposit, error) {
    return 0, nil
}


// Get node details
func (mp *Minipool) GetNodeAddress() (common.Address, error) {
    return common.Address{}, nil
}
func (mp *Minipool) GetNodeFee() (float64, error) {
    return 0, nil
}
func (mp *Minipool) GetNodeDepositBalance() (*big.Int, error) {
    return nil, nil
}
func (mp *Minipool) GetNodeRefundBalance() (*big.Int, error) {
    return nil, nil
}
func (mp *Minipool) GetNodeDepositAssigned() (bool, error) {
    return false, nil
}


// Get user deposit details
func (mp *Minipool) GetUserDepositBalance() (*big.Int, error) {
    return nil, nil
}
func (mp *Minipool) GetUserDepositAssigned() (bool, error) {
    return false, nil
}


// Get staking details
func (mp *Minipool) GetStakingStartBalance() (*big.Int, error) {
    return nil, nil
}
func (mp *Minipool) GetStakingEndBalance() (*big.Int, error) {
    return nil, nil
}
func (mp *Minipool) GetStakingStartBlock() (int64, error) {
    return 0, nil
}
func (mp *Minipool) GetStakingUserStartBlock() (int64, error) {
    return 0, nil
}
func (mp *Minipool) GetStakingEndBlock() (int64, error) {
    return 0, nil
}


// Get a minipool contract
func getMinipoolContract(rp *rocketpool.RocketPool, minipoolAddress common.Address) (*bind.BoundContract, error) {
    rocketMinipoolLock.Lock()
    defer rocketMinipoolLock.Unlock()
    return rp.MakeContract("rocketMinipool", minipoolAddress)
}

