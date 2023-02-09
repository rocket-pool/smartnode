package minipool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get the address of a minipool based on the node address and a salt
func GetExpectedAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, salt *big.Int, opts *bind.CallOpts) (common.Address, error) {
	rocketMinipoolFactory, err := getRocketMinipoolFactory(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	address := new(common.Address)
	if err := rocketMinipoolFactory.Call(opts, address, "getExpectedAddress", nodeAddress, salt); err != nil {
		return common.Address{}, fmt.Errorf("Could not get minipool expected address: %w", err)
	}
	return *address, nil
}

// Get contracts
var rocketMinipoolFactoryLock sync.Mutex

func getRocketMinipoolFactory(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMinipoolFactoryLock.Lock()
	defer rocketMinipoolFactoryLock.Unlock()
	return rp.GetContract("rocketMinipoolFactory", opts)
}
