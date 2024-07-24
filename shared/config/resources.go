package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/config"
	"gopkg.in/yaml.v3"
)

var (
	// Mainnet resources for reference in testing
	MainnetResourcesReference *SmartNodeResources = &SmartNodeResources{
		StakeUrl:                       "https://stake.rocketpool.net",
		StorageAddress:                 common.HexToAddress("0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46"),
		RethAddress:                    common.HexToAddress("0xae78736Cd615f374D3085123A210448E74Fc6393"),
		RplTokenAddress:                common.HexToAddress("0xD33526068D116cE69F19A9ee46F0bd304F21A51f"),
		V1_0_0_RewardsPoolAddress:      config.HexToAddressPtr("0xA3a18348e6E2d3897B6f2671bb8c120e36554802"),
		V1_0_0_ClaimNodeAddress:        config.HexToAddressPtr("0x899336A2a86053705E65dB61f52C686dcFaeF548"),
		V1_0_0_ClaimTrustedNodeAddress: config.HexToAddressPtr("0x6af730deB0463b432433318dC8002C0A4e9315e8"),
		V1_0_0_MinipoolManagerAddress:  config.HexToAddressPtr("0x6293B8abC1F36aFB22406Be5f96D893072A8cF3a"),
		V1_1_0_NetworkPricesAddress:    config.HexToAddressPtr("0xd3f500F550F46e504A4D2153127B47e007e11166"),
		V1_1_0_NodeStakingAddress:      config.HexToAddressPtr("0xA73ec45Fe405B5BFCdC0bF4cbc9014Bb32a01cd2"),
		V1_1_0_NodeDepositAddress:      config.HexToAddressPtr("0x1Cc9cF5586522c6F483E84A19c3C2B0B6d027bF0"),
		V1_1_0_MinipoolQueueAddress:    config.HexToAddressPtr("0x5870dA524635D1310Dc0e6F256Ce331012C9C19E"),
		V1_1_0_MinipoolFactoryAddress:  config.HexToAddressPtr("0x54705f80D7C51Fcffd9C659ce3f3C9a7dCCf5788"),
		V1_2_0_NetworkPricesAddress:    config.HexToAddressPtr("0x751826b107672360b764327631cC5764515fFC37"),
		V1_2_0_NetworkBalancesAddress:  config.HexToAddressPtr("0x07FCaBCbe4ff0d80c2b1eb42855C0131b6cba2F4"),
		SnapshotDelegationAddress:      config.HexToAddressPtr("0x469788fE6E9E9681C6ebF3bF78e7Fd26Fc015446"),
		SnapshotApiDomain:              "hub.snapshot.org",
		PreviousRewardsPoolAddresses: []common.Address{
			common.HexToAddress("0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1"),
		},
		PreviousProtocolDaoVerifierAddresses: []common.Address{},
		OptimismPriceMessengerAddress:        config.HexToAddressPtr("0xdddcf2c25d50ec22e67218e873d46938650d03a7"),
		PolygonPriceMessengerAddress:         config.HexToAddressPtr("0xb1029Ac2Be4e08516697093e2AFeC435057f3511"),
		ArbitrumPriceMessengerAddress:        config.HexToAddressPtr("0x05330300f829AD3fC8f33838BC88CFC4093baD53"),
		ArbitrumPriceMessengerAddressV2:      config.HexToAddressPtr("0x312FcFB03eC9B1Ea38CB7BFCd26ee7bC3b505aB1"),
		ZkSyncEraPriceMessengerAddress:       config.HexToAddressPtr("0x6cf6CB29754aEBf88AF12089224429bD68b0b8c8"),
		BasePriceMessengerAddress:            config.HexToAddressPtr("0x64A5856869C06B0188C84A5F83d712bbAc03517d"),
		ScrollPriceMessengerAddress:          config.HexToAddressPtr("0x0f22dc9b9c03757d4676539203d7549c8f22c15c"),
		ScrollFeeEstimatorAddress:            config.HexToAddressPtr("0x0d7E906BD9cAFa154b048cFa766Cc1E54E39AF9B"),
		RplTwapPoolAddress:                   config.HexToAddressPtr("0xe42318ea3b998e8355a3da364eb9d48ec725eb45"),
	}

	// Holesky resources for reference in testing
	HoleskyResourcesReference *SmartNodeResources = &SmartNodeResources{
		StakeUrl:                       "https://testnet.rocketpool.net",
		StorageAddress:                 common.HexToAddress("0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1"),
		RethAddress:                    common.HexToAddress("0x7322c24752f79c05FFD1E2a6FCB97020C1C264F1"),
		RplTokenAddress:                common.HexToAddress("0x1Cc9cF5586522c6F483E84A19c3C2B0B6d027bF0"),
		V1_0_0_RewardsPoolAddress:      nil,
		V1_0_0_ClaimNodeAddress:        nil,
		V1_0_0_ClaimTrustedNodeAddress: nil,
		V1_0_0_MinipoolManagerAddress:  nil,
		V1_1_0_NetworkPricesAddress:    nil,
		V1_1_0_NodeStakingAddress:      nil,
		V1_1_0_NodeDepositAddress:      nil,
		V1_1_0_MinipoolQueueAddress:    nil,
		V1_1_0_MinipoolFactoryAddress:  nil,
		V1_2_0_NetworkPricesAddress:    config.HexToAddressPtr("0x029d946F28F93399a5b0D09c879FC8c94E596AEb"),
		V1_2_0_NetworkBalancesAddress:  config.HexToAddressPtr("0x9294Fc6F03c64Cc217f5BE8697EA3Ed2De77e2F8"),
		SnapshotDelegationAddress:      nil,
		SnapshotApiDomain:              "",
		PreviousRewardsPoolAddresses: []common.Address{
			common.HexToAddress("0x4a625C617a44E60F74E3fe3bf6d6333b63766e91"),
		},
		PreviousProtocolDaoVerifierAddresses: nil,
		OptimismPriceMessengerAddress:        nil,
		PolygonPriceMessengerAddress:         nil,
		ArbitrumPriceMessengerAddress:        nil,
		ArbitrumPriceMessengerAddressV2:      nil,
		ZkSyncEraPriceMessengerAddress:       nil,
		BasePriceMessengerAddress:            nil,
		ScrollPriceMessengerAddress:          nil,
		ScrollFeeEstimatorAddress:            nil,
		RplTwapPoolAddress:                   config.HexToAddressPtr("0x7bb10d2a3105ed5cc150c099a06cafe43d8aa15d"),
	}

	// Devnet resources for reference in testing
	HoleskyDevResourcesReference *SmartNodeResources = &SmartNodeResources{
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
		V1_2_0_NetworkPricesAddress:    config.HexToAddressPtr("0xBba3FBCD4Bdbfc79118B1B31218602E5A71B426c"),
		V1_2_0_NetworkBalancesAddress:  config.HexToAddressPtr("0xBe8Dc8CA5f339c196Aef634DfcDFbA61E30DC743"),
		SnapshotDelegationAddress:      nil,
		SnapshotApiDomain:              "",
		PreviousRewardsPoolAddresses: []common.Address{
			common.HexToAddress("0x4d581a552490fb6fce5F978e66560C8b7E481818"),
		},
		PreviousProtocolDaoVerifierAddresses: nil,
		OptimismPriceMessengerAddress:        nil,
		PolygonPriceMessengerAddress:         nil,
		ArbitrumPriceMessengerAddress:        nil,
		ArbitrumPriceMessengerAddressV2:      nil,
		ZkSyncEraPriceMessengerAddress:       nil,
		BasePriceMessengerAddress:            nil,
		ScrollPriceMessengerAddress:          nil,
		ScrollFeeEstimatorAddress:            nil,
		RplTwapPoolAddress:                   config.HexToAddressPtr("0x7bb10d2a3105ed5cc150c099a06cafe43d8aa15d"),
	}
)

