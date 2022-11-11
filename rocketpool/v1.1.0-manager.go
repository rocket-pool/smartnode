package rocketpool

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-version"
)

// A wrapper that holds the updated contract information for this version
type LegacyVersionWrapper_v1_1_0 struct {
	rp              *RocketPool
	rpVersion       *version.Version
	contractNameMap map[string]string
	abiMap          map[string]string
}

// Creates a new wrapper for this version
func newLegacyVersionWrapper_v1_1_0(rp *RocketPool) *LegacyVersionWrapper_v1_1_0 {
	rpVersion, _ := version.NewSemver("1.1.0")
	return &LegacyVersionWrapper_v1_1_0{
		rp:        rp,
		rpVersion: rpVersion,
		contractNameMap: map[string]string{
			"rocketNetworkPrices": "rocketNetworkPrices.v1",
		},
		abiMap: map[string]string{
			"rocketNetworkPrices": "eNrtVsFqwkAQ/ZWy55wKLcVbW3ooWBBtTyIy2UxkMdkNu7O2Iv57JxoTA9HYNkUP3pLszJu3b3beZrwSSmeenOiN80dCqyF5X2YoekIaTRYk3QyNnCONyFiY4WseFINEEQgNaR44tfsBj1Fk0Tlepi0OFB/Wk0A4AsI3TxCqRNGSV7XRGSwhTLDK4MqOrJcMKNbBSgAHLVPjmWYMicOgzjrCL4xEjzM2K7VNQMmmIBtbkzZwC/aAyho1JM/vt3f3FVKY8LYrqN36b6BslgysktgNGsYxSlILHGbJiGDeESyptAlpUgZstuBGPkwVEQOXsbhATT/o5LUBf2nARxbBAflLsaucBVqnjOZo4+mQFeS1HioidRIPByZ7ofCzioy9Zkm40AEeM6Qt/aeiqy10aso0q9IBpeGgvzsXF8HnpTxZg/7uZF0mseIUnrub7dWmwH7E83FssJSWFsHhcflPvt/+k2qEl0H1qDm35La5cUv6KfZbCeY299XWe64aNWvE15T0VMx091LtD9oz//uhdt4dN43QmKTJMTbfO/K0PiM4GmJmLOXbOIeXTb4BvAjxkA==",
		},
	}
}

// Get the version for this manager
func (m *LegacyVersionWrapper_v1_1_0) GetVersion() *version.Version {
	return m.rpVersion
}

// Get the versioned name of the contract if it was upgraded as part of this deployment
func (m *LegacyVersionWrapper_v1_1_0) GetVersionedContractName(contractName string) (string, bool) {
	legacyName, exists := m.contractNameMap[contractName]
	return legacyName, exists
}

// Get the ABI for the provided contract
func (m *LegacyVersionWrapper_v1_1_0) GetEncodedABI(contractName string) string {
	return m.abiMap[contractName]
}

// Get the contract with the provided name for this version of Rocket Pool
func (m *LegacyVersionWrapper_v1_1_0) GetContract(contractName string, opts *bind.CallOpts) (*Contract, error) {
	return getLegacyContract(m.rp, contractName, m, opts)
}

func (m *LegacyVersionWrapper_v1_1_0) GetContractWithAddress(contractName string, address common.Address) (*Contract, error) {
	return getLegacyContractWithAddress(m.rp, contractName, address, m)
}
