package config

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/config"
)

// A collection of network-specific resources and getters for them
type RocketPoolResources struct {
	*config.NetworkResources

	// The URL to use for staking rETH
	StakeUrl string

	// The contract address of RocketStorage
	StorageAddress common.Address

	// The contract address of rETH
	RethAddress common.Address

	// The contract address of the RPL token
	RplTokenAddress common.Address

	// The contract address of rocketRewardsPool from v1.0.0
	V1_0_0_RewardsPoolAddress *common.Address

	// The contract address of rocketClaimNode from v1.0.0
	V1_0_0_ClaimNodeAddress *common.Address

	// The contract address of rocketClaimTrustedNode from v1.0.0
	V1_0_0_ClaimTrustedNodeAddress *common.Address

	// The contract address of rocketMinipoolManager from v1.0.0
	V1_0_0_MinipoolManagerAddress *common.Address

	// The contract address of rocketNetworkPrices from v1.1.0
	V1_1_0_NetworkPricesAddress *common.Address

	// The contract address of rocketNodeStaking from v1.1.0
	V1_1_0_NodeStakingAddress *common.Address

	// The contract address of rocketNodeDeposit from v1.1.0
	V1_1_0_NodeDepositAddress *common.Address

	// The contract address of rocketMinipoolQueue from v1.1.0
	V1_1_0_MinipoolQueueAddress *common.Address

	// The contract address of rocketMinipoolFactory from v1.1.0
	V1_1_0_MinipoolFactoryAddress *common.Address

	// The contract address of rocketNetworkPrices from v1.2.0
	V1_2_0_NetworkPricesAddress *common.Address

	// The contract address of rocketNetworkBalances from v1.2.0
	V1_2_0_NetworkBalancesAddress *common.Address

	// The contract address for Snapshot delegation
	SnapshotDelegationAddress *common.Address

	// The Snapshot API domain
	SnapshotApiDomain string

	// Addresses for RocketRewardsPool that have been upgraded during development
	PreviousRewardsPoolAddresses []common.Address

	// Addresses for RocketDAOProtocolVerifier that have been upgraded during development
	PreviousProtocolDaoVerifierAddresses []common.Address

	// Addresses for RocketNetworkPrices that have been upgraded during development
	PreviousRocketNetworkPricesAddresses []common.Address

	// Addresses for RocketNetworkBalances that have been upgraded during development
	PreviousRocketNetworkBalancesAddresses []common.Address

	// The RocketOvmPriceMessenger Optimism address for each network
	OptimismPriceMessengerAddress *common.Address

	// The RocketPolygonPriceMessenger Polygon address for each network
	PolygonPriceMessengerAddress *common.Address

	// The RocketArbitumPriceMessenger Arbitrum address for each network
	ArbitrumPriceMessengerAddress *common.Address

	// The RocketArbitumPriceMessengerV2 Arbitrum address for each network
	ArbitrumPriceMessengerAddressV2 *common.Address

	// The RocketZkSyncPriceMessenger zkSyncEra address for each network
	ZkSyncEraPriceMessengerAddress *common.Address

	// The RocketBasePriceMessenger Base address for each network
	BasePriceMessengerAddress *common.Address

	// The RocketScrollPriceMessenger Scroll address for each network
	ScrollPriceMessengerAddress *common.Address

	// The Scroll L2 message fee estimator address for each network
	ScrollFeeEstimatorAddress *common.Address

	// The UniswapV3 pool address for each network (used for RPL price TWAP info)
	RplTwapPoolAddress *common.Address
}

