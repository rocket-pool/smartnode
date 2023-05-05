package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	smartnodeTag                       string = "rocketpool/smartnode:v" + shared.RocketPoolVersion
	pruneProvisionerTag                string = "rocketpool/eth1-prune-provision:v0.0.1"
	ecMigratorTag                      string = "rocketpool/ec-migrator:v1.0.0"
	NetworkID                          string = "network"
	ProjectNameID                      string = "projectName"
	SnapshotID                         string = "rocketpool-dao.eth"
	RewardsTreeFilenameFormat          string = "rp-rewards-%s-%d.json"
	MinipoolPerformanceFilenameFormat  string = "rp-minipool-performance-%s-%d.json"
	RewardsTreeIpfsExtension           string = ".zst"
	RewardsTreesFolder                 string = "rewards-trees"
	DaemonDataPath                     string = "/.rocketpool/data"
	WatchtowerFolder                   string = "watchtower"
	WatchtowerStateFile                string = "state.yml"
	RegenerateRewardsTreeRequestSuffix string = ".request"
	RegenerateRewardsTreeRequestFormat string = "%d" + RegenerateRewardsTreeRequestSuffix
	PrimaryRewardsFileUrl              string = "https://%s.ipfs.dweb.link/%s"
	SecondaryRewardsFileUrl            string = "https://ipfs.io/ipfs/%s/%s"
	GithubRewardsFileUrl               string = "https://github.com/rocket-pool/rewards-trees/raw/main/%s/%s"
	FeeRecipientFilename               string = "rp-fee-recipient.txt"
	NativeFeeRecipientFilename         string = "rp-fee-recipient-env.txt"
)

// Defaults
const (
	defaultProjectName       string = "rocketpool"
	WatchtowerMaxFeeDefault  uint64 = 200
	WatchtowerPrioFeeDefault uint64 = 3
)

// Configuration for the Smartnode
type SmartnodeConfig struct {
	Title string `yaml:"-"`

	// The parent config
	parent *RocketPoolConfig

	////////////////////////////
	// User-editable settings //
	////////////////////////////

	// Docker container prefix
	ProjectName config.Parameter `yaml:"projectName,omitempty"`

	// The path of the data folder where everything is stored
	DataPath config.Parameter `yaml:"dataPath,omitempty"`

	// The path of the watchtower's persistent state storage
	WatchtowerStatePath config.Parameter `yaml:"watchtowerStatePath"`

	// Which network we're on
	Network config.Parameter `yaml:"network,omitempty"`

	// Manual max fee override
	ManualMaxFee config.Parameter `yaml:"manualMaxFee,omitempty"`

	// Manual priority fee override
	PriorityFee config.Parameter `yaml:"priorityFee,omitempty"`

	// Threshold for automatic transactions
	AutoTxGasThreshold config.Parameter `yaml:"minipoolStakeGasThreshold,omitempty"`

	// The amount of ETH in a minipool's balance before auto-distribute kicks in
	DistributeThreshold config.Parameter `yaml:"distributeThreshold,omitempty"`

	// Mode for acquiring Merkle rewards trees
	RewardsTreeMode config.Parameter `yaml:"rewardsTreeMode,omitempty"`

	// URL for an EC with archive mode, for manual rewards tree generation
	ArchiveECUrl config.Parameter `yaml:"archiveEcUrl,omitempty"`

	// Token for Oracle DAO members to use when uploading Merkle trees to Web3.Storage
	Web3StorageApiToken config.Parameter `yaml:"web3StorageApiToken,omitempty"`

	// Manual override for the watchtower's max fee
	WatchtowerMaxFeeOverride config.Parameter `yaml:"watchtowerMaxFeeOverride,omitempty"`

	// Manual override for the watchtower's priority fee
	WatchtowerPrioFeeOverride config.Parameter `yaml:"watchtowerPrioFeeOverride,omitempty"`

	// The epoch to switch over to TWAP for RPL price reporting
	RplTwapEpoch config.Parameter `yaml:"rplTwapEpoch,omitempty"`

	// The epoch to start using the new network balance calculation implementation
	BalancesModernizationEpoch config.Parameter `yaml:"balancesModernizationEpoch,omitempty"`

	// The epoch to start using the new fee distributor share calculation
	NewFeeDistributorCalcEpoch config.Parameter `yaml:"newFeeDistributorCalcEpoch,omitempty"`

	///////////////////////////
	// Non-editable settings //
	///////////////////////////

	// The URL to provide the user so they can follow pending transactions
	txWatchUrl map[config.Network]string `yaml:"-"`

	// The URL to use for staking rETH
	stakeUrl map[config.Network]string `yaml:"-"`

	// The map of networks to execution chain IDs
	chainID map[config.Network]uint `yaml:"-"`

	// The contract address of RocketStorage
	storageAddress map[config.Network]string `yaml:"-"`

	// The contract address of the 1inch oracle
	oneInchOracleAddress map[config.Network]string `yaml:"-"`

	// The contract address of the RPL token
	rplTokenAddress map[config.Network]string `yaml:"-"`

	// The contract address of the RPL faucet
	rplFaucetAddress map[config.Network]string `yaml:"-"`

	// The contract address for Snapshot delegation
	snapshotDelegationAddress map[config.Network]string `yaml:"-"`

	// The Snapshot API domain
	snapshotApiDomain map[config.Network]string `yaml:"-"`

	// The contract address of rETH
	rethAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketRewardsPool from v1.0.0
	v1_0_0_RewardsPoolAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketClaimNode from v1.0.0
	v1_0_0_ClaimNodeAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketClaimTrustedNode from v1.0.0
	v1_0_0_ClaimTrustedNodeAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketMinipoolManager from v1.0.0
	v1_0_0_MinipoolManagerAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketNetworkPrices from v1.1.0
	v1_1_0_NetworkPricesAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketNodeStaking from v1.1.0
	v1_1_0_NodeStakingAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketNodeDeposit from v1.1.0
	v1_1_0_NodeDepositAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketMinipoolQueue from v1.1.0
	v1_1_0_MinipoolQueueAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketMinipoolFactory from v1.1.0
	v1_1_0_MinipoolFactoryAddress map[config.Network]string `yaml:"-"`

	// Addresses for RocketRewardsPool that have been upgraded during development
	previousRewardsPoolAddresses map[config.Network]map[string][]common.Address `yaml:"-"`

	// The RocketOvmPriceMessenger Optimism address for each network
	optimismPriceMessengerAddress map[config.Network]string `yaml:"-"`

	// The RocketPolygonPriceMessenger Polygon address for each network
	polygonPriceMessengerAddress map[config.Network]string `yaml:"-"`

	// The RocketArbitumPriceMessenger Arbitrum address for each network
	arbitrumPriceMessengerAddress map[config.Network]string `yaml:"-"`

	// The RocketZkSyncPriceMessenger zkSyncEra address for each network
	zkSyncEraPriceMessengerAddress map[config.Network]string `yaml:"-"`

	// Rewards submission block maps
	rewardsSubmissionBlockMaps map[config.Network][]uint64 `yaml:"-"`

	// The UniswapV3 pool address for each network (used for RPL price TWAP info)
	rplTwapPoolAddress map[config.Network]string `yaml:"-"`

	// The multicall contract address
	multicallAddress map[config.Network]string `yaml:"-"`

	// The BalanceChecker contract address
	balancebatcherAddress map[config.Network]string `yaml:"-"`

	// The FlashBots Protect RPC endpoint
	flashbotsProtectUrl map[config.Network]string `yaml:"-"`
}

