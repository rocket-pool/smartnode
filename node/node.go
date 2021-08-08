package node

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/storage"
	"github.com/rocket-pool/rocketpool-go/utils/strings"
)

// Settings
const (
    NodeAddressBatchSize = 50
    NodeDetailsBatchSize = 20
)


// Node details
type NodeDetails struct {
    Address common.Address                      `json:"address"`
    Exists bool                                 `json:"exists"`
    WithdrawalAddress common.Address            `json:"withdrawalAddress"`
    PendingWithdrawalAddress common.Address     `json:"pendingWithdrawalAddress"`
    TimezoneLocation string                     `json:"timezoneLocation"`
}


// Count of nodes belonging to a timezone
type TimezoneCount struct {
    Timezone string     `abi:"timezone"`
    Count *big.Int      `abi:"count"`
}


// Get all node details
func GetNodes(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]NodeDetails, error) {

    // Get node addresses
    nodeAddresses, err := GetNodeAddresses(rp, opts)
    if err != nil {
        return []NodeDetails{}, err
    }

    // Load node details in batches
    details := make([]NodeDetails, len(nodeAddresses))
    for bsi := 0; bsi < len(nodeAddresses); bsi += NodeDetailsBatchSize {

        // Get batch start & end index
        nsi := bsi
        nei := bsi + NodeDetailsBatchSize
        if nei > len(nodeAddresses) { nei = len(nodeAddresses) }

        // Load details
        var wg errgroup.Group
        for ni := nsi; ni < nei; ni++ {
            ni := ni
            wg.Go(func() error {
                nodeAddress := nodeAddresses[ni]
                nodeDetails, err := GetNodeDetails(rp, nodeAddress, opts)
                if err == nil { details[ni] = nodeDetails }
                return err
            })
        }
        if err := wg.Wait(); err != nil {
            return []NodeDetails{}, err
        }

    }

    // Return
    return details, nil

}


// Get all node addresses
func GetNodeAddresses(rp *rocketpool.RocketPool, opts *bind.CallOpts) ([]common.Address, error) {

    // Get node count
    nodeCount, err := GetNodeCount(rp, opts)
    if err != nil {
        return []common.Address{}, err
    }

    // Load node addresses in batches
    addresses := make([]common.Address, nodeCount)
    for bsi := uint64(0); bsi < nodeCount; bsi += NodeAddressBatchSize {

        // Get batch start & end index
        nsi := bsi
        nei := bsi + NodeAddressBatchSize
        if nei > nodeCount { nei = nodeCount }

        // Load addresses
        var wg errgroup.Group
        for ni := nsi; ni < nei; ni++ {
            ni := ni
            wg.Go(func() error {
                address, err := GetNodeAt(rp, ni, opts)
                if err == nil { addresses[ni] = address }
                return err
            })
        }
        if err := wg.Wait(); err != nil {
            return []common.Address{}, err
        }

    }

    // Return
    return addresses, nil

}


// Get a node's details
func GetNodeDetails(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (NodeDetails, error) {

    // Data
    var wg errgroup.Group
    var exists bool
    var withdrawalAddress common.Address
    var pendingWithdrawalAddress common.Address
    var timezoneLocation string

    // Load data
    wg.Go(func() error {
        var err error
        exists, err = GetNodeExists(rp, nodeAddress, opts)
        return err
    })
    wg.Go(func() error {
        var err error
        withdrawalAddress, err = storage.GetNodeWithdrawalAddress(rp, nodeAddress, opts)
        return err
    })
    wg.Go(func() error {
        var err error
        pendingWithdrawalAddress, err = storage.GetNodePendingWithdrawalAddress(rp, nodeAddress, opts)
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
        Address: nodeAddress,
        Exists: exists,
        WithdrawalAddress: withdrawalAddress,
        PendingWithdrawalAddress: pendingWithdrawalAddress,
        TimezoneLocation: timezoneLocation,
    }, nil

}


// Get the number of nodes in the network
func GetNodeCount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return 0, err
    }
    nodeCount := new(*big.Int)
    if err := rocketNodeManager.Call(opts, nodeCount, "getNodeCount"); err != nil {
        return 0, fmt.Errorf("Could not get node count: %w", err)
    }
    return (*nodeCount).Uint64(), nil
}


