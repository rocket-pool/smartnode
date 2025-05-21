package network

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

type RevenueSplit struct {
	NodeShare  *big.Int `abi:"nodeShare"`
	VoterShare *big.Int `abi:"voterShare"`
	RethShare  *big.Int `abi:"rethShare"`
}

// Get the current node share
func GetCurrentNodeShare(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*big.Int, error) {
	rocketNetworkRevenues, err := getRocketNetworkRevenues(rp, opts)
	if err != nil {
		return nil, err
	}
	nodeShare := new(*big.Int)
	if err := rocketNetworkRevenues.Call(opts, nodeShare, "getCurrentNodeShare"); err != nil {
		return nil, fmt.Errorf("error getting network node share: %w", err)
	}
	return *nodeShare, nil
}

// Calculates the time-weighted average revenue split values between the supplied block number and now
func CalculateSplit(rp *rocketpool.RocketPool, sinceBlock uint64, opts *bind.CallOpts) (RevenueSplit, error) {
	rocketNetworkRevenues, err := getRocketNetworkRevenues(rp, opts)
	if err != nil {
		return RevenueSplit{}, err
	}
	revenueSplit := new(RevenueSplit)
	if err := rocketNetworkRevenues.Call(opts, revenueSplit, "calculateSplit", big.NewInt(int64(sinceBlock))); err != nil {
		return RevenueSplit{}, fmt.Errorf("error calculating the revenue split %w", err)
	}
	return *revenueSplit, nil
}

// Get contracts
var rocketNetworkRevenuesLock sync.Mutex

func getRocketNetworkRevenues(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkRevenuesLock.Lock()
	defer rocketNetworkRevenuesLock.Unlock()
	return rp.GetContract("rocketNetworkRevenues", opts)
}
