package protocol

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/units"
)

// Config
const (
	InflationSettingsContractName string = "rocketDAOProtocolSettingsInflation"
)

// RPL inflation rate per interval
func GetInflationIntervalRate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (units.Eth, error) {
	inflationSettingsContract, err := getInflationSettingsContract(rp, opts)
	if err != nil {
		return units.Eth{}, err
	}
	var value units.Wei
	if err := inflationSettingsContract.Call(opts, &value, "getInflationIntervalRate"); err != nil {
		return units.Eth{}, fmt.Errorf("error getting inflation rate: %w", err)
	}
	return value.ToEth(), nil
}

// RPL inflation rate per interval
func GetInflationIntervalRateRaw(rp *rocketpool.RocketPool, opts *bind.CallOpts) (units.Wei, error) {
	inflationSettingsContract, err := getInflationSettingsContract(rp, opts)
	if err != nil {
		return units.Wei{}, err
	}
	value := new(units.Wei)
	if err := inflationSettingsContract.Call(opts, value, "getInflationIntervalRate"); err != nil {
		return units.Wei{}, fmt.Errorf("error getting inflation rate: %w", err)
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