// Generates a new Smartnode configuration
func NewSmartnodeConfig(cfg *RocketPoolConfig) *SmartnodeConfig {

	return &SmartnodeConfig{
		Title:  "Smartnode Settings",
		parent: cfg,

		ProjectName: config.Parameter{
			ID:                   ProjectNameID,
			Name:                 "Project Name",
			Description:          "This is the prefix that will be attached to all of the Docker containers managed by the Smartnode.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: defaultProjectName},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth1, config.ContainerID_Eth2, config.ContainerID_Validator, config.ContainerID_Grafana, config.ContainerID_Prometheus, config.ContainerID_Exporter},
			EnvironmentVariables: []string{"COMPOSE_PROJECT_NAME"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		DataPath: config.Parameter{
			ID:                   "dataPath",
			Name:                 "Data Path",
			Description:          "The absolute path of the `data` folder that contains your node wallet's encrypted file, the password for your node wallet, and all of the validator keys for your minipools. You may use environment variables in this string.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: getDefaultDataDir(cfg)},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Validator},
			EnvironmentVariables: []string{"ROCKETPOOL_DATA_FOLDER"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		WatchtowerStatePath: config.Parameter{
			ID:                   "watchtowerPath",
			Name:                 "Watchtower Path",
			Description:          "The absolute path of the watchtower state folder that contains persistent state that is used by the watchtower process on trusted nodes. **Only relevant for trusted nodes.**",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: "$HOME/.rocketpool/watchtower"},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Watchtower},
			EnvironmentVariables: []string{"ROCKETPOOL_WATCHTOWER_FOLDER"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Network: config.Parameter{
			ID:                   NetworkID,
			Name:                 "Network",
			Description:          "The Ethereum network you want to use - select Prater Testnet to practice with fake ETH, or Mainnet to stake on the real network using real ETH.",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: config.Network_Mainnet},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth1, config.ContainerID_Eth2, config.ContainerID_Validator},
			EnvironmentVariables: []string{"NETWORK"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options:              getNetworkOptions(),
		},

		ManualMaxFee: config.Parameter{
			ID:                   "manualMaxFee",
			Name:                 "Manual Max Fee",
			Description:          "Set this if you want all of the Smartnode's transactions to use this specific max fee value (in gwei), which is the most you'd be willing to pay (*including the priority fee*).\n\nA value of 0 will show you the current suggested max fee based on the current network conditions and let you specify it each time you do a transaction.\n\nAny other value will ignore the recommended max fee and explicitly use this value instead.\n\nThis applies to automated transactions (such as claiming RPL and staking minipools) as well.",
			Type:                 config.ParameterType_Float,
			Default:              map[config.Network]interface{}{config.Network_All: float64(0)},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		PriorityFee: config.Parameter{
			ID:                   "priorityFee",
			Name:                 "Priority Fee",
			Description:          "The default value for the priority fee (in gwei) for all of your transactions. This describes how much you're willing to pay *above the network's current base fee* - the higher this is, the more ETH you give to the validators for including your transaction, which generally means it will be included in a block faster (as long as your max fee is sufficiently high to cover the current network conditions).\n\nMust be larger than 0.",
			Type:                 config.ParameterType_Float,
			Default:              map[config.Network]interface{}{config.Network_All: float64(2)},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		AutoTxGasThreshold: config.Parameter{
			ID:   "minipoolStakeGasThreshold",
			Name: "Automatic TX Gas Threshold",
			Description: "Occasionally, the Smartnode will attempt to perform some automatic transactions (such as the second `stake` transaction to finish launching a minipool or the `reduce bond` transaction to convert a 16-ETH minipool to an 8-ETH one). During these, your node will use the `Rapid` suggestion from the gas estimator as its max fee.\n\nThis threshold is a limit (in gwei) you can put on that suggestion; your node will not `stake` the new minipool until the suggestion is below this limit.\n\n" +
				"A value of 0 will disable non-essential automatic transactions (such as minipool balance distribution and bond reduction), but essential transactions (such as minipool staking and solo migration promotion) will not be disabled.\n\n" +
				"NOTE: the node will ignore this limit and automatically execute transactions at whatever the suggested fee happens to be once too much time has passed since those transactions were first eligible. You may end up paying more than you wanted to if you set this too low!",
			Type:                 config.ParameterType_Float,
			Default:              map[config.Network]interface{}{config.Network_All: float64(150)},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Node},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		DistributeThreshold: config.Parameter{
			ID:                   "distributeThreshold",
			Name:                 "Auto-Distribute Threshold",
			Description:          "The Smartnode will regularly check the balance of each of your minipools on the Execution Layer (**not** the Beacon Chain).\nIf any of them have a balance greater than this threshold (in ETH), the Smartnode will automatically distribute the balance. This will send your share of the balance to your withdrawal address.\n\nMust be less than 8 ETH.\n\nSet this to 0 to disable automatic distributes.\n[orange]WARNING: if you disable automatic distribution, you **must** ensure you distribute your minipool's balance before it reaches 8 ETH or you will no longer be able to distribute your rewards until you exit the minipool!",
			Type:                 config.ParameterType_Float,
			Default:              map[config.Network]interface{}{config.Network_All: float64(1)},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Node},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		RewardsTreeMode: config.Parameter{
			ID:                   "rewardsTreeMode",
			Name:                 "Rewards Tree Mode",
			Description:          "Select how you want to acquire the Merkle Tree files for each rewards interval.",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: config.RewardsMode_Download},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []config.ParameterOption{{
				Name:        "Download",
				Description: "Automatically download the Merkle Tree rewards files that were published by the Oracle DAO after a rewards checkpoint.",
				Value:       config.RewardsMode_Download,
			}, {
				Name:        "Generate",
				Description: "Use your node to automatically generate the Merkle Tree rewards file once a checkpoint has passed. This option lets you build and verify the file that the Oracle DAO created if you prefer not to trust it and want to generate the tree yourself.\n\n[orange]WARNING: Generating the tree can take a *very long time* if many node operators are opted into the Smoothing Pool, which could impact your attestation performance!",
				Value:       config.RewardsMode_Generate,
			}},
		},

		ArchiveECUrl: config.Parameter{
			ID:                   "archiveECUrl",
			Name:                 "Archive-Mode EC URL",
			Description:          "[orange]**For manual Merkle rewards tree generation only.**[white]\n\nGenerating the Merkle rewards tree files for past rewards intervals typically requires an Execution client with Archive mode enabled, which is usually disabled on your primary and fallback Execution clients to save disk space.\nIf you want to generate your own rewards tree files for intervals from a long time ago, you may enter the URL of an Execution client with Archive access here.\n\nFor a free light client with Archive access, you may use https://www.alchemy.com/supernode.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		Web3StorageApiToken: config.Parameter{
			ID:                   "web3StorageApiToken",
			Name:                 "Web3.Storage API Token",
			Description:          "[orange]**For Oracle DAO members only.**\n\n[white]The API token for your https://web3.storage/ account. This is required in order for you to upload Merkle rewards trees to Web3.Storage at each rewards interval.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		WatchtowerMaxFeeOverride: config.Parameter{
			ID:                   "watchtowerMaxFeeOverride",
			Name:                 "Watchtower Max Fee Override",
			Description:          fmt.Sprintf("[orange]**For Oracle DAO members only.**\n\n[white]Use this to override the max fee (in gwei) for watchtower transactions. Note that if you set it below %d, the setting will be ignored; it can only be used to set the max fee higher than %d during times of extreme network stress.", WatchtowerMaxFeeDefault, WatchtowerMaxFeeDefault),
			Type:                 config.ParameterType_Float,
			Default:              map[config.Network]interface{}{config.Network_All: float64(WatchtowerMaxFeeDefault)},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		WatchtowerPrioFeeOverride: config.Parameter{
			ID:                   "watchtowerPrioFeeOverride",
			Name:                 "Watchtower Priority Fee Override",
			Description:          fmt.Sprintf("[orange]**For Oracle DAO members only.**\n\n[white]Use this to override the priority fee (in gwei) for watchtower transactions. Note that if you set it below %d, the setting will be ignored; it can only be used to set the priority fee higher than %d during times of extreme network stress.", WatchtowerPrioFeeDefault, WatchtowerPrioFeeDefault),
			Type:                 config.ParameterType_Float,
			Default:              map[config.Network]interface{}{config.Network_All: float64(WatchtowerPrioFeeDefault)},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		RplTwapEpoch: config.Parameter{
			ID:          "rplTwapEpoch",
			Name:        "RPL TWAP Epoch",
			Description: "[orange]**For Oracle DAO members only.**\n\n[white]The epoch to switch from spot prices to TWAP for RPL price submission.",
			Type:        config.ParameterType_Uint,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: uint64(193414),
				config.Network_Prater:  uint64(162094),
				config.Network_Devnet:  uint64(162094),
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		BalancesModernizationEpoch: config.Parameter{
			ID:          "balancesModernizationEpoch",
			Name:        "Balances Modernization Epoch",
			Description: "[orange]**For Oracle DAO members only.**\n\n[white]The epoch to switch from the old network balance calculation method to the new one.",
			Type:        config.ParameterType_Uint,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: uint64(194089),
				config.Network_Prater:  uint64(162094),
				config.Network_Devnet:  uint64(162094),
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		NewFeeDistributorCalcEpoch: config.Parameter{
			ID:          "newFeeDistributorCalcEpoch",
			Name:        "New Fee Distributor Calculation Epoch",
			Description: "[orange]**For Oracle DAO members only.**\n\n[white]The epoch to switch from the old fee distributor share calculation method to the new one.",
			Type:        config.ParameterType_Uint,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: uint64(999999),
				config.Network_Prater:  uint64(999999),
				config.Network_Devnet:  uint64(999999),
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		txWatchUrl: map[config.Network]string{
			config.Network_Mainnet: "https://etherscan.io/tx",
			config.Network_Prater:  "https://goerli.etherscan.io/tx",
			config.Network_Devnet:  "https://goerli.etherscan.io/tx",
		},

		stakeUrl: map[config.Network]string{
			config.Network_Mainnet: "https://stake.rocketpool.net",
			config.Network_Prater:  "https://testnet.rocketpool.net",
			config.Network_Devnet:  "TBD",
		},

		chainID: map[config.Network]uint{
			config.Network_Mainnet: 1, // Mainnet
			config.Network_Prater:  5, // Goerli
			config.Network_Devnet:  5, // Also goerli
		},

		storageAddress: map[config.Network]string{
			config.Network_Mainnet: "0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46",
			config.Network_Prater:  "0xd8Cd47263414aFEca62d6e2a3917d6600abDceB3",
			config.Network_Devnet:  "0x6A18E47f8CcB453Dd0894AC003f74BEE7e47A368",
		},

		oneInchOracleAddress: map[config.Network]string{
			config.Network_Mainnet: "0x07D91f5fb9Bf7798734C3f606dB065549F6893bb",
			config.Network_Prater:  "0x4eDC966Df24264C9C817295a0753804EcC46Dd22",
			config.Network_Devnet:  "0x4eDC966Df24264C9C817295a0753804EcC46Dd22",
		},

		rplTokenAddress: map[config.Network]string{
			config.Network_Mainnet: "0xD33526068D116cE69F19A9ee46F0bd304F21A51f",
			config.Network_Prater:  "0x5e932688e81a182e3de211db6544f98b8e4f89c7",
			config.Network_Devnet:  "0x09b6aEF57B580f5CB46746BA59ed312Ba80E8Ad4",
		},

		rplFaucetAddress: map[config.Network]string{
			config.Network_Mainnet: "",
			config.Network_Prater:  "0x95D6b8E2106E3B30a72fC87e2B56ce15E37853F9",
			config.Network_Devnet:  "0x218a718A1B23B13737E2F566Dd45730E8DAD451b",
		},

		rethAddress: map[config.Network]string{
			config.Network_Mainnet: "0xae78736Cd615f374D3085123A210448E74Fc6393",
			config.Network_Prater:  "0x178E141a0E3b34152f73Ff610437A7bf9B83267A",
			config.Network_Devnet:  "0x2DF914425da6d0067EF1775AfDBDd7B24fc8100E",
		},

		v1_0_0_RewardsPoolAddress: map[config.Network]string{
			config.Network_Mainnet: "0xA3a18348e6E2d3897B6f2671bb8c120e36554802",
			config.Network_Prater:  "0xf9aE18eB0CE4930Bc3d7d1A5E33e4286d4FB0f8B",
			config.Network_Devnet:  "0x4A1b5Ab9F6C36E7168dE5F994172028Ca8554e02",
		},

		v1_0_0_ClaimNodeAddress: map[config.Network]string{
			config.Network_Mainnet: "0x899336A2a86053705E65dB61f52C686dcFaeF548",
			config.Network_Prater:  "0xc05b7A2a03A6d2736d1D0ebf4d4a0aFE2cc32cE1",
			config.Network_Devnet:  "",
		},

		v1_0_0_ClaimTrustedNodeAddress: map[config.Network]string{
			config.Network_Mainnet: "0x6af730deB0463b432433318dC8002C0A4e9315e8",
			config.Network_Prater:  "0x730982F4439E5AC30292333ff7d0C478907f2219",
			config.Network_Devnet:  "",
		},

		v1_0_0_MinipoolManagerAddress: map[config.Network]string{
			config.Network_Mainnet: "0x6293B8abC1F36aFB22406Be5f96D893072A8cF3a",
			config.Network_Prater:  "0xB815a94430f08dD2ab61143cE1D5739Ac81D3C6d",
			config.Network_Devnet:  "",
		},

		v1_1_0_NetworkPricesAddress: map[config.Network]string{
			config.Network_Mainnet: "0xd3f500F550F46e504A4D2153127B47e007e11166",
			config.Network_Prater:  "0x12f96dC173a806D18d71fAFe3C1BA2149c3E3Dc6",
			config.Network_Devnet:  "",
		},

		v1_1_0_NodeStakingAddress: map[config.Network]string{
			config.Network_Mainnet: "0xA73ec45Fe405B5BFCdC0bF4cbc9014Bb32a01cd2",
			config.Network_Prater:  "0xA73ec45Fe405B5BFCdC0bF4cbc9014Bb32a01cd2",
			config.Network_Devnet:  "",
		},

		v1_1_0_NodeDepositAddress: map[config.Network]string{
			config.Network_Mainnet: "0x1Cc9cF5586522c6F483E84A19c3C2B0B6d027bF0",
			config.Network_Prater:  "0x1Cc9cF5586522c6F483E84A19c3C2B0B6d027bF0",
			config.Network_Devnet:  "",
		},

		v1_1_0_MinipoolQueueAddress: map[config.Network]string{
			config.Network_Mainnet: "0x5870dA524635D1310Dc0e6F256Ce331012C9C19E",
			config.Network_Prater:  "0xEF5EF45bf1CC08D5694f87F8c4023f00CCCB7237",
			config.Network_Devnet:  "",
		},

		v1_1_0_MinipoolFactoryAddress: map[config.Network]string{
			config.Network_Mainnet: "0x54705f80D7C51Fcffd9C659ce3f3C9a7dCCf5788",
			config.Network_Prater:  "0x54705f80D7C51Fcffd9C659ce3f3C9a7dCCf5788",
			config.Network_Devnet:  "",
		},

		snapshotDelegationAddress: map[config.Network]string{
			config.Network_Mainnet: "0x469788fE6E9E9681C6ebF3bF78e7Fd26Fc015446",
			config.Network_Prater:  "0xD0897D68Cd66A710dDCecDe30F7557972181BEDc",
			config.Network_Devnet:  "",
		},

		snapshotApiDomain: map[config.Network]string{
			config.Network_Mainnet: "hub.snapshot.org",
			config.Network_Prater:  "testnet.snapshot.org",
			config.Network_Devnet:  "",
		},

		previousRewardsPoolAddresses: map[config.Network]map[string][]common.Address{
			config.Network_Mainnet: {
				"v1.1.0": []common.Address{
					common.HexToAddress("0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1"),
				},
			},
			config.Network_Prater: {
				"v1.1.0-rc1": []common.Address{
					common.HexToAddress("0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1"),
				},
				"v1.2.0-rc1": []common.Address{
					common.HexToAddress("0x6e91E3416acf3d015358eeAAF247a0674F6c306f"),
				},
			},
			config.Network_Devnet: {},
		},

		optimismPriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0xdddcf2c25d50ec22e67218e873d46938650d03a7",
			config.Network_Prater:  "0x87E2deCE7d0A080D579f63cbcD7e1629BEcd7E7d",
			config.Network_Devnet:  "",
		},

		polygonPriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0xb1029Ac2Be4e08516697093e2AFeC435057f3511",
			config.Network_Prater:  "0x6D736da1dC2562DBeA9998385A0A27d8c2B2793e",
			config.Network_Devnet:  "0x6D736da1dC2562DBeA9998385A0A27d8c2B2793e",
		},

		arbitrumPriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0x05330300f829AD3fC8f33838BC88CFC4093baD53",
			config.Network_Prater:  "0x2b52479F6ea009907e46fc43e91064D1b92Fdc86",
			config.Network_Devnet:  "0x2b52479F6ea009907e46fc43e91064D1b92Fdc86",
		},

		zkSyncEraPriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0x6cf6CB29754aEBf88AF12089224429bD68b0b8c8",
			config.Network_Prater:  "0x3Fd49431bD05875AeD449Bc8C07352942A7fBA75",
			config.Network_Devnet:  "0x3Fd49431bD05875AeD449Bc8C07352942A7fBA75",
		},

		rplTwapPoolAddress: map[config.Network]string{
			config.Network_Mainnet: "0xe42318ea3b998e8355a3da364eb9d48ec725eb45",
			config.Network_Prater:  "0x5cE71E603B138F7e65029Cc1918C0566ed0dBD4B",
			config.Network_Devnet:  "0x5cE71E603B138F7e65029Cc1918C0566ed0dBD4B",
		},

		multicallAddress: map[config.Network]string{
			config.Network_Mainnet: "0x5BA1e12693Dc8F9c48aAD8770482f4739bEeD696",
			config.Network_Prater:  "0x5BA1e12693Dc8F9c48aAD8770482f4739bEeD696",
			config.Network_Devnet:  "0x5BA1e12693Dc8F9c48aAD8770482f4739bEeD696",
		},

		balancebatcherAddress: map[config.Network]string{
			config.Network_Mainnet: "0xb1f8e55c7f64d203c1400b9d8555d050f94adf39",
			config.Network_Prater:  "0x9788C4E93f9002a7ad8e72633b11E8d1ecd51f9b",
			config.Network_Devnet:  "0x9788C4E93f9002a7ad8e72633b11E8d1ecd51f9b",
		},

		flashbotsProtectUrl: map[config.Network]string{
			config.Network_Mainnet: "https://rpc.flashbots.net/",
			config.Network_Prater:  "https://rpc-goerli.flashbots.net/",
			config.Network_Devnet:  "https://rpc-goerli.flashbots.net/",
		},

		rewardsSubmissionBlockMaps: map[config.Network][]uint64{
			config.Network_Mainnet: {
				15451165, 15637542, 15839520, 16038366, 16238906, 16439406, // 5
				16639856, 16841781, 17037278,
			},
			config.Network_Prater: {
				7287326, 7297026, 7314231, 7331462, 7387271, 7412366, // 5
				7420574, 7436546, 7456423, 7473017, 7489726, 7506706, // 11
				7525902, 7544630, 7562851, 7581623, 7600343, 7618815, // 17
				7636720, 7654452, 7672147, 7689735, 7707617, 7725232, // 23
				7742548, 7760702, 7777078, 7794263, 7811800, 7829115, // 29
				7846870, 7863708, 7881537, 7900095, 7918951, 7937222, // 35
				7955161, 7972837, 7990504, 8008474, 8027271, 8045546, // 41
				8063957, 8082659, 8101400, 8119473, 8136892, 8154565, // 47
				8172349, 8189717, 8207105, 8224279, 8241674, 8258210, // 53
				8274526, 8290763, 8307407, 8324452, 8341708, 8359470, // 59
				8377175, 8394786, 8412599, 8430221, 8447800, 8465317, // 65
				8482337, 8499227, 8516593, 8533890, 8551379, 8569494, // 71
				8587146, 8604666, 8621961, 8639563, 8656830, 8673617, // 77
				8690655, 8707453, 8724467, 8742735, 8758413, 8775532, // 83
				8792725, 8809501,
			},
			config.Network_Devnet: {
				7955303, 7972424, 8009064, 8026821, 8045113, 8063501, // 5
				8082186, 8100941, 8119074, 8136452, 8154152, 8171923, // 11
				8189312, 8206689, 8223857, 8241269, 8257834, 8274178, // 17
				8290333, 8307005, 8324055, 8341308, 8359051, 8376744, // 23
				8394338, 8412142,
			},
		},
	}

}

// Get the parameters for this config
func (cfg *SmartnodeConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Network,
		&cfg.ProjectName,
		&cfg.DataPath,
		&cfg.ManualMaxFee,
		&cfg.PriorityFee,
		&cfg.AutoTxGasThreshold,
		&cfg.DistributeThreshold,
		&cfg.RewardsTreeMode,
		&cfg.ArchiveECUrl,
		&cfg.Web3StorageApiToken,
		&cfg.WatchtowerMaxFeeOverride,
		&cfg.WatchtowerPrioFeeOverride,
		&cfg.RplTwapEpoch,
		&cfg.BalancesModernizationEpoch,
		&cfg.NewFeeDistributorCalcEpoch,
	}
}

// Getters for the non-editable parameters

func (cfg *SmartnodeConfig) GetTxWatchUrl() string {
	return cfg.txWatchUrl[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetStakeUrl() string {
	return cfg.stakeUrl[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetChainID() uint {
	return cfg.chainID[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetWalletPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "wallet")
	}

	return filepath.Join(DaemonDataPath, "wallet")
}

func (cfg *SmartnodeConfig) GetPasswordPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "password")
	}

	return filepath.Join(DaemonDataPath, "password")
}

func (cfg *SmartnodeConfig) GetValidatorKeychainPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "validators")
	}

	return filepath.Join(DaemonDataPath, "validators")
}

func (cfg *SmartnodeConfig) GetWalletPathInCLI() string {
	return filepath.Join(cfg.DataPath.Value.(string), "wallet")
}

func (cfg *SmartnodeConfig) GetPasswordPathInCLI() string {
	return filepath.Join(cfg.DataPath.Value.(string), "password")
}

func (cfg *SmartnodeConfig) GetValidatorKeychainPathInCLI() string {
	return filepath.Join(cfg.DataPath.Value.(string), "validators")
}

func (config *SmartnodeConfig) GetWatchtowerStatePath() string {
	if config.parent.IsNativeMode {
		return filepath.Join(config.DataPath.Value.(string), WatchtowerFolder, "state.yml")
	}

	return filepath.Join(DaemonDataPath, WatchtowerFolder, "state.yml")
}

func (cfg *SmartnodeConfig) GetCustomKeyPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "custom-keys")
	}

	return filepath.Join(DaemonDataPath, "custom-keys")
}

