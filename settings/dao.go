package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Bootstrap a bool setting
func bootstrapBool(rp *rocketpool.RocketPool, contractName, settingPath string, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAOProtocol, err := getRocketDAOProtocol(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAOProtocol.Transact(opts, "bootstrapSettingBool", contractName, settingPath, value)
    if err != nil {
        return nil, fmt.Errorf("Could not bootstrap setting %s.%s: %w", contractName, settingPath, err)
    }
    return txReceipt, nil
}


// Bootstrap a uint256 setting
func bootstrapUint(rp *rocketpool.RocketPool, contractName, settingPath string, value *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAOProtocol, err := getRocketDAOProtocol(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAOProtocol.Transact(opts, "bootstrapSettingUint", contractName, settingPath, value)
    if err != nil {
        return nil, fmt.Errorf("Could not bootstrap setting %s.%s: %w", contractName, settingPath, err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketDAOProtocolLock sync.Mutex
func getRocketDAOProtocol(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketDAOProtocolLock.Lock()
    defer rocketDAOProtocolLock.Unlock()
    return rp.GetContract("rocketDAOProtocol")
}

