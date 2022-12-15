package rocketpool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Get the version of the given contract
func GetContractVersion(rp *RocketPool, contractAddress common.Address, opts *bind.CallOpts) (uint8, error) {

	// Get the version interface ABI
	abi, err := rp.GetABI("rocketVersionInterface", opts)
	if err != nil {
		return 0, fmt.Errorf("error loading version interface ABI: %w", err)
	}

	// Create contract
	contract := &Contract{
		Contract: bind.NewBoundContract(contractAddress, *abi, rp.Client, rp.Client, rp.Client),
		Address:  &contractAddress,
		ABI:      abi,
		Client:   rp.Client,
	}

	// Get the contract version
	version := new(uint8)
	if err := contract.Call(opts, version, "version"); err != nil {
		return 0, fmt.Errorf("Could not get contract version: %w", err)
	}

	return *version, nil
}
