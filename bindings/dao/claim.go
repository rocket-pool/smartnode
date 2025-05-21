package dao

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

func GetContractExists(rp *rocketpool.RocketPool, contractName string, opts *bind.CallOpts) (bool, error) {
	rocketClaimDAO, err := getRocketClaimDAO(rp, opts)
	if err != nil {
		return false, err
	}
	result := new(bool)
	if err := rocketClaimDAO.Call(opts, result, "getContractExists", contractName); err != nil {
		return false, fmt.Errorf("error checking if contract %s exists: %w", contractName, err)
	}
	return *result, nil
}

// Get contracts
var rocketClaimDAOLock sync.Mutex

func getRocketClaimDAO(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketClaimDAOLock.Lock()
	defer rocketClaimDAOLock.Unlock()
	return rp.GetContract("rocketClaimDAO", opts)
}