// Network settings with a field for Rocket Pool-specific settings
type SmartNodeSettings struct {
	*config.NetworkSettings `yaml:",inline"`

	// Rocket Pool resources for the network
	SmartNodeResources *SmartNodeResources `yaml:"smartNodeResources" json:"smartNodeResources"`
}

// A collection of network-specific resources and getters for them
type SmartNodeResources struct {
	// The URL to use for staking rETH
	StakeUrl string `yaml:"stakeUrl" json:"stakeUrl"`

	// The contract address of RocketStorage
	StorageAddress common.Address `yaml:"storageAddress" json:"storageAddress"`

	// The contract address of rETH
	RethAddress common.Address `yaml:"rethAddress" json:"rethAddress"`

	// The contract address of the RPL token
	RplTokenAddress common.Address `yaml:"rplTokenAddress" json:"rplTokenAddress"`

	// The contract address of rocketRewardsPool from v1.0.0
	V1_0_0_RewardsPoolAddress *common.Address `yaml:"v1_0_0_RewardsPoolAddress" json:"v1_0_0_RewardsPoolAddress"`

	// The contract address of rocketClaimNode from v1.0.0
	V1_0_0_ClaimNodeAddress *common.Address `yaml:"v1_0_0_ClaimNodeAddress" json:"v1_0_0_ClaimNodeAddress"`

	// The contract address of rocketClaimTrustedNode from v1.0.0
	V1_0_0_ClaimTrustedNodeAddress *common.Address `yaml:"v1_0_0_ClaimTrustedNodeAddress" json:"v1_0_0_ClaimTrustedNodeAddress"`

	// The contract address of rocketMinipoolManager from v1.0.0
	V1_0_0_MinipoolManagerAddress *common.Address `yaml:"v1_0_0_MinipoolManagerAddress" json:"v1_0_0_MinipoolManagerAddress"`

	// The contract address of rocketNetworkPrices from v1.1.0
	V1_1_0_NetworkPricesAddress *common.Address `yaml:"v1_1_0_NetworkPricesAddress" json:"v1_1_0_NetworkPricesAddress"`

	// The contract address of rocketNodeStaking from v1.1.0
	V1_1_0_NodeStakingAddress *common.Address `yaml:"v1_1_0_NodeStakingAddress" json:"v1_1_0_NodeStakingAddress"`

	// The contract address of rocketNodeDeposit from v1.1.0
	V1_1_0_NodeDepositAddress *common.Address `yaml:"v1_1_0_NodeDepositAddress" json:"v1_1_0_NodeDepositAddress"`

	// The contract address of rocketMinipoolQueue from v1.1.0
	V1_1_0_MinipoolQueueAddress *common.Address `yaml:"v1_1_0_MinipoolQueueAddress" json:"v1_1_0_MinipoolQueueAddress"`

	// The contract address of rocketMinipoolFactory from v1.1.0
	V1_1_0_MinipoolFactoryAddress *common.Address `yaml:"v1_1_0_MinipoolFactoryAddress" json:"v1_1_0_MinipoolFactoryAddress"`

	// The contract address of rocketNetworkPrices from v1.2.0
	V1_2_0_NetworkPricesAddress *common.Address `yaml:"v1_2_0_NetworkPricesAddress" json:"v1_2_0_NetworkPricesAddress"`

	// The contract address of rocketNetworkBalances from v1.2.0
	V1_2_0_NetworkBalancesAddress *common.Address `yaml:"v1_2_0_NetworkBalancesAddress" json:"v1_2_0_NetworkBalancesAddress"`

	// The contract address for Snapshot delegation
	SnapshotDelegationAddress *common.Address `yaml:"snapshotDelegationAddress" json:"snapshotDelegationAddress"`

	// The Snapshot API domain
	SnapshotApiDomain string `yaml:"snapshotApiDomain" json:"snapshotApiDomain"`

	// Addresses for RocketRewardsPool that have been upgraded during development
	PreviousRewardsPoolAddresses []common.Address `yaml:"previousRewardsPoolAddresses" json:"previousRewardsPoolAddresses"`

	// Addresses for RocketDAOProtocolVerifier that have been upgraded during development
	PreviousProtocolDaoVerifierAddresses []common.Address `yaml:"previousProtocolDaoVerifierAddresses" json:"previousProtocolDaoVerifierAddresses"`

	// The RocketOvmPriceMessenger Optimism address for each network
	OptimismPriceMessengerAddress *common.Address `yaml:"optimismPriceMessengerAddress" json:"optimismPriceMessengerAddress"`

	// The RocketPolygonPriceMessenger Polygon address for each network
	PolygonPriceMessengerAddress *common.Address `yaml:"polygonPriceMessengerAddress" json:"polygonPriceMessengerAddress"`

	// The RocketArbitumPriceMessenger Arbitrum address for each network
	ArbitrumPriceMessengerAddress *common.Address `yaml:"arbitrumPriceMessengerAddress" json:"arbitrumPriceMessengerAddress"`

	// The RocketArbitumPriceMessengerV2 Arbitrum address for each network
	ArbitrumPriceMessengerAddressV2 *common.Address `yaml:"arbitrumPriceMessengerAddressV2" json:"arbitrumPriceMessengerAddressV2"`

	// The RocketZkSyncPriceMessenger zkSyncEra address for each network
	ZkSyncEraPriceMessengerAddress *common.Address `yaml:"zkSyncEraPriceMessengerAddress" json:"zkSyncEraPriceMessengerAddress"`

	// The RocketBasePriceMessenger Base address for each network
	BasePriceMessengerAddress *common.Address `yaml:"basePriceMessengerAddress" json:"basePriceMessengerAddress"`

	// The RocketScrollPriceMessenger Scroll address for each network
	ScrollPriceMessengerAddress *common.Address `yaml:"scrollPriceMessengerAddress" json:"scrollPriceMessengerAddress"`

	// The Scroll L2 message fee estimator address for each network
	ScrollFeeEstimatorAddress *common.Address `yaml:"scrollFeeEstimatorAddress" json:"scrollFeeEstimatorAddress"`

	// The UniswapV3 pool address for each network (used for RPL price TWAP info)
	RplTwapPoolAddress *common.Address `yaml:"rplTwapPoolAddress" json:"rplTwapPoolAddress"`
}

