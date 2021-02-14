package network

import (
    "fmt"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Get the system withdrawal contract address
func GetSystemWithdrawalContractAddress(rp *rocketpool.RocketPool, opts *bind.CallOpts) (common.Address, error) {
    rocketNetworkWithdrawal, err := getRocketNetworkWithdrawal(rp)
    if err != nil {
        return common.Address{}, err
    }
    swcAddress := new(common.Address)
    if err := rocketNetworkWithdrawal.Call(opts, swcAddress, "getSystemWithdrawalContractAddress"); err != nil {
        return common.Address{}, fmt.Errorf("Could not get system withdrawal contract address: %w", err)
    }
    return *swcAddress, nil
}


// Set the system withdrawal contract address
func SetSystemWithdrawalContractAddress(rp *rocketpool.RocketPool, swcAddress common.Address, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkWithdrawal, err := getRocketNetworkWithdrawal(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkWithdrawal.Transact(opts, "setSystemWithdrawalContractAddress", swcAddress)
    if err != nil {
        return nil, fmt.Errorf("Could not set system withdrawal contract address: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketNetworkWithdrawalLock sync.Mutex
func getRocketNetworkWithdrawal(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketNetworkWithdrawalLock.Lock()
    defer rocketNetworkWithdrawalLock.Unlock()
    return rp.GetContract("rocketNetworkWithdrawal")
}

