package node

import (
    "fmt"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/contract"
)


// Node details
type NodeDetails struct {
    exists bool
    trusted bool
    timezoneLocation string
}


// Get a node's details
func GetNodeDetails(rp *rocketpool.RocketPool, nodeAddress common.Address) (*NodeDetails, error) {

    // Node data
    var wg errgroup.Group
    var nodeExists bool
    var nodeTrusted bool
    var nodeTimezoneLocation string

    // Get exists status
    wg.Go(func() error {
        exists, err := GetNodeExists(rp, nodeAddress)
        if err == nil { nodeExists = exists }
        return err
    })

    // Get trusted status
    wg.Go(func() error {
        trusted, err := GetNodeTrusted(rp, nodeAddress)
        if err == nil { nodeTrusted = trusted }
        return err
    })

    // Get timezone location
    wg.Go(func() error {
        timezoneLocation, err := GetNodeTimezoneLocation(rp, nodeAddress)
        if err == nil { nodeTimezoneLocation = timezoneLocation }
        return err
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Return
    return &NodeDetails{
        exists: nodeExists,
        trusted: nodeTrusted,
        timezoneLocation: nodeTimezoneLocation,
    }, nil

}


// Check whether a node exists
func GetNodeExists(rp *rocketpool.RocketPool, nodeAddress common.Address) (bool, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return false, err
    }
    exists := new(bool)
    if err := rocketNodeManager.Call(nil, exists, "getNodeExists", nodeAddress); err != nil {
        return false, fmt.Errorf("Could not get node %v exists status: %w", nodeAddress.Hex(), err)
    }
    return *exists, nil
}


// Get a node's trusted status
func GetNodeTrusted(rp *rocketpool.RocketPool, nodeAddress common.Address) (bool, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return false, err
    }
    trusted := new(bool)
    if err := rocketNodeManager.Call(nil, trusted, "getNodeTrusted", nodeAddress); err != nil {
        return false, fmt.Errorf("Could not get node %v trusted status: %w", nodeAddress.Hex(), err)
    }
    return *trusted, nil
}


// Get a node's timezone location
func GetNodeTimezoneLocation(rp *rocketpool.RocketPool, nodeAddress common.Address) (string, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return "", err
    }
    timezoneLocation := new(string)
    if err := rocketNodeManager.Call(nil, timezoneLocation, "getNodeTimezoneLocation", nodeAddress); err != nil {
        return "", fmt.Errorf("Could not get node %v timezone location: %w", nodeAddress.Hex(), err)
    }
    return *timezoneLocation, nil
}


// Register a node
func RegisterNode(rp *rocketpool.RocketPool, timezoneLocation string, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNodeManager, err := getRocketNodeManager(rp)
    if err != nil {
        return nil, err
    }
    return contract.Transact(rp.Client, rocketNodeManager, opts, "registerNode", timezoneLocation)
}


// Get contracts
func getRocketNodeManager(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    return rp.GetContract("rocketNodeManager")
}

