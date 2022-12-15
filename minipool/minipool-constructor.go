package minipool

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func NewMinipool(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (Minipool, error) {

	// Get the contract version
	version, err := rocketpool.GetContractVersion(rp, address, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting minipool contract version: %w", err)
	}

	switch version {
	case 1, 2:
		return newMinipool_v2(rp, address, opts)
	case 3:
		return newMinipool_v3(rp, address, opts)
	default:
		return nil, fmt.Errorf("unexpected minipool contract version [%d]", version)
	}
}

// Get a minipool contract
var rocketMinipoolLock sync.Mutex

func getMinipoolContract(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMinipoolLock.Lock()
	defer rocketMinipoolLock.Unlock()
	return rp.MakeContract("rocketMinipool", minipoolAddress, opts)
}
