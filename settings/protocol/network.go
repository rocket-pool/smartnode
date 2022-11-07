package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	protocoldao "github.com/rocket-pool/rocketpool-go/dao/protocol"
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
func BootstrapNodeConsensusThreshold(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, NetworkSettingsContractName, "network.consensus.threshold", eth.EthToWei(value), opts)
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
func BootstrapSubmitBalancesEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapBool(rp, NetworkSettingsContractName, "network.submit.balances.enabled", value, opts)
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
func BootstrapSubmitBalancesFrequency(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, NetworkSettingsContractName, "network.submit.balances.frequency", big.NewInt(int64(value)), opts)
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
func BootstrapSubmitPricesEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapBool(rp, NetworkSettingsContractName, "network.submit.prices.enabled", value, opts)
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
func BootstrapSubmitPricesFrequency(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, NetworkSettingsContractName, "network.submit.prices.frequency", big.NewInt(int64(value)), opts)
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
func BootstrapMinimumNodeFee(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, NetworkSettingsContractName, "network.node.fee.minimum", eth.EthToWei(value), opts)
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
func BootstrapTargetNodeFee(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, NetworkSettingsContractName, "network.node.fee.target", eth.EthToWei(value), opts)
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
func BootstrapMaximumNodeFee(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, NetworkSettingsContractName, "network.node.fee.maximum", eth.EthToWei(value), opts)
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
func BootstrapNodeFeeDemandRange(rp *rocketpool.RocketPool, value *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, NetworkSettingsContractName, "network.node.fee.demand.range", value, opts)
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
func BootstrapTargetRethCollateralRate(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, NetworkSettingsContractName, "network.reth.collateral.target", eth.EthToWei(value), opts)
}

// Get contracts
var networkSettingsContractLock sync.Mutex

func getNetworkSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	networkSettingsContractLock.Lock()
	defer networkSettingsContractLock.Unlock()
	return rp.GetContract(NetworkSettingsContractName, opts)
}
