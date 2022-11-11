package network

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Estimate the gas of SubmitPenalty
func EstimateSubmitPenaltyGas(rp *rocketpool.RocketPool, minipoolAddress common.Address, block *big.Int, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	rocketNetworkPenalties, err := getRocketNetworkPenalties(rp, nil)
	if err != nil {
		return rocketpool.GasInfo{}, err
	}
	return rocketNetworkPenalties.GetTransactionGasInfo(opts, "submitPenalty", minipoolAddress, block)
}

// Submit penalty for given minipool
func SubmitPenalty(rp *rocketpool.RocketPool, minipoolAddress common.Address, block *big.Int, opts *bind.TransactOpts) (common.Hash, error) {
	rocketNetworkPrices, err := getRocketNetworkPenalties(rp, nil)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := rocketNetworkPrices.Transact(opts, "submitPenalty", minipoolAddress, block)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not submit penalty: %w", err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketNetworkPenaltiesLock sync.Mutex

func getRocketNetworkPenalties(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNetworkPenaltiesLock.Lock()
	defer rocketNetworkPenaltiesLock.Unlock()
	return rp.GetContract("rocketNetworkPenalties", opts)
}
