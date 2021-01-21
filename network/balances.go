package network

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/contract"
)


// Get the block number which network balances are current for
func GetBalancesBlock(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketNetworkBalances, err := getRocketNetworkBalances(rp)
    if err != nil {
        return 0, err
    }
    balancesBlock := new(*big.Int)
    if err := rocketNetworkBalances.Call(opts, balancesBlock, "getBalancesBlock"); err != nil {
        return 0, fmt.Errorf("Could not get network balances block: %w", err)
    }
    return (*balancesBlock).Uint64(), nil
}


// Get the current network total ETH balance
func GetTotalETHBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketNetworkBalances, err := getRocketNetworkBalances(rp)
    if err != nil {
        return nil, err
    }
    balance := new(*big.Int)
    if err := rocketNetworkBalances.Call(opts, balance, "getBalance"); err != nil {
        return nil, fmt.Errorf("Could not get withdrawal pool balance: %w", err)
    }
    return *balance, nil
}


// Submit network balances for an epoch
func SubmitBalances(rp *rocketpool.RocketPool, block uint64, totalEth, stakingEth, rethSupply *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkBalances, err := getRocketNetworkBalances(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := contract.Transact(rp.Client, rocketNetworkBalances, opts, "submitBalances", big.NewInt(int64(block)), totalEth, stakingEth, rethSupply)
    if err != nil {
        return nil, fmt.Errorf("Could not submit network balances: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketNetworkBalancesLock sync.Mutex
func getRocketNetworkBalances(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNetworkBalancesLock.Lock()
    defer rocketNetworkBalancesLock.Unlock()
    return rp.GetContract("rocketNetworkBalances")
}

