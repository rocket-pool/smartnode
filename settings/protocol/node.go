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
const NodeSettingsContractName = "rocketDAOProtocolSettingsNode"

// Node registrations currently enabled
func GetNodeRegistrationEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := nodeSettingsContract.Call(opts, value, "getRegistrationEnabled"); err != nil {
		return false, fmt.Errorf("Could not get node registrations enabled status: %w", err)
	}
	return *value, nil
}
func BootstrapNodeRegistrationEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapBool(rp, NodeSettingsContractName, "node.registration.enabled", value, opts)
}

// Node deposits currently enabled
func GetNodeDepositEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := nodeSettingsContract.Call(opts, value, "getDepositEnabled"); err != nil {
		return false, fmt.Errorf("Could not get node deposits enabled status: %w", err)
	}
	return *value, nil
}
func BootstrapNodeDepositEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapBool(rp, NodeSettingsContractName, "node.deposit.enabled", value, opts)
}

// The minimum RPL stake per minipool as a fraction of assigned user ETH
func GetMinimumPerMinipoolStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMinimumPerMinipoolStake"); err != nil {
		return 0, fmt.Errorf("Could not get minimum RPL stake per minipool: %w", err)
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
		return nil, fmt.Errorf("Could not get minimum RPL stake per minipool: %w", err)
	}
	return *value, nil
}
func BootstrapMinimumPerMinipoolStake(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, NodeSettingsContractName, "node.per.minipool.stake.minimum", eth.EthToWei(value), opts)
}

// The maximum RPL stake per minipool as a fraction of assigned user ETH
func GetMaximumPerMinipoolStake(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	nodeSettingsContract, err := getNodeSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := nodeSettingsContract.Call(opts, value, "getMaximumPerMinipoolStake"); err != nil {
		return 0, fmt.Errorf("Could not get maximum RPL stake per minipool: %w", err)
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
		return nil, fmt.Errorf("Could not get maximum RPL stake per minipool: %w", err)
	}
	return *value, nil
}
func BootstrapMaximumPerMinipoolStake(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, NodeSettingsContractName, "node.per.minipool.stake.maximum", eth.EthToWei(value), opts)
}

// Get contracts
var nodeSettingsContractLock sync.Mutex

func getNodeSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	nodeSettingsContractLock.Lock()
	defer nodeSettingsContractLock.Unlock()
	return rp.GetContract(NodeSettingsContractName, opts)
}
