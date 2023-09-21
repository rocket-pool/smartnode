package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Config
const NetworkSettingsContractName = "rocketDAOProtocolSettingsNetwork"

// The threshold of trusted nodes that must reach consensus on oracle data to commit it
func GetNodeConsensusThreshold(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getNodeConsensusThreshold"); err != nil {
		return 0, fmt.Errorf("Could not get trusted node consensus threshold: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// Network balance submissions currently enabled
func GetSubmitBalancesEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := networkSettingsContract.Call(opts, value, "getSubmitBalancesEnabled"); err != nil {
		return false, fmt.Errorf("Could not get network balance submissions enabled status: %w", err)
	}
	return *value, nil
}

// The frequency in blocks at which network balances should be submitted by trusted nodes
func GetSubmitBalancesFrequency(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getSubmitBalancesFrequency"); err != nil {
		return 0, fmt.Errorf("Could not get network balance submission frequency: %w", err)
	}
	return (*value).Uint64(), nil
}

// Network price submissions currently enabled
func GetSubmitPricesEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := networkSettingsContract.Call(opts, value, "getSubmitPricesEnabled"); err != nil {
		return false, fmt.Errorf("Could not get network price submissions enabled status: %w", err)
	}
	return *value, nil
}

// The frequency in blocks at which network prices should be submitted by trusted nodes
func GetSubmitPricesFrequency(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getSubmitPricesFrequency"); err != nil {
		return 0, fmt.Errorf("Could not get network price submission frequency: %w", err)
	}
	return (*value).Uint64(), nil
}

// Minimum node commission rate
func GetMinimumNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getMinimumNodeFee"); err != nil {
		return 0, fmt.Errorf("Could not get minimum node fee: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// Target node commission rate
func GetTargetNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getTargetNodeFee"); err != nil {
		return 0, fmt.Errorf("Could not get target node fee: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// Maximum node commission rate
func GetMaximumNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getMaximumNodeFee"); err != nil {
		return 0, fmt.Errorf("Could not get maximum node fee: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// The range of node demand values to base fee calculations on
func GetNodeFeeDemandRange(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getNodeFeeDemandRange"); err != nil {
		return nil, fmt.Errorf("Could not get node fee demand range: %w", err)
	}
	return *value, nil
}

// The target collateralization rate for the rETH contract as a fraction
func GetTargetRethCollateralRate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	networkSettingsContract, err := getNetworkSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := networkSettingsContract.Call(opts, value, "getTargetRethCollateralRate"); err != nil {
		return 0, fmt.Errorf("Could not get target rETH contract collateralization rate: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// Get contracts
var networkSettingsContractLock sync.Mutex

func getNetworkSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	networkSettingsContractLock.Lock()
	defer networkSettingsContractLock.Unlock()
	return rp.GetContract(NetworkSettingsContractName, opts)
}
