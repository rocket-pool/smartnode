package rocketpool

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/go-version"
)

// Wrapper for legacy contract versions
type LegacyVersionWrapper interface {
	GetVersion() *version.Version
	GetVersionedContractName(contractName string) (string, bool)
	GetEncodedABI(contractName string) string
	GetContract(contractName string, opts *bind.CallOpts) (*Contract, error)
	GetContractWithAddress(contractName string, address common.Address) (*Contract, error)
}

type VersionManager struct {
	V1_0_0     LegacyVersionWrapper
	V1_5_0_RC1 LegacyVersionWrapper

	rp *RocketPool
}

func NewVersionManager(rp *RocketPool) *VersionManager {
	return &VersionManager{
		V1_0_0:     newLegacyVersionWrapper_v1_0_0(rp),
		V1_5_0_RC1: newLegacyVersionWrapper_v1_5_0_rc1(rp),
		rp:         rp,
	}
}

// Get the contract with the provided name and version wrapper
func getLegacyContract(rp *RocketPool, contractName string, m LegacyVersionWrapper, opts *bind.CallOpts) (*Contract, error) {

	legacyName, exists := m.GetVersionedContractName(contractName)
	if !exists {
		// This wasn't upgraded in previous versions
		return rp.GetContract(contractName, opts)
	}

	// Check for cached contract
	if cached, ok := rp.getCachedContract(legacyName); ok {
		if time.Now().Unix()-cached.time <= CacheTTL {
			return cached.contract, nil
		} else {
			rp.deleteCachedContract(legacyName)
		}
	}

	// Try to get the legacy address from RocketStorage first
	emptyAddress := common.Address{}
	address, err := rp.RocketStorage.GetAddress(nil, crypto.Keccak256Hash([]byte("contract.address"), []byte(legacyName)))
	if err != nil {
		return nil, fmt.Errorf("Could not load v%s contract %s address: %w", m.GetVersion().String(), contractName, err)
	}

	if address == emptyAddress {
		// Not found, so the legacy contract is still on the network - try loading the original contract name instead
		return rp.GetContract(contractName, opts)
	}

	// If we're here, we have a legacy contract
	abiEncoded := m.GetEncodedABI(contractName)
	abi, err := DecodeAbi(abiEncoded)
	if err != nil {
		return nil, fmt.Errorf("Could not decode contract %s ABI: %w", contractName, err)
	}

	contract := &Contract{
		Contract: bind.NewBoundContract(address, *abi, rp.Client, rp.Client, rp.Client),
		Address:  &address,
		ABI:      abi,
		Client:   rp.Client,
	}

	return contract, nil

}

// Get the contract with the provided name, address, and version wrapper
func getLegacyContractWithAddress(rp *RocketPool, contractName string, address common.Address, m LegacyVersionWrapper) (*Contract, error) {

	abiEncoded := m.GetEncodedABI(contractName)
	abi, err := DecodeAbi(abiEncoded)
	if err != nil {
		return nil, fmt.Errorf("Could not decode contract %s ABI: %w", contractName, err)
	}

	contract := &Contract{
		Contract: bind.NewBoundContract(address, *abi, rp.Client, rp.Client, rp.Client),
		Address:  &address,
		ABI:      abi,
		Client:   rp.Client,
	}

	return contract, nil

}
