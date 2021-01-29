package rocketpool

import (
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
)


// Contract type wraps go-ethereum bound contract
type Contract struct {
    Contract *bind.BoundContract
    Address *common.Address
    ABI *abi.ABI
}


// Call a contract method
func (c *Contract) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
    return c.Contract.Call(opts, result, method, params...)
}

