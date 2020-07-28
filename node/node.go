package node

import (
    "fmt"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/contract"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Node details
type NodeDetails struct {
    Exists bool                 `json:"exists"`
    Trusted bool                `json:"trusted"`
    TimezoneLocation string     `json:"timezoneLocation"`
}


// Get a node's details
func GetNodeDetails(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (NodeDetails, error) {

    // Data
    var wg errgroup.Group
    var exists bool
    var trusted bool
    var timezoneLocation string

    // Load data
    wg.Go(func() error {
        var err error
        exists, err = GetNodeExists(rp, nodeAddress, opts)
        return err
    })
    wg.Go(func() error {
        var err error
        trusted, err = GetNodeTrusted(rp, nodeAddress, opts)
        return err
    })
    wg.Go(func() error {
        var err error
        timezoneLocation, err = GetNodeTimezoneLocation(rp, nodeAddress, opts)
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return NodeDetails{}, err
    }

    // Return
    return NodeDetails{
        Exists: exists,
        Trusted: trusted,
        TimezoneLocation: timezoneLocation,
    }, nil

}


// Check whether a node exists
func GetNodeExists(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (bool, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return false, err
    }
    exists := new(bool)
    if err := rocketNodeManager.Call(opts, exists, "getNodeExists", nodeAddress); err != nil {
        return false, fmt.Errorf("Could not get node %s exists status: %w", nodeAddress.Hex(), err)
    }
    return *exists, nil
}


// Get a node's trusted status
func GetNodeTrusted(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (bool, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return false, err
    }
    trusted := new(bool)
    if err := rocketNodeManager.Call(opts, trusted, "getNodeTrusted", nodeAddress); err != nil {
        return false, fmt.Errorf("Could not get node %s trusted status: %w", nodeAddress.Hex(), err)
    }
    return *trusted, nil
}


// Get a node's timezone location
func GetNodeTimezoneLocation(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (string, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return "", err
    }
    timezoneLocation := new(string)
    if err := rocketNodeManager.Call(opts, timezoneLocation, "getNodeTimezoneLocation", nodeAddress); err != nil {
        return "", fmt.Errorf("Could not get node %s timezone location: %w", nodeAddress.Hex(), err)
    }
    return *timezoneLocation, nil
}


// Register a node
func RegisterNode(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := contract.Transact(rp.Client, rocketNodeManager, opts, "registerNode", timezoneLocation)
    if err != nil {
        return nil, fmt.Errorf("Could not register node: %w", err)
    }
    return txReceipt, nil
}


// Set a node's timezone location
func SetTimezoneLocation(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := contract.Transact(rp.Client, rocketNodeManager, opts, "setTimezoneLocation", timezoneLocation)
    if err != nil {
        return nil, fmt.Errorf("Could not set node timezone location: %w", err)
    }
    return txReceipt, nil
}


// Make a node deposit
func Deposit(rp *rocketpool.RocketPool, minimumNodeFee float64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNodeDeposit, err := getRocketNodeDeposit(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := contract.Transact(rp.Client, rocketNodeDeposit, opts, "deposit", eth.EthToWei(minimumNodeFee))
    if err != nil {
        return nil, fmt.Errorf("Could not make node deposit: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketNodeManagerLock sync.Mutex
func getRocketNodeManager(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNodeManagerLock.Lock()
    defer rocketNodeManagerLock.Unlock()
    return rp.GetContract("rocketNodeManager")
}
var rocketNodeDepositLock sync.Mutex
func getRocketNodeDeposit(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNodeDepositLock.Lock()
    defer rocketNodeDepositLock.Unlock()
    return rp.GetContract("rocketNodeDeposit")
}

