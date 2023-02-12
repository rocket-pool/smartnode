package state

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Container for network contracts
type NetworkContracts struct {
	RocketNodeManager            *rocketpool.Contract
	RocketNodeStaking            *rocketpool.Contract
	RocketMinipoolManager        *rocketpool.Contract
	RocketNodeDistributorFactory *rocketpool.Contract
	RocketTokenRETH              *rocketpool.Contract
	RocketTokenRPL               *rocketpool.Contract
	RocketTokenRPLFixedSupply    *rocketpool.Contract
	RocketStorage                *rocketpool.Contract
	RocketMinipoolBondReducer    *rocketpool.Contract
}

// Get a new network contracts container
func NewNetworkContracts(rp *rocketpool.RocketPool, isAtlasDeployed bool, opts *bind.CallOpts) (*NetworkContracts, error) {
	contracts := &NetworkContracts{}
	var err error
	contracts.RocketNodeManager, err = rp.GetContract("rocketNodeManager", opts)
	if err != nil {
		return nil, err
	}
	contracts.RocketNodeStaking, err = rp.GetContract("rocketNodeStaking", opts)
	if err != nil {
		return nil, err
	}
	contracts.RocketMinipoolManager, err = rp.GetContract("rocketMinipoolManager", opts)
	if err != nil {
		return nil, err
	}
	contracts.RocketNodeDistributorFactory, err = rp.GetContract("rocketNodeDistributorFactory", opts)
	if err != nil {
		return nil, err
	}
	contracts.RocketTokenRETH, err = rp.GetContract("rocketTokenRETH", opts)
	if err != nil {
		return nil, err
	}
	contracts.RocketTokenRPL, err = rp.GetContract("rocketTokenRPL", opts)
	if err != nil {
		return nil, err
	}
	contracts.RocketTokenRPLFixedSupply, err = rp.GetContract("rocketTokenRPLFixedSupply", opts)
	if err != nil {
		return nil, err
	}
	contracts.RocketStorage = rp.RocketStorageContract

	if isAtlasDeployed {
		contracts.RocketMinipoolBondReducer, err = rp.GetContract("rocketMinipoolBondReducer", opts)
		if err != nil {
			return nil, err
		}
	}

	return contracts, nil
}
