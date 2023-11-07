package rocketpool

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-version"
)

// A wrapper that holds the updated contract information for this version
type LegacyVersionWrapper_v1_2_0 struct {
	rp              *RocketPool
	rpVersion       *version.Version
	contractNameMap map[string]string
	abiMap          map[string]string
}

// Creates a new wrapper for this version
func newLegacyVersionWrapper_v1_2_0(rp *RocketPool) *LegacyVersionWrapper_v1_2_0 {
	rpVersion, _ := version.NewSemver("1.2.0")
	return &LegacyVersionWrapper_v1_2_0{
		rp:        rp,
		rpVersion: rpVersion,
		contractNameMap: map[string]string{
			"rocketNetworkPrices":   "rocketNetworkPrices.v2",
			"rocketNetworkBalances": "rocketNetworkBalances.v2",
		},
		abiMap: map[string]string{
			"rocketNetworkPrices":   "",
			"rocketNetworkBalances": "",
		},
	}
}

// Get the version for this manager
func (m *LegacyVersionWrapper_v1_2_0) GetVersion() *version.Version {
	return m.rpVersion
}

// Get the versioned name of the contract if it was upgraded as part of this deployment
func (m *LegacyVersionWrapper_v1_2_0) GetVersionedContractName(contractName string) (string, bool) {
	legacyName, exists := m.contractNameMap[contractName]
	return legacyName, exists
}

// Get the ABI for the provided contract
func (m *LegacyVersionWrapper_v1_2_0) GetEncodedABI(contractName string) string {
	return m.abiMap[contractName]
}

// Get the contract with the provided name for this version of Rocket Pool
func (m *LegacyVersionWrapper_v1_2_0) GetContract(contractName string, opts *bind.CallOpts) (*Contract, error) {
	return getLegacyContract(m.rp, contractName, m, opts)
}

func (m *LegacyVersionWrapper_v1_2_0) GetContractWithAddress(contractName string, address common.Address) (*Contract, error) {
	return getLegacyContractWithAddress(m.rp, contractName, address, m)
}
