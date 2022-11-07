package node

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Distributor contract
type Distributor struct {
	Address    common.Address
	Contract   *rocketpool.Contract
	RocketPool *rocketpool.RocketPool
}

// Create new distributor contract
func NewDistributor(rp *rocketpool.RocketPool, address common.Address, opts *bind.CallOpts) (*Distributor, error) {

	// Get contract
	contract, err := getDistributorContract(rp, address, opts)
	if err != nil {
		return nil, err
	}

	// Create and return
	return &Distributor{
		Address:    address,
		Contract:   contract,
		RocketPool: rp,
	}, nil
}

// Gets the deterministic address for a node's reward distributor contract
func GetDistributorAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (common.Address, error) {
	rocketNodeDistributorFactory, err := getRocketNodeDistributorFactory(rp, opts)
	if err != nil {
		return common.Address{}, err
	}
	var address common.Address
	if err := rocketNodeDistributorFactory.Call(opts, &address, "getProxyAddress", nodeAddress); err != nil {
		return common.Address{}, fmt.Errorf("Could not get distributor address: %w", err)
	}
	return address, nil
}

// Estimate the gas of a distribute
func (d *Distributor) EstimateDistributeGas(opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return d.Contract.GetTransactionGasInfo(opts, "distribute")
}

// Distribute the contract's balance to the rETH contract and the user
func (d *Distributor) Distribute(opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := d.Contract.Transact(opts, "distribute")
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not distribute fee distributor balance: %w", err)
	}
	return tx.Hash(), nil
}

// Get contracts
var rocketNodeDistributorFactoryLock sync.Mutex

func getRocketNodeDistributorFactory(rp *rocketpool.RocketPool, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketNodeDistributorFactoryLock.Lock()
	defer rocketNodeDistributorFactoryLock.Unlock()
	return rp.GetContract("rocketNodeDistributorFactory", opts)
}

// Get a distributor contract
var rocketDistributorLock sync.Mutex

func getDistributorContract(rp *rocketpool.RocketPool, distributorAddress common.Address, opts *bind.CallOpts) (*rocketpool.Contract, error) {
	rocketDistributorLock.Lock()
	defer rocketDistributorLock.Unlock()
	return rp.MakeContract("rocketNodeDistributorDelegate", distributorAddress, opts)
}
