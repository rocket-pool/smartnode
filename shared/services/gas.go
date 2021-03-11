package services

import (
    "github.com/ethereum/go-ethereum/accounts/abi/bind"

    "github.com/rocket-pool/rocketpool-go/rocketpool"
)


// Get Gas Price and Gas Limit for transaction
func GetGasInfo(rp *rocketpool.RocketPool, opts *bind.TransactOpts, contractName, methodName string, params ...interface{}) (rocketpool.GasInfo, error) {

    contract, err := rp.GetContract(contractName)
    if err != nil { 
        return rocketpool.GasInfo{}, err
    }

    return contract.GetGasInfo(methodName, opts, params...)
}

