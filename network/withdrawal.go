package network

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    rptypes "github.com/rocket-pool/rocketpool-go/types"
)


// Get the withdrawal pool balance
func GetWithdrawalBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketNetworkWithdrawal, err := getRocketNetworkWithdrawal(rp)
    if err != nil {
        return nil, err
    }
    balance := new(*big.Int)
    if err := rocketNetworkWithdrawal.Call(opts, balance, "getBalance"); err != nil {
        return nil, fmt.Errorf("Could not get withdrawal pool balance: %w", err)
    }
    return *balance, nil
}


// Get the current network validator withdrawal credentials
func GetWithdrawalCredentials(rp *rocketpool.RocketPool, opts *bind.CallOpts) (common.Hash, error) {
    rocketNetworkWithdrawal, err := getRocketNetworkWithdrawal(rp)
    if err != nil {
        return common.Hash{}, err
    }
    withdrawalCredentials := new(common.Hash)
    if err := rocketNetworkWithdrawal.Call(opts, withdrawalCredentials, "getWithdrawalCredentials"); err != nil {
        return common.Hash{}, fmt.Errorf("Could not get network withdrawal credentials: %w", err)
    }
    return *withdrawalCredentials, nil
}


// Set the network validator withdrawal credentials
func SetWithdrawalCredentials(rp *rocketpool.RocketPool, withdrawalCredentials common.Hash, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkWithdrawal, err := getRocketNetworkWithdrawal(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkWithdrawal.Transact(opts, "setWithdrawalCredentials", withdrawalCredentials.Bytes())
    if err != nil {
        return nil, fmt.Errorf("Could not set network withdrawal credentials: %w", err)
    }
    return txReceipt, nil
}


// Transfer a validator balance to the withdrawal contract
func TransferWithdrawal(rp *rocketpool.RocketPool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkWithdrawal, err := getRocketNetworkWithdrawal(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkWithdrawal.Transfer(opts)
    if err != nil {
        return nil, fmt.Errorf("Could not transfer validator balance: %w", err)
    }
    return txReceipt, nil
}


// Process a validator withdrawal from the beacon chain
func ProcessWithdrawal(rp *rocketpool.RocketPool, validatorPubkey rptypes.ValidatorPubkey, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkWithdrawal, err := getRocketNetworkWithdrawal(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkWithdrawal.Transact(opts, "processWithdrawal", validatorPubkey[:])
    if err != nil {
        return nil, fmt.Errorf("Could not process validator %s withdrawal: %w", validatorPubkey.Hex(), err)
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

