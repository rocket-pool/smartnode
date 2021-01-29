package node

import (
    "fmt"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Make a node deposit
func Deposit(rp *rocketpool.RocketPool, minimumNodeFee float64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNodeDeposit, err := getRocketNodeDeposit(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNodeDeposit.Transact(opts, "deposit", eth.EthToWei(minimumNodeFee))
    if err != nil {
        return nil, fmt.Errorf("Could not make node deposit: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketNodeDepositLock sync.Mutex
func getRocketNodeDeposit(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketNodeDepositLock.Lock()
    defer rocketNodeDepositLock.Unlock()
    return rp.GetContract("rocketNodeDeposit")
}

