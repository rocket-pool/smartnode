package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	protocoldao "github.com/rocket-pool/rocketpool-go/dao/protocol"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Config
const DepositSettingsContractName = "rocketDAOProtocolSettingsDeposit"

// Deposits currently enabled
func GetDepositEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := depositSettingsContract.Call(opts, value, "getDepositEnabled"); err != nil {
		return false, fmt.Errorf("Could not get deposits enabled status: %w", err)
	}
	return *value, nil
}
func BootstrapDepositEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapBool(rp, DepositSettingsContractName, "deposit.enabled", value, opts)
}

// Deposit assignments currently enabled
func GetAssignDepositsEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := depositSettingsContract.Call(opts, value, "getAssignDepositsEnabled"); err != nil {
		return false, fmt.Errorf("Could not get deposit assignments enabled status: %w", err)
	}
	return *value, nil
}
func BootstrapAssignDepositsEnabled(rp *rocketpool.RocketPool, value bool, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapBool(rp, DepositSettingsContractName, "deposit.assign.enabled", value, opts)
}

// Minimum deposit amount
func GetMinimumDeposit(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := depositSettingsContract.Call(opts, value, "getMinimumDeposit"); err != nil {
		return nil, fmt.Errorf("Could not get minimum deposit amount: %w", err)
	}
	return *value, nil
}
func BootstrapMinimumDeposit(rp *rocketpool.RocketPool, value *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, DepositSettingsContractName, "deposit.minimum", value, opts)
}

// Maximum deposit pool size
func GetMaximumDepositPoolSize(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := depositSettingsContract.Call(opts, value, "getMaximumDepositPoolSize"); err != nil {
		return nil, fmt.Errorf("Could not get maximum deposit pool size: %w", err)
	}
	return *value, nil
}
func BootstrapMaximumDepositPoolSize(rp *rocketpool.RocketPool, value *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, DepositSettingsContractName, "deposit.pool.maximum", value, opts)
}

// Maximum deposit assignments per transaction
func GetMaximumDepositAssignments(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	depositSettingsContract, err := getDepositSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := depositSettingsContract.Call(opts, value, "getMaximumDepositAssignments"); err != nil {
		return 0, fmt.Errorf("Could not get maximum deposit assignments: %w", err)
	}
	return (*value).Uint64(), nil
}
func BootstrapMaximumDepositAssignments(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
	return protocoldao.BootstrapUint(rp, DepositSettingsContractName, "deposit.assign.maximum", big.NewInt(int64(value)), opts)
}

// Get contracts
var depositSettingsContractLock sync.Mutex

func getDepositSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	depositSettingsContractLock.Lock()
	defer depositSettingsContractLock.Unlock()
	return rp.GetContract(DepositSettingsContractName, opts)
}