func (cfg *SmartnodeConfig) GetCustomKeyPasswordFilePath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "custom-key-passwords")
	}

	return filepath.Join(DaemonDataPath, "custom-key-passwords")
}

func (cfg *SmartnodeConfig) GetStorageAddress() string {
	return cfg.storageAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetOneInchOracleAddress() string {
	return cfg.oneInchOracleAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetRplTokenAddress() string {
	return cfg.rplTokenAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetRplFaucetAddress() string {
	return cfg.rplFaucetAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetSnapshotDelegationAddress() string {
	return cfg.snapshotDelegationAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetSmartnodeContainerTag() string {
	return smartnodeTag
}

func (config *SmartnodeConfig) GetPruneProvisionerContainerTag() string {
	return pruneProvisionerTag
}

func (cfg *SmartnodeConfig) GetEcMigratorContainerTag() string {
	return ecMigratorTag
}

func (cfg *SmartnodeConfig) GetSnapshotApiDomain() string {
	return cfg.snapshotApiDomain[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetVotingSnapshotID() [32]byte {
	// So the contract wants a Keccak'd hash of the voting ID, but Snapshot's service wants ASCII so it can display the ID in plain text; we have to do this to make it play nicely with Snapshot
	buffer := [32]byte{}
	idBytes := []byte(SnapshotID)
	copy(buffer[0:], idBytes)
	return buffer
}

func (config *SmartnodeConfig) GetSnapshotID() string {
	return SnapshotID
}

// The the title for the config
func (cfg *SmartnodeConfig) GetConfigTitle() string {
	return cfg.Title
}

func (cfg *SmartnodeConfig) GetRethAddress() common.Address {
	return common.HexToAddress(cfg.rethAddress[cfg.Network.Value.(config.Network)])
}

func getDefaultDataDir(config *RocketPoolConfig) string {
	return filepath.Join(config.RocketPoolDirectory, "data")
}

func (cfg *SmartnodeConfig) GetRewardsTreePath(interval uint64, daemon bool) string {
	if daemon && !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, RewardsTreesFolder, fmt.Sprintf(RewardsTreeFilenameFormat, string(cfg.Network.Value.(config.Network)), interval))
	}

	return filepath.Join(cfg.DataPath.Value.(string), RewardsTreesFolder, fmt.Sprintf(RewardsTreeFilenameFormat, string(cfg.Network.Value.(config.Network)), interval))
}

func (cfg *SmartnodeConfig) GetMinipoolPerformancePath(interval uint64, daemon bool) string {
	if daemon && !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, RewardsTreesFolder, fmt.Sprintf(MinipoolPerformanceFilenameFormat, string(cfg.Network.Value.(config.Network)), interval))
	}

	return filepath.Join(cfg.DataPath.Value.(string), RewardsTreesFolder, fmt.Sprintf(MinipoolPerformanceFilenameFormat, string(cfg.Network.Value.(config.Network)), interval))
}

func (cfg *SmartnodeConfig) GetRegenerateRewardsTreeRequestPath(interval uint64, daemon bool) string {
	if daemon && !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, WatchtowerFolder, fmt.Sprintf(RegenerateRewardsTreeRequestFormat, interval))
	}

	return filepath.Join(cfg.DataPath.Value.(string), WatchtowerFolder, fmt.Sprintf(RegenerateRewardsTreeRequestFormat, interval))
}

func (cfg *SmartnodeConfig) GetWatchtowerFolder(daemon bool) string {
	if daemon && !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, WatchtowerFolder)
	}

	return filepath.Join(cfg.DataPath.Value.(string), WatchtowerFolder)
}

