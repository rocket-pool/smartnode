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


// Get the current network node commission rate
func GetNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkFees, err := getRocketNetworkFees(rp)
    if err != nil {
        return 0, err
    }
    nodeFee := new(*big.Int)
    if err := rocketNetworkFees.Call(opts, nodeFee, "getNodeFee"); err != nil {
        return 0, fmt.Errorf("Could not get network node fee: %w", err)
    }
    return eth.WeiToEth(*nodeFee), nil
}


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


// Process a validator withdrawal from the beacon chain
func ProcessWithdrawal(rp *rocketpool.RocketPool, validatorPubkey rptypes.ValidatorPubkey, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkWithdrawal, err := getRocketNetworkWithdrawal(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := contract.Transact(rp.Client, rocketNetworkWithdrawal, opts, "processWithdrawal", validatorPubkey)
    if err != nil {
        return nil, fmt.Errorf("Could not process validator %s withdrawal: %w", validatorPubkey.Hex(), err)
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
var rocketNetworkFeesLock sync.Mutex
func getRocketNetworkFees(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNetworkFeesLock.Lock()
    defer rocketNetworkFeesLock.Unlock()
    return rp.GetContract("rocketNetworkFees")
}
var rocketNetworkWithdrawalLock sync.Mutex
func getRocketNetworkWithdrawal(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNetworkWithdrawalLock.Lock()
    defer rocketNetworkWithdrawalLock.Unlock()
    return rp.GetContract("rocketNetworkWithdrawal")
}

