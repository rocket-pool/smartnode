package rocketpool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// rocketpool-go's dependencies are all inverted: the subfolders have dependencies back to
// the main package. This is less than ideal, but hard to fix- instead, I will be migrating content
// out of the subpackages into the main package to fulfill interfaces as needed.

// Get the index of the active rewards period
func (rp *RocketPool) GetRewardIndex(opts *bind.CallOpts) (*big.Int, error) {
	rocketRewardsPool, err := getRocketRewardsPool(rp, opts)
	if err != nil {
		return nil, err
	}
	index := new(*big.Int)
	if err := rocketRewardsPool.Call(opts, index, "getRewardIndex"); err != nil {
		return nil, fmt.Errorf("error getting current reward index: %w", err)
	}
	return *index, nil
}

// Get contracts
var rocketRewardsPoolLock sync.Mutex

func getRocketRewardsPool(rp *RocketPool, opts *bind.CallOpts) (*Contract, error) {
	rocketRewardsPoolLock.Lock()
	defer rocketRewardsPoolLock.Unlock()
	return rp.GetContract("rocketRewardsPool", opts)
}
