package rocketpool

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

const (
	rocketVersionInterfaceAbiString string = `[
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

var versionAbi *abi.ABI

// Get the version of the given contract
func GetContractVersion(rp *RocketPool, contractAddress common.Address, opts *bind.CallOpts) (uint8, error) {
	if versionAbi == nil {
		// Parse ABI using the hardcoded string until the contract is deployed
		abiParsed, err := abi.JSON(strings.NewReader(rocketVersionInterfaceAbiString))
		if err != nil {
			return 0, fmt.Errorf("Could not parse version interface JSON: %w", err)
		}
		versionAbi = &abiParsed
	}

	// Create contract
	contract := &Contract{
		Contract: bind.NewBoundContract(contractAddress, *versionAbi, rp.Client, rp.Client, rp.Client),
		Address:  &contractAddress,
		ABI:      versionAbi,
		Client:   rp.Client,
	}

	// Get the contract version
	version := new(uint8)
	if err := contract.Call(opts, version, "version"); err != nil {
		return 0, fmt.Errorf("Could not get contract version: %w", err)
	}

	return *version, nil
}

// Get the rocketVersion contract binding at the given address
func GetRocketVersionContractForAddress(rp *RocketPool, address common.Address) (*Contract, error) {
	if versionAbi == nil {
		// Parse ABI using the hardcoded string until the contract is deployed
		abiParsed, err := abi.JSON(strings.NewReader(rocketVersionInterfaceAbiString))
		if err != nil {
			return nil, fmt.Errorf("Could not parse version interface JSON: %w", err)
		}
		versionAbi = &abiParsed
	}

	return &Contract{
		Contract: bind.NewBoundContract(address, *versionAbi, rp.Client, rp.Client, rp.Client),
		Address:  &address,
		ABI:      versionAbi,
		Client:   rp.Client,
	}, nil
}
