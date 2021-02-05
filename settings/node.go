package settings

import (
    "fmt"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Config
const NodeSettingsContractName = "rocketDAOProtocolSettingsNode"


// Node registrations currently enabled
func GetNodeRegistrationEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    nodeSettingsContract, err := getNodeSettingsContract(rp)
    if err != nil {
        return false, err
    }
    value := new(bool)
    if err := nodeSettingsContract.Call(opts, value, "getRegistrationEnabled"); err != nil {
        return false, fmt.Errorf("Could not get node registrations enabled status: %w", err)
    }
    return *value, nil
}
func BootstrapNodeRegistrationEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    return bootstrapBool(rp, NodeSettingsContractName, "node.registration.enabled", value, opts)
}


// Node deposits currently enabled
func GetNodeDepositEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    nodeSettingsContract, err := getNodeSettingsContract(rp)
    if err != nil {
        return false, err
    }
    value := new(bool)
    if err := nodeSettingsContract.Call(opts, value, "getDepositEnabled"); err != nil {
        return false, fmt.Errorf("Could not get node deposits enabled status: %w", err)
    }
    return *value, nil
}
func BootstrapNodeDepositEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    return bootstrapBool(rp, NodeSettingsContractName, "node.deposit.enabled", value, opts)
}


// Get contracts
var nodeSettingsContractLock sync.Mutex
func getNodeSettingsContract(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    nodeSettingsContractLock.Lock()
    defer nodeSettingsContractLock.Unlock()
    return rp.GetContract(NodeSettingsContractName)
}

