package megapool

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
)

// Returns true if this megapool always uses the latest delegate contract
func GetUseLatestDelegate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	megapoolProxy, err := getRocketMegapoolProxy(rp, opts)
	if err != nil {
		return false, err
	}
	isUsingLatestDelegate := new(bool)
	if err := megapoolProxy.Call(opts, isUsingLatestDelegate, "getUseLatestDelegate"); err != nil {
		return false, fmt.Errorf("error checking if this megapool is using the latest delegate:, %w", err)
	}
	return *isUsingLatestDelegate, nil
}

// Returns the address of the megapool's stored delegate
func GetDelegate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (common.Address, error) {
	megapoolProxy, err := getRocketMegapoolProxy(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	delegateAddress := new(common.Address)
	if err := megapoolProxy.Call(opts, delegateAddress, "getDelegate"); err != nil {
		return common.Address{}, fmt.Errorf("error getting the delegate address: %w", err)
	}
	return *delegateAddress, nil
}

// Returns the delegate which will be used when calling this megapool taking into account useLatestDelegate setting
func GetEffectiveDelegate(rp *rocketpool.RocketPool, opts *bind.CallOpts) (common.Address, error) {
	megapoolProxy, err := getRocketMegapoolProxy(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	effectiveDelegateAddress := new(common.Address)
	if err := megapoolProxy.Call(opts, effectiveDelegateAddress, "getEffectiveDelegate"); err != nil {
		return common.Address{}, fmt.Errorf("error getting the effective delegate address: %w", err)
	}
	return *effectiveDelegateAddress, nil
}

// Returns true if the megapools current delegate has expired
func GetDelegateExpired(rp *rocketpool.RocketPool, opts *bind.CallOpts) (bool, error) {
	megapoolProxy, err := getRocketMegapoolProxy(rp, opts)
	if err != nil {
		return false, err
	}
	delegateExpired := new(bool)
	if err := megapoolProxy.Call(opts, delegateExpired, "getDelegateExpired"); err != nil {
		return false, fmt.Errorf("error checking if the megapool's delegate has expired:, %w", err)
	}
	return *delegateExpired, nil
}

// Get contracts
var rocketMegapoolProxyLock sync.Mutex

func getRocketMegapoolProxy(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMegapoolProxyLock.Lock()
	defer rocketMegapoolProxyLock.Unlock()
	return rp.GetContract("rocketMegapoolProxy", opts)
}
