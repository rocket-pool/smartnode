package storage

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/rocketpool"
)

// Get a node's withdrawal address
func GetNodeWithdrawalAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (common.Address, error) {
	withdrawalAddress := new(common.Address)
	if err := rp.RocketStorageContract.Call(opts, withdrawalAddress, "getNodeWithdrawalAddress", nodeAddress); err != nil {
		return common.Address{}, fmt.Errorf("Could not get node %s withdrawal address: %w", nodeAddress.Hex(), err)
	}
	return *withdrawalAddress, nil
}

// Get a node's pending withdrawal address
func GetNodePendingWithdrawalAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.CallOpts) (common.Address, error) {
	withdrawalAddress := new(common.Address)
	if err := rp.RocketStorageContract.Call(opts, withdrawalAddress, "getNodePendingWithdrawalAddress", nodeAddress); err != nil {
		return common.Address{}, fmt.Errorf("Could not get node %s pending withdrawal address: %w", nodeAddress.Hex(), err)
	}
	return *withdrawalAddress, nil
}

// Estimate the gas of SetWithdrawalAddress
func EstimateSetWithdrawalAddressGas(rp *rocketpool.RocketPool, nodeAddress common.Address, withdrawalAddress common.Address, confirm bool, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return rp.RocketStorageContract.GetTransactionGasInfo(opts, "setWithdrawalAddress", nodeAddress, withdrawalAddress, confirm)
}

// Set a node's withdrawal address
func SetWithdrawalAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, withdrawalAddress common.Address, confirm bool, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := rp.RocketStorageContract.Transact(opts, "setWithdrawalAddress", nodeAddress, withdrawalAddress, confirm)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not set node withdrawal address: %w", err)
	}
	return tx.Hash(), nil
}

// Estimate the gas of ConfirmWithdrawalAddress
func EstimateConfirmWithdrawalAddressGas(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.TransactOpts) (rocketpool.GasInfo, error) {
	return rp.RocketStorageContract.GetTransactionGasInfo(opts, "confirmWithdrawalAddress", nodeAddress)
}

// Set a node's withdrawal address
func ConfirmWithdrawalAddress(rp *rocketpool.RocketPool, nodeAddress common.Address, opts *bind.TransactOpts) (common.Hash, error) {
	tx, err := rp.RocketStorageContract.Transact(opts, "confirmWithdrawalAddress", nodeAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("Could not confirm node withdrawal address: %w", err)
	}
	return tx.Hash(), nil
}
