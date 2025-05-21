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
			"rocketNetworkPrices":   "eJzlVE1v2zAM/SuDzznIimRLuW23AR1QpNupKApKogNjjm1IdNag6H+fkrhx890NHRBgN1skH997lHj/nJR121FIJverT0JfQ/V92WIySWxTkwdLn6aN/Yl0R42HGX5dJRVgMRklNcxXiY/+bcJn5zyGEMO0wYH+4OVhlAQCwm8dgSmrkpYxWjd1C0swFQ4VsXMg39kIGA9DOauBOr8feRk9JxDLl/OmiwIKqAKOdvU4fEKXTGLFOrIjD7Y8exmFb+ZHWI/eAG177CB18Z/LbEAyVTRkgHqN/w2Ub6tbX1r8GDQq58eQHrYJ617hrjPzkigCb3NxgTXtDYM98VyqtDBGaa0BlRijhkwxlaHBnGVKKJSFVpKhcxpAa25S7kCIXI9jZ/sHM/xPrP/ROniH8ZAZNWYmFyi5y12aCoZ5xiUvOJfO5gYdV7niMtVKyIwxY9w4VXwsUmfTVPYSTj39A/aPZ329UHvGyEF+NNN2hBv9Gy9iftNRz/Hdy6PoaktlUx94pguTiULrXe0DgxnSTWwQaIpt42mF+qVXPdC4LPe4ykPuixJ/HWW93nEQh97vrb3Ja2PSLN6Q0yo27l0hd1agYMae4z69vXm9K1dEPOfCCWD6+p5NWK/qf/JeuEanGYdT01qgD5u6C5NSp+akPnZKUhROOMki6m+WhvGX",
			"rocketNetworkBalances": "eJztVk1vm0AQ/SsVZx/4XMC3RoqUSu3Fdk6RFc3uzjooGBAMbtwo/71jm5g4tiGpXIlDfMK7M2/e2zfscPdsJVlRU2WN7zaPhGUG6WxdoDW2VJ5RCYq+TXL1iDSlvIQF/tgEGVBojawMlpvA+/JtwHetS6wq3qYdDjQLL/ORVREQ/qoJZJImtObdLM8KWINMsc3gyhWVtWJAXqySRQZUl+93XkbPFnD6epnXLMBAWuHoUI/GJ9TWmDO2OwfyYM+zkWHKfHmC9egN0L7GAVLN/91AtEgy5QNpoV73/wWKcoL0mh4ug8bH/5hki4vhlUgP07oo0vWF1CZLPIE03wdcQQqZwmpay2VCxND7aFxhRu/axX5CEYQgtIiD0PgGpPRCGQnta153MHK1sYUjtasj8FBjoJ04lLaMIwiA180nuuyrOQbSHLeFhg+0Rij553ih0aB9z9PCEZF0VBio0PFV6MoIJEbaDn1hG1vZkoNs4WEojWO82G9EnLs+j/jfdzrfk9tndU96v7c9AJ1mthawoaom3Dnw6gdn5DU15/ThIWDqTFGSZ0e+BcIYT7jB4fm3HBZIr5WvmhNvy/dLPa3vmPMqwd8n2W5nFHDDNXPnkH3s2LZyPN3B/np2c0tc5g9sECdceGAStAQhwHRI+MmlKppgkZe0sXWARkDMb7/ALiOmzVszu2kaalgKjIMahOd3KJhtr42B8o+Fr23fVX38Jyxgf/cMiL/yFY8H0F+D4HgQVNsPtP80ATxXy8AP7XONs8Ky2uX1dEt0rleiy3ZK4Bv+5AyY8PwvSEB9Dw==",
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
