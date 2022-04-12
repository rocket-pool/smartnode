package network

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Estimate the gas of SubmitPenalty
func EstimateSubmitPenaltyGas(rp *rocketpool.RocketPool, minipoolAddress common.Address, block uint64, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkPenalties, err := getRocketNetworkPenalties(rp)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkPenalties.GetTransactionGasInfo(opts, "submitPenalty", minipoolAddress, block)
}

// Submit penalty for given minipool
func SubmitPenalty(rp *rocketpool.RocketPool, minipoolAddress common.Address, block uint64, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkPrices, err := getRocketNetworkPenalties(rp)
	if err != nil {
		return common.Hash{}, err
	}
	hash, err := rocketNetworkPrices.Transact(opts, "submitPenalty", minipoolAddress, block)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not submit penalty: %w", err)
	}
	return hash, nil
}

// Get contracts
var rocketNetworkPenaltiesLock sync.Mutex

func getRocketNetworkPenalties(rp *rocketpool.RocketPool) (*rocketpool.Contract, error) {
	rocketNetworkPenaltiesLock.Lock()
	defer rocketNetworkPenaltiesLock.Unlock()
	return rp.GetContract("rocketNetworkPenalties")
}