func (cfg *SmartnodeConfig) GetFeeRecipientFilePath() string {
	if !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, "validators", FeeRecipientFilename)
	}

	return filepath.Join(cfg.DataPath.Value.(string), "validators", NativeFeeRecipientFilename)
}

func (cfg *SmartnodeConfig) GetV100RewardsPoolAddress() common.Address {
	return common.HexToAddress(cfg.v1_0_0_RewardsPoolAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetV100ClaimNodeAddress() common.Address {
	return common.HexToAddress(cfg.v1_0_0_ClaimNodeAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetV100ClaimTrustedNodeAddress() common.Address {
	return common.HexToAddress(cfg.v1_0_0_ClaimTrustedNodeAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetV100MinipoolManagerAddress() common.Address {
	return common.HexToAddress(cfg.v1_0_0_MinipoolManagerAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetV110NetworkPricesAddress() common.Address {
	return common.HexToAddress(cfg.v1_1_0_NetworkPricesAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetV110NodeStakingAddress() common.Address {
	return common.HexToAddress(cfg.v1_1_0_NodeStakingAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetV110NodeDepositAddress() common.Address {
	return common.HexToAddress(cfg.v1_1_0_NodeDepositAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetV110MinipoolQueueAddress() common.Address {
	return common.HexToAddress(cfg.v1_1_0_MinipoolQueueAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetV110MinipoolFactoryAddress() common.Address {
	return common.HexToAddress(cfg.v1_1_0_MinipoolFactoryAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetPreviousRewardsPoolAddresses() map[string][]common.Address {
	return cfg.previousRewardsPoolAddresses[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetOptimismMessengerAddress() string {
	return cfg.optimismPriceMessengerAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetPolygonMessengerAddress() string {
	return cfg.polygonPriceMessengerAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetArbitrumMessengerAddress() string {
	return cfg.arbitrumPriceMessengerAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetZkSyncEraMessengerAddress() string {
	return cfg.zkSyncEraPriceMessengerAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetRplTwapPoolAddress() string {
	return cfg.rplTwapPoolAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetMulticallAddress() string {
	return cfg.multicallAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetBalanceBatcherAddress() string {
	return cfg.balancebatcherAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetFlashbotsProtectUrl() string {
	return cfg.flashbotsProtectUrl[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetRewardsSubmissionBlockMaps() []uint64 {
	return cfg.rewardsSubmissionBlockMaps[cfg.Network.Value.(config.Network)]
}

func getNetworkOptions() []config.ParameterOption {
	options := []config.ParameterOption{
		{
			Name:        "Ethereum Mainnet",
			Description: "This is the real Ethereum main network, using real ETH and real RPL to make real validators.",
			Value:       config.Network_Mainnet,
		}, {
			Name:        "Prater Testnet",
			Description: "This is the Prater test network, using free fake ETH and free fake RPL to make fake validators.\nUse this if you want to practice running the Smartnode in a free, safe environment before moving to Mainnet.",
			Value:       config.Network_Prater,
		},
	}

	if strings.HasSuffix(shared.RocketPoolVersion, "-dev") {
		options = append(options, config.ParameterOption{
			Name:        "Devnet",
			Description: "This is a development network used by Rocket Pool engineers to test new features and contract upgrades before they are promoted to Prater for staging. You should not use this network unless invited to do so by the developers.",
			Value:       config.Network_Devnet,
		})
	}

	return options
}
