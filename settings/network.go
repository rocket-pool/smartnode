package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// Network balance submissions currently enabled
func GetSubmitBalancesEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return false, err
    }
    submitBalancesEnabled := new(bool)
    if err := rocketNetworkSettings.Call(opts, submitBalancesEnabled, "getSubmitBalancesEnabled"); err != nil {
        return false, fmt.Errorf("Could not get network balance submissions enabled status: %w", err)
    }
    return *submitBalancesEnabled, nil
}


// The frequency in blocks at which network balances should be submitted by trusted nodes
func GetSubmitBalancesFrequency(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return 0, err
    }
    submitBalancesFrequency := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, submitBalancesFrequency, "getSubmitBalancesFrequency"); err != nil {
        return 0, fmt.Errorf("Could not get network balance submission frequency: %w", err)
    }
    return (*submitBalancesFrequency).Uint64(), nil
}


// Processing validator withdrawals currently enabled
func GetProcessWithdrawalsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return false, err
    }
    processWithdrawalsEnabled := new(bool)
    if err := rocketNetworkSettings.Call(opts, processWithdrawalsEnabled, "getProcessWithdrawalsEnabled"); err != nil {
        return false, fmt.Errorf("Could not get processing withdrawals enabled status: %w", err)
    }
    return *processWithdrawalsEnabled, nil
}


// Node commission rate parameters
func GetMinimumNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return 0, err
    }
    minimumNodeFee := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, minimumNodeFee, "getMinimumNodeFee"); err != nil {
        return 0, fmt.Errorf("Could not get minimum node fee: %w", err)
    }
    return eth.WeiToEth(*minimumNodeFee), nil
}
func GetTargetNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return 0, err
    }
    targetNodeFee := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, targetNodeFee, "getTargetNodeFee"); err != nil {
        return 0, fmt.Errorf("Could not get target node fee: %w", err)
    }
    return eth.WeiToEth(*targetNodeFee), nil
}
func GetMaximumNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return 0, err
    }
    maximumNodeFee := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, maximumNodeFee, "getMaximumNodeFee"); err != nil {
        return 0, fmt.Errorf("Could not get maximum node fee: %w", err)
    }
    return eth.WeiToEth(*maximumNodeFee), nil
}
func GetNodeFeeDemandRange(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    nodeFeeDemandRange := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, nodeFeeDemandRange, "getNodeFeeDemandRange"); err != nil {
        return nil, fmt.Errorf("Could not get node fee demand range: %w", err)
    }
    return *nodeFeeDemandRange, nil
}


// Get contracts
var rocketNetworkSettingsLock sync.Mutex
func getRocketNetworkSettings(rp *rocketpool.RocketPool) (*bind.BoundContract, error) {
    rocketNetworkSettingsLock.Lock()
    defer rocketNetworkSettingsLock.Unlock()
    return rp.GetContract("rocketNetworkSettings")
}