// An aggregated collection of resources for the selected network, including Rocket Pool resources
type MergedResources struct {
	// Base network resources
	*config.NetworkResources

	// Rocket Pool resources
	*SmartNodeResources
}

// Load network settings from a folder
func LoadSettingsFiles(sourceDir string) ([]*SmartNodeSettings, error) {
	// Make sure the folder exists
	_, err := os.Stat(sourceDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("network settings folder [%s] does not exist", sourceDir)
	}

	// Enumerate the dir
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("error enumerating override source folder: %w", err)
	}

	settingsList := []*SmartNodeSettings{}
	for _, file := range files {
		// Ignore dirs and nonstandard files
		if file.IsDir() || !file.Type().IsRegular() {
			continue
		}

		// Load the file
		filename := file.Name()
		ext := filepath.Ext(filename)
		if ext != ".yaml" && ext != ".yml" {
			// Only load YAML files
			continue
		}
		settingsFilePath := filepath.Join(sourceDir, filename)
		bytes, err := os.ReadFile(settingsFilePath)
		if err != nil {
			return nil, fmt.Errorf("error reading network settings file [%s]: %w", settingsFilePath, err)
		}

		// Unmarshal the settings
		settings := new(SmartNodeSettings)
		err = yaml.Unmarshal(bytes, settings)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling network settings file [%s]: %w", settingsFilePath, err)
		}
		settingsList = append(settingsList, settings)
	}
	return settingsList, nil
}
