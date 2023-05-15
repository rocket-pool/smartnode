package minipool

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Create a minipool binding
func NewMinipool(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (Minipool, error) {

	// Get the contract version
	version, err := rocketpool.GetContractVersion(rp, address, opts)
	if err != nil {
		errMsg := err.Error()
		errMsg = strings.ToLower(errMsg)
		if strings.Contains(errMsg, "execution reverted") ||
			strings.Contains(errMsg, "vm execution error") {
			// Reversions happen for minipool v1 on Prater which didn't have version() yet
			version = 1
		} else {
			return nil, fmt.Errorf("error getting minipool contract version: %w", err)
		}
	}

	switch version {
	case 1, 2:
		return newMinipool_v2(rp, address)
	case 3:
		return newMinipool_v3(rp, address, opts)
	default:
		return nil, fmt.Errorf("unexpected minipool contract version [%d]", version)
	}
}

// Create a minipool binding from an explicit version number
func NewMinipoolFromVersion(rp *rocketpool.RocketPool, address common.Address, version uint8, opts *bind.CallOpts) (Minipool, error) {
	switch version {
	case 1, 2:
		return newMinipool_v2(rp, address)
	case 3:
		return newMinipool_v3(rp, address, opts)
	default:
		return nil, fmt.Errorf("unexpected minipool contract version [%d]", version)
	}
}

// Create a minipool contract directly from its ABI, encoded in string form
func createMinipoolContractFromEncodedAbi(rp *rocketpool.RocketPool, address common.Address, encodedAbi string) (*rocketpool.Contract, error) {
	// Decode ABI
	abi, err := rocketpool.DecodeAbi(encodedAbi)
	if err != nil {
		return nil, fmt.Errorf("Could not decode minipool %s ABI: %w", address, err)
	}

	// Create and return
	return &rocketpool.Contract{
		Contract: bind.NewBoundContract(address, *abi, rp.Client, rp.Client, rp.Client),
		Address:  &address,
		ABI:      abi,
		Client:   rp.Client,
	}, nil
}

// Create a minipool contract directly from its ABI
func createMinipoolContractFromAbi(rp *rocketpool.RocketPool, address common.Address, abi *abi.ABI) (*rocketpool.Contract, error) {
	// Create and return
	return &rocketpool.Contract{
		Contract: bind.NewBoundContract(address, *abi, rp.Client, rp.Client, rp.Client),
		Address:  &address,
		ABI:      abi,
		Client:   rp.Client,
	}, nil
}

// Get a minipool contract
var rocketMinipoolLock sync.Mutex

func getMinipoolContract(rp *rocketpool.RocketPool, minipoolAddress common.Address, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketMinipoolLock.Lock()
	defer rocketMinipoolLock.Unlock()
	return rp.MakeContract("rocketMinipool", minipoolAddress, opts)
}
