package rocketpool

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

const (
	rocketVersionInterfaceABI string = `[
		{
		  "inputs": [],
		  "name": "version",
		  "outputs": [
			{
			  "internalType": "uint8",
			  "name": "",
			  "type": "uint8"
			}
		  ],
		  "stateMutability": "view",
		  "type": "function"
		}
	]`
)

// Get the version of the given contract
func GetContractVersion(rp *RocketPool, contractAddress common.Address, opts *bind.CallOpts) (uint8, error) {

	// Get the version interface ABI
	/*
		abi, err := rp.GetABI("rocketVersionInterface", opts)
		if err != nil {
			return 0, fmt.Errorf("error loading version interface ABI: %w", err)
		}
	*/

	// Parse ABI using the hardcoded string until the contract is deployed
	abiParsed, err := abi.JSON(strings.NewReader(rocketVersionInterfaceABI))
	if err != nil {
		return 0, fmt.Errorf("Could not parse version interface JSON: %w", err)
	}

	// Create contract
	contract := &Contract{
		Contract: bind.NewBoundContract(contractAddress, abiParsed, rp.Client, rp.Client, rp.Client),
		Address:  &contractAddress,
		ABI:      &abiParsed,
		Client:   rp.Client,
	}

	// Get the contract version
	version := new(uint8)
	if err := contract.Call(opts, version, "version"); err != nil {
		return 0, fmt.Errorf("Could not get contract version: %w", err)
	}

	return *version, nil
}
