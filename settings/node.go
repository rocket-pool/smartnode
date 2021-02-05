package settings

import (
    "fmt"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Node registrations currently enabled
func GetNodeRegistrationEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketDAOProtocolSettingsNode, err := getRocketDAOProtocolSettingsNode(rp)
    if err != nil {
        return false, err
    }
    value := new(bool)
    if err := rocketDAOProtocolSettingsNode.Call(opts, value, "getRegistrationEnabled"); err != nil {
        return false, fmt.Errorf("Could not get node registrations enabled status: %w", err)
    }
    return *value, nil
}
func SetNodeRegistrationEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAOProtocolSettingsNode, err := getRocketDAOProtocolSettingsNode(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAOProtocolSettingsNode.Transact(opts, "setRegistrationEnabled", value)
    if err != nil {
        return nil, fmt.Errorf("Could not set node registrations enabled status: %w", err)
    }
    return txReceipt, nil
}


// Node deposits currently enabled
func GetNodeDepositEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketDAOProtocolSettingsNode, err := getRocketDAOProtocolSettingsNode(rp)
    if err != nil {
        return false, err
    }
    value := new(bool)
    if err := rocketDAOProtocolSettingsNode.Call(opts, value, "getDepositEnabled"); err != nil {
        return false, fmt.Errorf("Could not get node deposits enabled status: %w", err)
    }
    return *value, nil
}
func SetNodeDepositEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAOProtocolSettingsNode, err := getRocketDAOProtocolSettingsNode(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAOProtocolSettingsNode.Transact(opts, "setDepositEnabled", value)
    if err != nil {
        return nil, fmt.Errorf("Could not set node deposits enabled status: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketDAOProtocolSettingsNodeLock sync.Mutex
func getRocketDAOProtocolSettingsNode(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketDAOProtocolSettingsNodeLock.Lock()
    defer rocketDAOProtocolSettingsNodeLock.Unlock()
    return rp.GetContract("rocketDAOProtocolSettingsNode")
}

