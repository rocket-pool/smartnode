package protocol

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Config
const MinipoolSettingsContractName = "rocketDAOProtocolSettingsMinipool"

// Get the minipool launch balance
func GetMinipoolLaunchBalance(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getLaunchBalance"); err != nil {
		return nil, fmt.Errorf("error getting minipool launch balance: %w", err)
	}
	return *value, nil
}

// Required node deposit amounts
func GetMinipoolFullDepositNodeAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getFullDepositNodeAmount"); err != nil {
		return nil, fmt.Errorf("error getting full minipool deposit node amount: %w", err)
	}
	return *value, nil
}
func GetMinipoolHalfDepositNodeAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getHalfDepositNodeAmount"); err != nil {
		return nil, fmt.Errorf("error getting half minipool deposit node amount: %w", err)
	}
	return *value, nil
}
func GetMinipoolEmptyDepositNodeAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getEmptyDepositNodeAmount"); err != nil {
		return nil, fmt.Errorf("error getting empty minipool deposit node amount: %w", err)
	}
	return *value, nil
}

// Required user deposit amounts
func GetMinipoolFullDepositUserAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getFullDepositUserAmount"); err != nil {
		return nil, fmt.Errorf("error getting full minipool deposit user amount: %w", err)
	}
	return *value, nil
}
func GetMinipoolHalfDepositUserAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getHalfDepositUserAmount"); err != nil {
		return nil, fmt.Errorf("error getting half minipool deposit user amount: %w", err)
	}
	return *value, nil
}
func GetMinipoolEmptyDepositUserAmount(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getEmptyDepositUserAmount"); err != nil {
		return nil, fmt.Errorf("error getting empty minipool deposit user amount: %w", err)
	}
	return *value, nil
}

// Minipool withdrawable event submissions currently enabled
func GetMinipoolSubmitWithdrawableEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := minipoolSettingsContract.Call(opts, value, "getSubmitWithdrawableEnabled"); err != nil {
		return false, fmt.Errorf("error getting minipool withdrawable submissions enabled status: %w", err)
	}
	return *value, nil
}

// Timeout period in seconds for prelaunch minipools to launch
func GetMinipoolLaunchTimeout(rp *rocketpool.RocketPool, opts *bind.CallOpts) (time.Duration, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getLaunchTimeout"); err != nil {
		return 0, fmt.Errorf("error getting minipool launch timeout: %w", err)
	}
	seconds := time.Duration((*value).Int64()) * time.Second
	return seconds, nil
}

// Timeout period in seconds for prelaunch minipools to launch
func GetMinipoolLaunchTimeoutRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := minipoolSettingsContract.Call(opts, value, "getLaunchTimeout"); err != nil {
		return nil, fmt.Errorf("error getting minipool launch timeout: %w", err)
	}
	return *value, nil
}

// Minipool bond reductions currently enabled
func GetBondReductionEnabled(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	minipoolSettingsContract, err := getMinipoolSettingsContract(rp, opts)
	if err != nil {
		return false, err
	}
	value := new(bool)
	if err := minipoolSettingsContract.Call(opts, value, "getBondReductionEnabled"); err != nil {
		return false, fmt.Errorf("error getting bond reduction enabled status: %w", err)
	}
	return *value, nil
}

// Get contracts
var minipoolSettingsContractLock sync.Mutex

func getMinipoolSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	minipoolSettingsContractLock.Lock()
	defer minipoolSettingsContractLock.Unlock()
	return rp.GetContract(MinipoolSettingsContractName, opts)
}
