package node

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Gets the deterministic address for a node's reward distributor contract
func GetDistributorAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (common.Address, error) {
	rocketNodeDistributorFactory, err := getRocketNodeDistributorFactory(rp)
	if err != nil {
		return common.Address{}, err
	}
	var address common.Address
	if err := rocketNodeDistributorFactory.Call(opts, &address, "getProxyAddress", nodeAddress); err != nil {
		return common.Address{}, fmt.Errorf("Could not get distributor address: %w", err)
	}
	return address, nil
}

// Get contracts
var rocketNodeDistributorFactoryLock sync.Mutex

func getRocketNodeDistributorFactory(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
	rocketNodeDistributorFactoryLock.Lock()
	defer rocketNodeDistributorFactoryLock.Unlock()
	return rp.GetContract("rocketNodeDistributorFactory")
}
