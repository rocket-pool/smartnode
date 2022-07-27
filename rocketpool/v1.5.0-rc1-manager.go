package rocketpool

import (
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
			"rocketRewardsPool": "eNrtWFFvmzAQ/isTzzy12lTlbc0qLVI3RUneoggZuGSoYCP7nBZV/e87ExpDgISkkco03hJ8/u67s7+7g+WrE/FUo3JGS/MTQXIWL7IUnJETCI6SBfhlJoInwDkKyTYwMUZrFoDjOpwlxtCTZYPvYShBKVrGHQ4rHrytXEchQ/ilkflRHGFGq1zwlGXMj8HuIM8KpQ4I0HlzXx1GRlkiNNFcs1iBW2UdwguEzoh25CuVIDT9v/n6zZKV8MxkODGbrMN3K+MsEEkqOPDmpJyPdwqAyAcaI8HvY0rjZRgmYcCVVt0x/AxB3d5YjATkUwwzIdDufzdq2E8HFPHN4faFBBhPfliEwqxLEPn6ls53ypSiA70oEyiBKS2z2fTxHIDlqgyhFUL4W4TQhEKm3XD4NQAeFj8bAVauvfh7TRwekDbize/nXPtJpBTdMguvys8KD6hTEuKO3lH01tObI5O4iBJoS/9lsA88vB4oNiOt9gZF0jhL1Z+yHmBLdeGjJYntC2ThbS1F0lAu3aG2DbVtqG09qG1n1oucEGL5mG3h2JcJu3sLsmAqNLZNY8brnaVUpXPXMlxtI3i2lmvNAyPGNh4bwFlF8SfoVHLUnJ9rUJo+3rOY8Xzi7AWjKfCQhE/EdulSPSNGuuoZsXHMomRS1EDTx/MpobfseknMto7P4nayT3qBIUyPxsWrY71X1uMrmU9BBv2Lr9y2ahGqwxCLBtYepOocZdlxW6O8UqS1mdQrTQrtb/Qdhhbv+BBZSdXC+iz30COZ8oWIm5KUP/9IfoZJeZiU/5FJ+bzJ2DsyGlfEaDePheb4qXV5UOOgxv9YjbkR1j5KWT12/rI+CGwQ2CCwmsB21xGuq7DVX65rPsI=",
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
func (m *LegacyVersionWrapper_v1_5_0_rc1) GetContract(contractName string) (*Contract, error) {
	return getLegacyContract(m.rp, contractName, m)
}

func (m *LegacyVersionWrapper_v1_5_0_rc1) GetContractWithAddress(contractName string, address common.Address) (*Contract, error) {
	return getLegacyContractWithAddress(m.rp, contractName, address, m)
}
