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
const NodeSettingsContractName = "rocketDAOProtocolSettingsNode"

// Node registrations currently enabled
func GetNodeRegistrationEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := nodeSettingsContract.Call(opts, value, "getRegistrationEnabled"); err != nil {
		return false, fmt.Errorf("error getting node registrations enabled status: %w", err)
	}
	return *value, nil
}

// Node deposits currently enabled
func GetNodeDepositEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := nodeSettingsContract.Call(opts, value, "getDepositEnabled"); err != nil {
		return false, fmt.Errorf("error getting node deposits enabled status: %w", err)
	}
	return *value, nil
}

// Vacant minipools currently enabled
func GetVacantMinipoolsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := nodeSettingsContract.Call(opts, value, "getVacantMinipoolsEnabled"); err != nil {
		return false, fmt.Errorf("error getting vacant minipools enabled status: %w", err)
	}
	return *value, nil
}

// The minimum RPL stake per minipool as a fraction of assigned user ETH
func GetMinimumPerMinipoolStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMinimumPerMinipoolStake"); err != nil {
		return 0, fmt.Errorf("error getting minimum RPL stake per minipool: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// The minimum RPL stake per minipool as a fraction of assigned user ETH
func GetMinimumPerMinipoolStakeRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMinimumPerMinipoolStake"); err != nil {
		return nil, fmt.Errorf("error getting minimum RPL stake per minipool: %w", err)
	}
	return *value, nil
}

// The maximum RPL stake per minipool as a fraction of assigned user ETH
func GetMaximumPerMinipoolStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMaximumPerMinipoolStake"); err != nil {
		return 0, fmt.Errorf("error getting maximum RPL stake per minipool: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// The maximum RPL stake per minipool as a fraction of assigned user ETH
func GetMaximumPerMinipoolStakeRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMaximumPerMinipoolStake"); err != nil {
		return nil, fmt.Errorf("error getting maximum RPL stake per minipool: %w", err)
	}
	return *value, nil
}

// Get contracts
var nodeSettingsContractLock sync.Mutex

func getNodeSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	nodeSettingsContractLock.Lock()
	defer nodeSettingsContractLock.Unlock()
	return rp.GetContract(NodeSettingsContractName, opts)
}
