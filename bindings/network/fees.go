package network

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/shared/units"
)

// Get the current network node demand in ETH
func GetNodeDemand(rp *rocketpool.RocketPool, opts *bind.CallOpts) (units.Wei, error) {
	rocketNetworkFees, err := getRocketNetworkFees(rp, opts)
	if err != nil {
		return units.Wei{}, err
	}
	nodeDemand := new(units.Wei)
	if err := rocketNetworkFees.Call(opts, nodeDemand, "getNodeDemand"); err != nil {
		return units.Wei{}, fmt.Errorf("error getting network node demand: %w", err)
	}
	return *nodeDemand, nil
}

// Get the current network node commission rate
func GetNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (units.Eth, error) {
	rocketNetworkFees, err := getRocketNetworkFees(rp, opts)
	if err != nil {
		return units.Eth{}, err
	}
	var nodeFee units.Wei
	if err := rocketNetworkFees.Call(opts, &nodeFee, "getNodeFee"); err != nil {
		return units.Eth{}, fmt.Errorf("error getting network node fee: %w", err)
	}
	return nodeFee.ToEth(), nil
}

// Get the network node fee for a node demand value
func GetNodeFeeByDemand(rp *rocketpool.RocketPool, nodeDemand *big.Int, opts *bind.CallOpts) (units.Eth, error) {
	rocketNetworkFees, err := getRocketNetworkFees(rp, opts)
	if err != nil {
		return units.Eth{}, err
	}
	var nodeFee units.Wei
	if err := rocketNetworkFees.Call(opts, &nodeFee, "getNodeFeeByDemand", nodeDemand); err != nil {
		return units.Eth{}, fmt.Errorf("error getting node fee by node demand: %w", err)
	}
	return nodeFee.ToEth(), nil
}

// Get contracts
var rocketNetworkFeesLock sync.Mutex

func getRocketNetworkFees(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkFeesLock.Lock()
	defer rocketNetworkFeesLock.Unlock()
	return rp.GetContract("rocketNetworkFees", opts)
}
