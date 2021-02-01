package settings

import (
    "fmt"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/core/types"

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
func SetNodeConsensusThreshold(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkSettings.Transact(opts, "setNodeConsensusThreshold", eth.EthToWei(value))
    if err != nil {
        return nil, fmt.Errorf("Could not set trusted node consensus threshold: %w", err)
    }
    return txReceipt, nil
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
func SetSubmitBalancesEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkSettings.Transact(opts, "setSubmitBalancesEnabled", value)
    if err != nil {
        return nil, fmt.Errorf("Could not set network balance submissions enabled status: %w", err)
    }
    return txReceipt, nil
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
func SetSubmitBalancesFrequency(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkSettings.Transact(opts, "setSubmitBalancesFrequency", big.NewInt(value))
    if err != nil {
        return nil, fmt.Errorf("Could not set network balance submission frequency: %w", err)
    }
    return txReceipt, nil
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
func SetProcessWithdrawalsEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkSettings.Transact(opts, "setProcessWithdrawalsEnabled", value)
    if err != nil {
        return nil, fmt.Errorf("Could not set processing withdrawals enabled status: %w", err)
    }
    return txReceipt, nil
}


// Minimum node commission rate
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
func SetMinimumNodeFee(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkSettings.Transact(opts, "setMinimumNodeFee", eth.EthToWei(value))
    if err != nil {
        return nil, fmt.Errorf("Could not set minimum node fee: %w", err)
    }
    return txReceipt, nil
}


// Target node commission rate
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
func SetTargetNodeFee(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkSettings.Transact(opts, "setTargetNodeFee", eth.EthToWei(value))
    if err != nil {
        return nil, fmt.Errorf("Could not set target node fee: %w", err)
    }
    return txReceipt, nil
}


// Maximum node commission rate
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
func SetMaximumNodeFee(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkSettings.Transact(opts, "setMaximumNodeFee", eth.EthToWei(value))
    if err != nil {
        return nil, fmt.Errorf("Could not set maximum node fee: %w", err)
    }
    return txReceipt, nil
}


// The range of node demand values to base fee calculations on
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
func SetNodeFeeDemandRange(rp *rocketpool.RocketPool, value *big.Int, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkSettings.Transact(opts, "setNodeFeeDemandRange", value)
    if err != nil {
        return nil, fmt.Errorf("Could not set node fee demand range: %w", err)
    }
    return txReceipt, nil
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
func SetTargetRethCollateralRate(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (*types.Receipt, error) {
    rocketNetworkSettings, err := getRocketNetworkSettings(rp)
    if err != nil {
        return nil, err
    }
    txReceipt, err := rocketNetworkSettings.Transact(opts, "setTargetRethCollateralRate", eth.EthToWei(value))
    if err != nil {
        return nil, fmt.Errorf("Could not set target rETH contract collateralization rate: %w", err)
    }
    return txReceipt, nil
}


// Get contracts
var rocketNetworkSettingsLock sync.Mutex
func getRocketNetworkSettings(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    rocketNetworkSettingsLock.Lock()
    defer rocketNetworkSettingsLock.Unlock()
    return rp.GetContract("rocketNetworkSettings")
}

