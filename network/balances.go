package network

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/contract"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
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
    totalEthBalance := new(*big.Int)
    if err := rocketNetworkBalances.Call(opts, totalEthBalance, "getTotalETHBalance"); err != nil {
        return nil, fmt.Errorf("Could not get network total ETH balance: %w", err)
    }
    return *totalEthBalance, nil
}


// Get the current network staking ETH balance
func GetStakingETHBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketNetworkBalances, err := getRocketNetworkBalances(rp)
    if err != nil {
        return nil, err
    }
    stakingEthBalance := new(*big.Int)
    if err := rocketNetworkBalances.Call(opts, stakingEthBalance, "getStakingETHBalance"); err != nil {
        return nil, fmt.Errorf("Could not get network staking ETH balance: %w", err)
    }
    return *stakingEthBalance, nil
}


// Get the current network total rETH supply
func GetTotalRETHSupply(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketNetworkBalances, err := getRocketNetworkBalances(rp)
    if err != nil {
        return nil, err
    }
    totalRethSupply := new(*big.Int)
    if err := rocketNetworkBalances.Call(opts, totalRethSupply, "getTotalRETHSupply"); err != nil {
        return nil, fmt.Errorf("Could not get network total rETH supply: %w", err)
    }
    return *totalRethSupply, nil
}


// Get the current network ETH utilization rate
func GetETHUtilizationRate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkBalances, err := getRocketNetworkBalances(rp)
    if err != nil {
        return 0, err
    }
    ethUtilizationRate := new(*big.Int)
    if err := rocketNetworkBalances.Call(opts, ethUtilizationRate, "getETHUtilizationRate"); err != nil {
        return 0, fmt.Errorf("Could not get network ETH utilization rate: %w", err)
    }
    return eth.WeiToEth(*ethUtilizationRate), nil
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
func getRocketNetworkBalances(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketNetworkBalancesLock.Lock()
    defer rocketNetworkBalancesLock.Unlock()
    return rp.GetContract("rocketNetworkBalances")
}

