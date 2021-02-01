package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


// The threshold of trusted nodes that must reach consensus on oracle data to commit it
func GetNodeConsensusThreshold(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, value, "getNodeConsensusThreshold"); err != nil {
        return 0, fmt.Errorf("Could not get trusted node consensus threshold: %w", err)
    }
    return eth.WeiToEth(*value), nil
}


// Network balance submissions currently enabled
func GetSubmitBalancesEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return false, err
    }
    value := new(bool)
    if err := rocketNetworkSettings.Call(opts, value, "getSubmitBalancesEnabled"); err != nil {
        return false, fmt.Errorf("Could not get network balance submissions enabled status: %w", err)
    }
    return *value, nil
}


// The frequency in blocks at which network balances should be submitted by trusted nodes
func GetSubmitBalancesFrequency(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, value, "getSubmitBalancesFrequency"); err != nil {
        return 0, fmt.Errorf("Could not get network balance submission frequency: %w", err)
    }
    return (*value).Uint64(), nil
}


// Processing validator withdrawals currently enabled
func GetProcessWithdrawalsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return false, err
    }
    value := new(bool)
    if err := rocketNetworkSettings.Call(opts, value, "getProcessWithdrawalsEnabled"); err != nil {
        return false, fmt.Errorf("Could not get processing withdrawals enabled status: %w", err)
    }
    return *value, nil
}


// Node commission rate parameters
func GetMinimumNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, value, "getMinimumNodeFee"); err != nil {
        return 0, fmt.Errorf("Could not get minimum node fee: %w", err)
    }
    return eth.WeiToEth(*value), nil
}
func GetTargetNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, value, "getTargetNodeFee"); err != nil {
        return 0, fmt.Errorf("Could not get target node fee: %w", err)
    }
    return eth.WeiToEth(*value), nil
}
func GetMaximumNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, value, "getMaximumNodeFee"); err != nil {
        return 0, fmt.Errorf("Could not get maximum node fee: %w", err)
    }
    return eth.WeiToEth(*value), nil
}
func GetNodeFeeDemandRange(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    value := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, value, "getNodeFeeDemandRange"); err != nil {
        return nil, fmt.Errorf("Could not get node fee demand range: %w", err)
    }
    return *value, nil
}


// The target collateralization rate for the rETH contract as a fraction
func GetTargetRethCollateralRate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := rocketNetworkSettings.Call(opts, value, "getTargetRethCollateralRate"); err != nil {
        return 0, fmt.Errorf("Could not get target rETH contract collateralization rate: %w", err)
    }
    return eth.WeiToEth(*value), nil
}


// Get contracts
var rocketNetworkSettingsLock sync.Mutex
func getRocketNetworkSettings(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketNetworkSettingsLock.Lock()
    defer rocketNetworkSettingsLock.Unlock()
    return rp.GetContract("rocketNetworkSettings")
}

