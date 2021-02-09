package trustednode

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Bootstrap a bool setting
func BootstrapBool(rp *rocketpool.RocketPool, contractName, settingPath string, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrusted.Transact(opts, "bootstrapSettingBool", contractName, settingPath, value)
    if err != nil {
        return nil, fmt.Errorf("Could not bootstrap trusted node setting %s.%s: %w", contractName, settingPath, err)
    }
    return txReceipt, nil
}


// Bootstrap a uint256 setting
func BootstrapUint(rp *rocketpool.RocketPool, contractName, settingPath string, value *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrusted.Transact(opts, "bootstrapSettingUint", contractName, settingPath, value)
    if err != nil {
        return nil, fmt.Errorf("Could not bootstrap trusted node setting %s.%s: %w", contractName, settingPath, err)
    }
    return txReceipt, nil
}


// Bootstrap a DAO member
func BootstrapMember(rp *rocketpool.RocketPool, id, email string, nodeAddress common.Address, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketDAONodeTrusted, err := getRocketDAONodeTrusted(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketDAONodeTrusted.Transact(opts, "bootstrapMember", id, email, nodeAddress)
    if err != nil {
        return nil, fmt.Errorf("Could not bootstrap trusted node member %s: %w", id, err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketDAONodeTrustedLock sync.Mutex
func getRocketDAONodeTrusted(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketDAONodeTrustedLock.Lock()
    defer rocketDAONodeTrustedLock.Unlock()
    return rp.GetContract("rocketDAONodeTrusted")
}

