package minipool

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get the CreationCode binary for the RocketMinipool contract that will be created by node deposits
func GetMinipoolBytecode(rp *rocketpool.RocketPool, opts *bind.CallOpts, legacyRocketMinipoolFactoryAddress *common.Address) ([]byte, error) {
	rocketMinipoolFactory, err := getRocketMinipoolFactory(rp, legacyRocketMinipoolFactoryAddress, opts)
	if err != nil {
		return []byte{}, err
	}
	bytecode := new([]byte)
	if err := rocketMinipoolFactory.Call(opts, bytecode, "getMinipoolBytecode"); err != nil {
		return []byte{}, fmt.Errorf("Could not get minipool contract bytecode: %w", err)
	}
	return *bytecode, nil
}

// Get contracts
var rocketMinipoolFactoryLock sync.Mutex

func getRocketMinipoolFactory(rp *rocketpool.RocketPool, address *common.Address, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMinipoolFactoryLock.Lock()
	defer rocketMinipoolFactoryLock.Unlock()
	if address == nil {
		return rp.VersionManager.V1_1_0.GetContract("rocketMinipoolFactory", opts)
	} else {
		return rp.VersionManager.V1_1_0.GetContractWithAddress("rocketMinipoolFactory", *address)
	}
}
