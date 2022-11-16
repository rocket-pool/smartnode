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
			"rocketNodeStaking":   "rocketNodeStaking.v2",
		},
		abiMap: map[string]string{
			"rocketNetworkPrices": "eJztVslu2zAQ/ZVCZx+0kKKYW1v0UMAFDKc9BUEwJIeBEJkSSMqNEeTfSyuKFSVektYpjKIHAzI5y3sz5ONc3EWlaVrvorOL9adHa6D6vmowOotkbbwF6T/Ma3mD/tzXFq7x69pIg8RoEhlYrA2v7FODj0pZdC5s+4c40C/cX04i58Hjt9aDKKvSr8KuqU0DKxAVDh4hs/O2lSFgWHTltQHf2uc795O7CIL7alG3gYCGyuFkzEfhLaroLHh0OyN6sMHZ09C2XmxBPXkSaJNjFKkN/1OaD5FEFQoyhHrc/51QtqlmtpR4nGioNUpfLnHeVOcebo4U1peLbZEuNwYdBXfeikXpfQi8scUlGv+sx/FtDimNmS6IykETolLgMqMUmE4wkyTniUhIHtNEIqdxIkjKVUEIE8BplsjsDUfjf0f/pKM/GgWv6SfqVOsiyVBLHn4sZzzN8lwSkJDmjAkWukshyUBrSYniwSOjShbIhWBZT6Hv3gBiidaVtQn56tbvkrE1+GJgNmZV7FClZYk/B0vdmlDjLlEnQBA49qIyJkqJVkTReBfga/QPhfvUn6cDuEc92d6Po2GPNZJYSLUH+3w2fTy6JwScpUQRiPke4F82t2Q2fbwlJ8SASlGknONbGPRX7wQPEtJcJIoVYzaHYV1BUOtOPXbLTmmkRXC4v6GvHjSecBpzELRQBVfZe3BQ+Hc4xDQNrSDxmznsfewO+B563Q64v+Y5GyrpuoHiQVGPfQAgTYjM5L9bvDAPyNb3OvIuNdSMKgYi3yVrpfkchASNa8eZX1AVdV1tk69u/ajalQW8NI7JHiWehlTOz7GprV8X4gQVGMLUlORhLru//AWSfqX7",

			"rocketNodeStaking": "eJzdV8FOAjEQ/RXT855MNIabB01M0BglejCEDN1Zaey2m3YKEuO/OwXcdbMsIkJEbix9ffPem+5k+/QmlCkCedF5ij8JnQHdmxYoOkJaQw4kHd1Z+YJ0T9bBM15FUAYSRSIM5BE4cF8B52nq0HtepjkPLP547yfCExBeB4Kh0oqmvGqsKWAKQ43VDq7syQXJhOI9eRPAoGluA8vMQHtM6qpTfMVUdHjHbKVmAko1C7HGprhEW/KFqKxRYwr8fHxyWjEBKzJUcX0CNuFCGj2ADrgdNlL5MqZ+Cbi77d5r8CPmLGE4Rnaz7bgzZ/P9i3utgAhedp8P2f+ZzqOiUepgYpYFVMZRbRmj88pGtA3UNnFiqbNKR13DWcsAGSucVMgsGEmxUIuOZ6SeJdCfDf5eUC2a5bFsJqpRrXE4BnFYrZynlasbhh60qbQ3P5Z/5ax5hi6yDBk43svcG9VslnlsnRCr92qVqw23ukLfOiVXDxQJWgbNvvc/11+f58N2d62MykN+mN7g9XC9cd8Ka3V38abvjbNGtUHrp05lysf+cJ9qRta+fuwo8zWG5U+8XfLlaJf2NhI3WXwW/sfs+Q42uxSdr9GDiNuCx/4HCbNAvw==",
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
