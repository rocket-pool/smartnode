package megapool

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get a megapool deployment state
func GetMegapoolDeployed(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (bool, error) {
	rocketMegapoolFactory, err := getRocketMegapoolFactory(rp, opts)
	if err != nil {
		return false, err
	}
	deployed := false
	if err := rocketMegapoolFactory.Call(opts, deployed, "getMegapoolDeployed", nodeAddress); err != nil {
		return false, fmt.Errorf("error getting megapool deployed for node %s: %w", nodeAddress, err)
	}
	return deployed, nil
}

// Get a megapool expected address
func GetMegapoolExpectedAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (common.Address, error) {
	rocketMegapoolFactory, err := getRocketMegapoolFactory(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	expectedAddress := common.Address{}
	if err := rocketMegapoolFactory.Call(opts, expectedAddress, "getExpectedAddress", nodeAddress); err != nil {
		return common.Address{}, fmt.Errorf("error getting megapool expected address for node %s: %w", nodeAddress, err)
	}
	return expectedAddress, nil
}

// Get a megapool delegate expiration block
func GetMegapoolDelegateExpiry(rp *rocketpool.RocketPool, delegateAddress common.Address, opts *bind.CallOpts) (uint64, error) {
	rocketMegapoolFactory, err := getRocketMegapoolFactory(rp, opts)
	if err != nil {
		return 0, err
	}
	expiryBlock := new(*big.Int)
	if err := rocketMegapoolFactory.Call(opts, expiryBlock, "getDelegateExpiry", delegateAddress); err != nil {
		return 0, fmt.Errorf("error getting expiration block for delegate address %s: %w", delegateAddress, err)
	}
	return (*expiryBlock).Uint64(), nil
}

// Get contracts
var rocketMegapoolFactoryLock sync.Mutex

func getRocketMegapoolFactory(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMegapoolFactoryLock.Lock()
	defer rocketMegapoolFactoryLock.Unlock()
	return rp.GetContract("rocketMegapoolFactory", opts)
}
