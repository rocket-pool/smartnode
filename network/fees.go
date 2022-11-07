package network

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Get the current network node demand in ETH
func GetNodeDemand(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNetworkFees, err := getRocketNetworkFees(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeDemand := new(*big.Int)
	if err := rocketNetworkFees.Call(opts, nodeDemand, "getNodeDemand"); err != nil {
		return nil, fmt.Errorf("Could not get network node demand: %w", err)
	}
	return *nodeDemand, nil
}

// Get the current network node commission rate
func GetNodeFee(rp *rocketpool.RocketPool, opts *bind.CallOpts) (float64, error) {
	rocketNetworkFees, err := getRocketNetworkFees(rp, opts)
	if err != nil {
		return 0, err
	}
	nodeFee := new(*big.Int)
	if err := rocketNetworkFees.Call(opts, nodeFee, "getNodeFee"); err != nil {
		return 0, fmt.Errorf("Could not get network node fee: %w", err)
	}
	return eth.WeiToEth(*nodeFee), nil
}

// Get the network node fee for a node demand value
func GetNodeFeeByDemand(rp *rocketpool.RocketPool, nodeDemand *big.Int, opts *bind.CallOpts) (float64, error) {
	rocketNetworkFees, err := getRocketNetworkFees(rp, opts)
	if err != nil {
		return 0, err
	}
	nodeFee := new(*big.Int)
	if err := rocketNetworkFees.Call(opts, nodeFee, "getNodeFeeByDemand", nodeDemand); err != nil {
		return 0, fmt.Errorf("Could not get node fee by node demand: %w", err)
	}
	return eth.WeiToEth(*nodeFee), nil
}

// Get contracts
var rocketNetworkFeesLock sync.Mutex

func getRocketNetworkFees(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkFeesLock.Lock()
	defer rocketNetworkFeesLock.Unlock()
	return rp.GetContract("rocketNetworkFees", opts)
}
