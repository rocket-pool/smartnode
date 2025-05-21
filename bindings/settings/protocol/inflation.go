package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
)

// Config
const (
	InflationSettingsContractName string = "rocketDAOProtocolSettingsInflation"
)

// RPL inflation rate per interval
func GetInflationIntervalRate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	inflationSettingsContract, err := getInflationSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := inflationSettingsContract.Call(opts, value, "getInflationIntervalRate"); err != nil {
		return 0, fmt.Errorf("error getting inflation rate: %w", err)
	}
	return eth.WeiToEth(*value), nil
}

// RPL inflation rate per interval
func GetInflationIntervalRateRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	inflationSettingsContract, err := getInflationSettingsContract(rp, opts)
	if err != nil {
		return nil, err
	}
	value := new(*big.Int)
	if err := inflationSettingsContract.Call(opts, value, "getInflationIntervalRate"); err != nil {
		return nil, fmt.Errorf("error getting inflation rate: %w", err)
	}
	return *value, nil
}

// RPL inflation start time
func GetInflationStartTime(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
	inflationSettingsContract, err := getInflationSettingsContract(rp, opts)
	if err != nil {
		return 0, err
	}
	value := new(*big.Int)
	if err := inflationSettingsContract.Call(opts, value, "getInflationIntervalStartTime"); err != nil {
		return 0, fmt.Errorf("error getting inflation start time: %w", err)
	}
	return (*value).Uint64(), nil
}

// Get contracts
var inflationSettingsContractLock sync.Mutex

func getInflationSettingsContract(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	inflationSettingsContractLock.Lock()
	defer inflationSettingsContractLock.Unlock()
	return rp.GetContract(InflationSettingsContractName, opts)
}
