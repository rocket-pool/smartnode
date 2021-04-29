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
const InflationSettingsContractName = "rocketDAOProtocolSettingsInflation"


// RPL inflation rate per interval
func GetInflationIntervalRate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
    inflationSettingsContract, err := getInflationSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := inflationSettingsContract.Call(opts, value, "getInflationIntervalRate"); err != nil {
        return 0, fmt.Errorf("Could not get inflation rate: %w", err)
    }
    return eth.WeiToEth(*value), nil
}
func BootstrapInflationIntervalRate(rp *rocketpool.RocketPool, value float64, opts *bind.TransactOpts) (common.Hash, error) {
    return protocoldao.BootstrapUint(rp, InflationSettingsContractName, "rpl.inflation.interval.rate", eth.EthToWei(value), opts)
}


// RPL inflation interval in blocks
func GetInflationIntervalBlocks(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    inflationSettingsContract, err := getInflationSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := inflationSettingsContract.Call(opts, value, "getInflationIntervalBlocks"); err != nil {
        return 0, fmt.Errorf("Could not get inflation interval: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapInflationIntervalBlocks(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
    return protocoldao.BootstrapUint(rp, InflationSettingsContractName, "rpl.inflation.interval.blocks", big.NewInt(int64(value)), opts)
}


// RPL inflation start block
func GetInflationStartBlock(rp *rocketpool.RocketPool, opts *bind.CallOpts) (uint64, error) {
    inflationSettingsContract, err := getInflationSettingsContract(rp)
    if err != nil {
        return 0, err
    }
    value := new(*big.Int)
    if err := inflationSettingsContract.Call(opts, value, "getInflationIntervalStartBlock"); err != nil {
        return 0, fmt.Errorf("Could not get inflation start block: %w", err)
    }
    return (*value).Uint64(), nil
}
func BootstrapInflationStartBlock(rp *rocketpool.RocketPool, value uint64, opts *bind.TransactOpts) (common.Hash, error) {
    return protocoldao.BootstrapUint(rp, InflationSettingsContractName, "rpl.inflation.interval.start", big.NewInt(int64(value)), opts)
}


// Get contracts
var inflationSettingsContractLock sync.Mutex
func getInflationSettingsContract(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
    inflationSettingsContractLock.Lock()
    defer inflationSettingsContractLock.Unlock()
    return rp.GetContract(InflationSettingsContractName)
}

