package node

import (
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Node details
type NodeDetails struct {
    exists bool
    trusted bool
    timezoneLocation string
}


// Get a node's details
func GetDetails(rp *rocketpool.RocketPool, nodeAddress common.Address) (*NodeDetails, error) {

    // Node data
    var wg errgroup.Group
    var nodeExists bool
    var nodeTrusted bool
    var nodeTimezoneLocation string

    // Get exists status
    wg.Go(func() error {
        exists, err := GetExists(rp, nodeAddress)
        if err == nil { nodeExists = exists }
        return err
    })

    // Get trusted status
    wg.Go(func() error {
        trusted, err := GetTrusted(rp, nodeAddress)
        if err == nil { nodeTrusted = trusted }
        return err
    })

    // Get timezone location
    wg.Go(func() error {
        timezoneLocation, err := GetTimezoneLocation(rp, nodeAddress)
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
func GetExists(rp *rocketpool.RocketPool, nodeAddress common.Address) (bool, error) {

    // Get rocketNodeManager contract
    rocketNodeManager, err := rp.GetContract("rocketNodeManager")
    if err != nil {
        return false, err
    }

    // Get node exists status
    exists := new(bool)
    if err := rocketNodeManager.Call(nil, exists, "getNodeExists", nodeAddress); err != nil {
        return false, fmt.Errorf("Could not get node %v exists status: %w", nodeAddress.Hex(), err)
    }

    // Return
    return *exists, nil

}


// Get a node's trusted status
func GetTrusted(rp *rocketpool.RocketPool, nodeAddress common.Address) (bool, error) {

    // Get rocketNodeManager contract
    rocketNodeManager, err := rp.GetContract("rocketNodeManager")
    if err != nil {
        return false, err
    }

    // Get node trusted status
    trusted := new(bool)
    if err := rocketNodeManager.Call(nil, trusted, "getNodeTrusted", nodeAddress); err != nil {
        return false, fmt.Errorf("Could not get node %v trusted status: %w", nodeAddress.Hex(), err)
    }

    // Return
    return *trusted, nil

}


// Get a node's timezone location
func GetTimezoneLocation(rp *rocketpool.RocketPool, nodeAddress common.Address) (string, error) {

    // Get rocketNodeManager contract
    rocketNodeManager, err := rp.GetContract("rocketNodeManager")
    if err != nil {
        return "", err
    }

    // Get node timezone location
    timezoneLocation := new(string)
    if err := rocketNodeManager.Call(nil, timezoneLocation, "getNodeTimezoneLocation", nodeAddress); err != nil {
        return "", fmt.Errorf("Could not get node %v timezone location: %w", nodeAddress.Hex(), err)
    }

    // Return
    return *timezoneLocation, nil

}