// Creates a new resource collection for the given network
func newRocketPoolResources(network config.Network) *RocketPoolResources {
	// Mainnet
	mainnetResources := &RocketPoolResources{
		NetworkResources:               config.NewResources(config.Network_Mainnet),
		StakeUrl:                       "https://stake.rocketpool.net",
		StorageAddress:                 common.HexToAddress("0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46"),
		RethAddress:                    common.HexToAddress("0xae78736Cd615f374D3085123A210448E74Fc6393"),
		RplTokenAddress:                common.HexToAddress("0xD33526068D116cE69F19A9ee46F0bd304F21A51f"),
		V1_0_0_RewardsPoolAddress:      hexToAddressPtr("0xA3a18348e6E2d3897B6f2671bb8c120e36554802"),
		V1_0_0_ClaimNodeAddress:        hexToAddressPtr("0x899336A2a86053705E65dB61f52C686dcFaeF548"),
		V1_0_0_ClaimTrustedNodeAddress: hexToAddressPtr("0x6af730deB0463b432433318dC8002C0A4e9315e8"),
		V1_0_0_MinipoolManagerAddress:  hexToAddressPtr("0x6293B8abC1F36aFB22406Be5f96D893072A8cF3a"),
		V1_1_0_NetworkPricesAddress:    hexToAddressPtr("0xd3f500F550F46e504A4D2153127B47e007e11166"),
		V1_1_0_NodeStakingAddress:      hexToAddressPtr("0xA73ec45Fe405B5BFCdC0bF4cbc9014Bb32a01cd2"),
		V1_1_0_NodeDepositAddress:      hexToAddressPtr("0x1Cc9cF5586522c6F483E84A19c3C2B0B6d027bF0"),
		V1_1_0_MinipoolQueueAddress:    hexToAddressPtr("0x5870dA524635D1310Dc0e6F256Ce331012C9C19E"),
		V1_1_0_MinipoolFactoryAddress:  hexToAddressPtr("0x54705f80D7C51Fcffd9C659ce3f3C9a7dCCf5788"),
		V1_2_0_NetworkPricesAddress:    hexToAddressPtr("0x751826b107672360b764327631cC5764515fFC37"),
		V1_2_0_NetworkBalancesAddress:  hexToAddressPtr("0x07FCaBCbe4ff0d80c2b1eb42855C0131b6cba2F4"),
		SnapshotDelegationAddress:      hexToAddressPtr("0x469788fE6E9E9681C6ebF3bF78e7Fd26Fc015446"),
		SnapshotApiDomain:              "hub.snapshot.org",
		PreviousRewardsPoolAddresses: []common.Address{
			common.HexToAddress("0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1"),
		},
		PreviousProtocolDaoVerifierAddresses: []common.Address{},
		PreviousRocketNetworkPricesAddresses: []common.Address{
			common.HexToAddress("0x751826b107672360b764327631cC5764515fFC37"),
		},
		PreviousRocketNetworkBalancesAddresses: []common.Address{
			common.HexToAddress("0x07FCaBCbe4ff0d80c2b1eb42855C0131b6cba2F4"),
		},
		OptimismPriceMessengerAddress:   hexToAddressPtr("0xdddcf2c25d50ec22e67218e873d46938650d03a7"),
		PolygonPriceMessengerAddress:    hexToAddressPtr("0xb1029Ac2Be4e08516697093e2AFeC435057f3511"),
		ArbitrumPriceMessengerAddress:   hexToAddressPtr("0x05330300f829AD3fC8f33838BC88CFC4093baD53"),
		ArbitrumPriceMessengerAddressV2: hexToAddressPtr("0x312FcFB03eC9B1Ea38CB7BFCd26ee7bC3b505aB1"),
		ZkSyncEraPriceMessengerAddress:  hexToAddressPtr("0x6cf6CB29754aEBf88AF12089224429bD68b0b8c8"),
		BasePriceMessengerAddress:       hexToAddressPtr("0x64A5856869C06B0188C84A5F83d712bbAc03517d"),
		ScrollPriceMessengerAddress:     hexToAddressPtr("0x0f22dc9b9c03757d4676539203d7549c8f22c15c"),
		ScrollFeeEstimatorAddress:       hexToAddressPtr("0x0d7E906BD9cAFa154b048cFa766Cc1E54E39AF9B"),
		RplTwapPoolAddress:              hexToAddressPtr("0xe42318ea3b998e8355a3da364eb9d48ec725eb45"),
	}

	// Holesky
	holeskyResources := &RocketPoolResources{
		NetworkResources:                     config.NewResources(config.Network_Holesky),
		StakeUrl:                             "https://testnet.rocketpool.net",
		StorageAddress:                       common.HexToAddress("0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1"),
		RethAddress:                          common.HexToAddress("0x7322c24752f79c05FFD1E2a6FCB97020C1C264F1"),
		RplTokenAddress:                      common.HexToAddress("0x1Cc9cF5586522c6F483E84A19c3C2B0B6d027bF0"),
		V1_0_0_RewardsPoolAddress:            nil,
		V1_0_0_ClaimNodeAddress:              nil,
		V1_0_0_ClaimTrustedNodeAddress:       nil,
		V1_0_0_MinipoolManagerAddress:        nil,
		V1_1_0_NetworkPricesAddress:          nil,
		V1_1_0_NodeStakingAddress:            nil,
		V1_1_0_NodeDepositAddress:            nil,
		V1_1_0_MinipoolQueueAddress:          nil,
		V1_1_0_MinipoolFactoryAddress:        nil,
		V1_2_0_NetworkPricesAddress:          hexToAddressPtr("0x029d946F28F93399a5b0D09c879FC8c94E596AEb"),
		V1_2_0_NetworkBalancesAddress:        hexToAddressPtr("0x9294Fc6F03c64Cc217f5BE8697EA3Ed2De77e2F8"),
		SnapshotDelegationAddress:            nil,
		SnapshotApiDomain:                    "",
		PreviousRewardsPoolAddresses:         []common.Address{},
		PreviousProtocolDaoVerifierAddresses: []common.Address{},
		PreviousRocketNetworkPricesAddresses: []common.Address{
			common.HexToAddress("0x029d946f28f93399a5b0d09c879fc8c94e596aeb"),
		},
		PreviousRocketNetworkBalancesAddresses: []common.Address{
			common.HexToAddress("0x9294Fc6F03c64Cc217f5BE8697EA3Ed2De77e2F8"),
		},
		OptimismPriceMessengerAddress:   nil,
		PolygonPriceMessengerAddress:    nil,
		ArbitrumPriceMessengerAddress:   nil,
		ArbitrumPriceMessengerAddressV2: nil,
		ZkSyncEraPriceMessengerAddress:  nil,
		BasePriceMessengerAddress:       nil,
		ScrollPriceMessengerAddress:     nil,
		ScrollFeeEstimatorAddress:       nil,
		RplTwapPoolAddress:              hexToAddressPtr("0x7bb10d2a3105ed5cc150c099a06cafe43d8aa15d"),
	}

	// Devnet
	devnetResources := &RocketPoolResources{
		NetworkResources:               config.NewResources(config.Network_Holesky),
		StakeUrl:                       "TBD",
		StorageAddress:                 common.HexToAddress("0xf04de123993761Bb9F08c9C39112b0E0b0eccE50"),
		RethAddress:                    common.HexToAddress("0x4be7161080b5d890500194cee2c40B1428002Bd3"),
		RplTokenAddress:                common.HexToAddress("0x59A1a7AebCbF103B3C4f85261fbaC166117E1979"),
		V1_0_0_RewardsPoolAddress:      nil,
		V1_0_0_ClaimNodeAddress:        nil,
		V1_0_0_ClaimTrustedNodeAddress: nil,
		V1_0_0_MinipoolManagerAddress:  nil,
		V1_1_0_NetworkPricesAddress:    nil,
		V1_1_0_NodeStakingAddress:      nil,
		V1_1_0_NodeDepositAddress:      nil,
		V1_1_0_MinipoolQueueAddress:    nil,
		V1_1_0_MinipoolFactoryAddress:  nil,
		V1_2_0_NetworkPricesAddress:    hexToAddressPtr("0xBba3FBCD4Bdbfc79118B1B31218602E5A71B426c"),
		V1_2_0_NetworkBalancesAddress:  hexToAddressPtr("0xBe8Dc8CA5f339c196Aef634DfcDFbA61E30DC743"),
		SnapshotDelegationAddress:      nil,
		SnapshotApiDomain:              "",
		PreviousRewardsPoolAddresses: []common.Address{
			common.HexToAddress("0x4d581a552490fb6fce5F978e66560C8b7E481818"),
		},
		PreviousProtocolDaoVerifierAddresses:   []common.Address{},
		PreviousRocketNetworkPricesAddresses:   []common.Address{},
		PreviousRocketNetworkBalancesAddresses: []common.Address{},
		OptimismPriceMessengerAddress:          nil,
		PolygonPriceMessengerAddress:           nil,
		ArbitrumPriceMessengerAddress:          nil,
		ArbitrumPriceMessengerAddressV2:        nil,
		ZkSyncEraPriceMessengerAddress:         nil,
		BasePriceMessengerAddress:              nil,
		ScrollPriceMessengerAddress:            nil,
		ScrollFeeEstimatorAddress:              nil,
		RplTwapPoolAddress:                     hexToAddressPtr("0x7bb10d2a3105ed5cc150c099a06cafe43d8aa15d"),
	}
	devnetResources.NetworkResources.Network = Network_Devnet

	switch network {
	case config.Network_Mainnet:
		return mainnetResources
	case config.Network_Holesky:
		return holeskyResources
	case Network_Devnet:
		return devnetResources
	}

	panic(fmt.Sprintf("network %s is not supported", network))
}

// Convert a hex string to an address, wrapped in a pointer
func hexToAddressPtr(hexAddress string) *common.Address {
	address := common.HexToAddress(hexAddress)
	return &address
}