// Get a breakdown of the number of nodes per timezone
func GetNodeCountPerTimezone(rp *rocketpool.RocketPool, offset, limit *big.Int, opts *bind.CallOpts) ([]TimezoneCount, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return []TimezoneCount{}, err
    }
    timezoneCounts := new([]TimezoneCount)
    if err := rocketNodeManager.Call(opts, timezoneCounts, "getNodeCountPerTimezone", offset, limit); err != nil {
        return []TimezoneCount{}, fmt.Errorf("Could not get node count: %w", err)
    }
    return *timezoneCounts, nil
}


// Get a node address by index
func GetNodeAt(rp *rocketpool.RocketPool, index uint64, opts *bind.CallOpts) (common.Address, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return common.Address{}, err
    }
    nodeAddress := new(common.Address)
    if err := rocketNodeManager.Call(opts, nodeAddress, "getNodeAt", big.NewInt(int64(index))); err != nil {
        return common.Address{}, fmt.Errorf("Could not get node %d address: %w", index, err)
    }
    return *nodeAddress, nil
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
    return strings.Sanitize(*timezoneLocation), nil
}


// Estimate the gas of RegisterNode
func EstimateRegisterNodeGas(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return rocketpool.GasInfo{}, err
    }
    _, err = time.LoadLocation(timezoneLocation)
    if err != nil {
        return rocketpool.GasInfo{}, fmt.Errorf("Could not verify timezone [%s]: %w", timezoneLocation, err)
    }
    return rocketNodeManager.GetTransactionGasInfo(opts, "registerNode", timezoneLocation)
}


// Register a node
func RegisterNode(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (common.Hash, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return common.Hash{}, err
    }
    _, err = time.LoadLocation(timezoneLocation)
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not verify timezone [%s]: %w", timezoneLocation, err)
    }
    hash, err := rocketNodeManager.Transact(opts, "registerNode", timezoneLocation)
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not register node: %w", err)
    }
    return hash, nil
}


// Estimate the gas of SetTimezoneLocation
func EstimateSetTimezoneLocationGas(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return rocketpool.GasInfo{}, err
    }
    _, err = time.LoadLocation(timezoneLocation)
    if err != nil {
        return rocketpool.GasInfo{}, fmt.Errorf("Could not verify timezone [%s]: %w", timezoneLocation, err)
    }
    return rocketNodeManager.GetTransactionGasInfo(opts, "setTimezoneLocation", timezoneLocation)
}


// Set a node's timezone location
func SetTimezoneLocation(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (common.Hash, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return common.Hash{}, err
    }
    _, err = time.LoadLocation(timezoneLocation)
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not verify timezone [%s]: %w", timezoneLocation, err)
    }
    hash, err := rocketNodeManager.Transact(opts, "setTimezoneLocation", timezoneLocation)
    if err != nil {
        return common.Hash{}, fmt.Errorf("Could not set node timezone location: %w", err)
    }
    return hash, nil
}


// Get the time when a node was first registered
func GetNodeRegistrationTime(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (time.Time, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return time.Time{}, err
    }

    // Construct a filter query for RocketNodeManager.NodeRegistered(address nodeAddress, uint256 time)
    addressFilter := []common.Address{*rocketNodeManager.Address}
    topicFilter := [][]common.Hash{{rocketNodeManager.ABI.Events["NodeRegistered"].ID}, {nodeAddress.Hash()}}
    logs, err := rp.Client.FilterLogs(context.Background(), ethereum.FilterQuery{
        Addresses: addressFilter,
        Topics: topicFilter,
    })
    if err != nil {
        return time.Time{}, err
    }

    // If this address hasn't been registered, return nothing
    if len(logs) == 0 {
        return time.Time{}, nil
    }
    
    // Get the node registration time
    values := make(map[string]interface{})
    err = rocketNodeManager.ABI.Events["NodeRegistered"].Inputs.UnpackIntoMap(values, logs[0].Data)
    if err != nil {
        return time.Time{}, err
    }
    unixTime := values["time"].(*big.Int)
    time := time.Unix(unixTime.Int64(), 0)
    return time, nil
}


// Get contracts
var rocketNodeManagerLock sync.Mutex
func getRocketNodeManager(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketNodeManagerLock.Lock()
    defer rocketNodeManagerLock.Unlock()
    return rp.GetContract("rocketNodeManager")
}

