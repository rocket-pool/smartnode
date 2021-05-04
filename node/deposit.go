package node

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Estimate the gas of Deposit
func EstimateDepositGas(rp *rocketpool.RocketPool, minimumNodeFee float64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
    rocketNodeDeposit, err := getRocketNodeDeposit(rp)
    if err != nil {
        return rocketpool.GasInfo{}, err
    }
    return rocketNodeDeposit.GetTransactionGasInfo(opts, "deposit", eth.EthToWei(minimumNodeFee))
}


// Make a node deposit
func Deposit(rp *rocketpool.RocketPool, minimumNodeFee float64, opts *bind.TransactOpts) (common.Hash, error) {
    rocketNodeDeposit, err := getRocketNodeDeposit(rp)
    if err != nil {
        return common.Hash{}, err
    }
    hash, err := rocketNodeDeposit.Transact(opts, "deposit", eth.EthToWei(minimumNodeFee))
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not make node deposit: %w", err)
    }
    return hash, nil
}


// Get contracts
var rocketNodeDepositLock sync.Mutex
func getRocketNodeDeposit(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketNodeDepositLock.Lock()
    defer rocketNodeDepositLock.Unlock()
    return rp.GetContract("rocketNodeDeposit")
}

