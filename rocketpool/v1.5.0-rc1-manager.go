package rocketpool

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-version"
)

// A wrapper that holds the updated contract information for this version
type LegacyVersionWrapper_v1_5_0_rc1 struct {
	rp              *RocketPool
	rpVersion       *version.Version
	contractNameMap map[string]string
	abiMap          map[string]string
}

// Creates a new wrapper for this version
func newLegacyVersionWrapper_v1_5_0_rc1(rp *RocketPool) *LegacyVersionWrapper_v1_5_0_rc1 {
	rpVersion, _ := version.NewSemver("1.5.0-rc1")
	return &LegacyVersionWrapper_v1_5_0_rc1{
		rp:        rp,
		rpVersion: rpVersion,
		contractNameMap: map[string]string{
			"rocketRewardsPool": "rocketRewardsPool.v2",
		},
		abiMap: map[string]string{
			"rocketRewardsPool": "eJztWdtu2zgQ/ZWFnv2gC6lL3rbZAhugWxiJ3wLDGJHDVKhMGSSV1ij676UUNZJs+ZZ1ARXVk21yeDjnjGY4lB+/OZnclEY7N4/VV4NKQr7YbtC5cVghjQJm/rov2Gc0D6ZQ8IR3lZEAhs7MkbCuDFeqa/A35wq1ttPmBQeage/LmaMNGPyvNJBmeWa2dlYWcgNbSHNsV9idtVEls4B2UGdPEkypdme+z745YJdv10VpCQjINc76fDh+Re7c2BX1TI9eaX/7NGxpKPwCit9Vi1pXflpVm7FivSkkymG5Lsc7BWCdZ6XJCvkutwK/DaMSDKUu9fkY6dagDvwWY43qc473RWHa9T+NBtbbAGXyaXf5QiHe3v3TIjRm55Co559tfOegtQ3om5QwCkGXans//3AJwOOyC1Fqg/xjwXEIxZqehyOvAfB+8e8gwHLWPvivObEboLJK6/r5fCjTdaa1fcpaeN0da3Yw5cam6It7R9EPRu/BgDKLbI2H5H8b7HvJrwdqhpGWrwaNaBI2+lM3H/DZ1oWdYuV+5SkPOKaEhC53eYiYMkIJJiKlHjKG3Is540lkPykK4XHqh24EFHjIYzdM/m+Ng9da3LgvVLEeqMyzqVhOxXIqliMolhcWoNohY7phHq5EiUiCIGEh8xM3JIS5AefCVibBiE8x4BEVGIacB7EgFKkPwL3IgwTQDVLfTxoyTd1p3XlG1VAvSnOok6xoxC3HPr/4QGP4nOGX1lKUkpmXjeoeECzHplT1iVIiOOHUPeTwE5r7Xq054XcvOsORuZrvwveYR112zPf5h3eQg6z77xG57ntJjKF/zPU5Sm6LmGXwEgA9LgaRTQThYXSagS0mo2SQxl7KgXpHGNzmkK3vmhOiapvqpmxkNCAgLCXpJTTGxYC64IsI6bkM2qN6RCRcEsRRxHcK6ckGZsUqZnbotnl9sN/E7AvRMZ+jYuMSQoQQiBDFWUJ0G489KfSuFk0LclgNfbYc3Y0PtTpXk4Rxe8iix05Isnf9WHWawsPvic7oT1fH7ws9TRftnt126YikaVHkQ2rW49c9Nl2CPriwK+R0e5puT7/J7emy29LqyHWpl7Xt4tuilCNrUWgCjARs70iY0nZK2z84bWsjs/eKtE3cs/8B6mRiP/MCFgL47pR5U+ZNmbfceRzxl6YeCUgc8foo/QHwlCkA",
		},
	}
}

// Get the version for this manager
func (m *LegacyVersionWrapper_v1_5_0_rc1) GetVersion() *version.Version {
	return m.rpVersion
}

// Get the versioned name of the contract if it was upgraded as part of this deployment
func (m *LegacyVersionWrapper_v1_5_0_rc1) GetVersionedContractName(contractName string) (string, bool) {
	legacyName, exists := m.contractNameMap[contractName]
	return legacyName, exists
}

// Get the ABI for the provided contract
func (m *LegacyVersionWrapper_v1_5_0_rc1) GetEncodedABI(contractName string) string {
	return m.abiMap[contractName]
}

// Get the contract with the provided name for this version of Rocket Pool
func (m *LegacyVersionWrapper_v1_5_0_rc1) GetContract(contractName string, opts *bind.CallOpts) (*Contract, error) {
	return getLegacyContract(m.rp, contractName, m, opts)
}

func (m *LegacyVersionWrapper_v1_5_0_rc1) GetContractWithAddress(contractName string, address common.Address) (*Contract, error) {
	return getLegacyContractWithAddress(m.rp, contractName, address, m)
}
