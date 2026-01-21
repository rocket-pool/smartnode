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
	smartnodeTagPrefix                 string = "rocketpool/smartnode:v"
	NetworkID                          string = "network"
	ProjectNameID                      string = "projectName"
	SnapshotID                         string = "rocketpool-dao.eth"
	rewardsTreeFilenameFormat          string = "rp-rewards-%s-%d%s"
	minipoolPerformanceFilenameFormat  string = "rp-minipool-performance-%s-%d%s"
	performanceFilenameFormat          string = "rp-performance-%s-%d%s"
	RewardsTreeIpfsExtension           string = ".zst"
	RewardsTreesFolder                 string = "rewards-trees"
	ChecksumTableFilename              string = "checksums.sha384"
	DaemonDataPath                     string = "/.rocketpool/data"
	WatchtowerFolder                   string = "watchtower"
	WatchtowerStateFile                string = "state.yml"
	RegenerateRewardsTreeRequestSuffix string = ".request"
	RegenerateRewardsTreeRequestFormat string = "%d" + RegenerateRewardsTreeRequestSuffix
	PrimaryRewardsFileUrl              string = "https://%s.ipfs.dweb.link/%s"
	SecondaryRewardsFileUrl            string = "https://ipfs.io/ipfs/%s/%s"
	GithubRewardsFileUrl               string = "https://github.com/rocket-pool/rewards-trees/raw/main/%s/%s"
	GlobalFeeRecipientFilename         string = "rp-fee-recipient.txt"
	NativeFeeRecipientFilename         string = "rp-fee-recipient-env.txt"
	PerKeyFeeRecipientFilename         string = "rp-fee-recipient-per-key"
)

// Defaults
const (
	defaultProjectName       string = "rocketpool"
	WatchtowerMaxFeeDefault  uint64 = 50
	WatchtowerPrioFeeDefault uint64 = 3
)

type RewardsExtension string

const (
	RewardsExtensionJSON RewardsExtension = ".json"
	RewardsExtensionSSZ  RewardsExtension = ".ssz"
)

// Contract addresses for multicall / network state manager
type StateManagerContracts struct {
	Multicaller    common.Address
	BalanceBatcher common.Address
}

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

	// Timestamp used as reference for prices/balances submissions
	PriceBalanceSubmissionReferenceTimestamp config.Parameter `yaml:"priceBalanceSubmissionReferenceTimestamp,omitempty"`

	// Custom URL to download a rewards tree
	RewardsTreeCustomUrl config.Parameter `yaml:"rewardsTreeCustomUrl,omitempty"`

	// URL for an EC with archive mode, for manual rewards tree generation
	ArchiveECUrl config.Parameter `yaml:"archiveEcUrl,omitempty"`

	// Manual override for the watchtower's max fee
	WatchtowerMaxFeeOverride config.Parameter `yaml:"watchtowerMaxFeeOverride,omitempty"`

	// Manual override for the watchtower's priority fee
	WatchtowerPrioFeeOverride config.Parameter `yaml:"watchtowerPrioFeeOverride,omitempty"`

	// The toggle for enabling pDAO proposal verification duties
	VerifyProposals config.Parameter `yaml:"verifyProposals,omitempty"`

	// Delay for automatic queue assignment
	AutoAssignmentDelay config.Parameter `yaml:"autoAssignmentDelay,omitempty"`

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

	// The contract address of RocketSignerRegistry
	rocketSignerRegistryAddress map[config.Network]string `yaml:"-"`

	// The contract address of the RPL token
	rplTokenAddress map[config.Network]string `yaml:"-"`

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

	// The contract address of rocketNetworkPrices from v1.2.0
	v1_2_0_NetworkPricesAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketNetworkBalances from v1.2.0
	v1_2_0_NetworkBalancesAddress map[config.Network]string `yaml:"-"`

	// Addresses for RocketRewardsPool that have been upgraded during development
	previousRewardsPoolAddresses map[config.Network][]common.Address `yaml:"-"`

	// Addresses for RocketDAOProtocolVerifier that have been upgraded during development
	previousRocketDAOProtocolVerifier map[config.Network][]common.Address `yaml:"-"`

	// The RocketOvmPriceMessenger Optimism address for each network
	optimismPriceMessengerAddress map[config.Network]string `yaml:"-"`

	// The RocketPolygonPriceMessenger Polygon address for each network
	polygonPriceMessengerAddress map[config.Network]string `yaml:"-"`

	// The RocketArbitumPriceMessenger Arbitrum address for each network
	arbitrumPriceMessengerAddress map[config.Network]string `yaml:"-"`

	// The RocketArbitumPriceMessengerV2 Arbitrum address for each network
	arbitrumPriceMessengerAddressV2 map[config.Network]string `yaml:"-"`

	// The RocketZkSyncPriceMessenger zkSyncEra address for each network
	zkSyncEraPriceMessengerAddress map[config.Network]string `yaml:"-"`

	// The RocketBasePriceMessenger Base address for each network
	basePriceMessengerAddress map[config.Network]string `yaml:"-"`

	// The RocketScrollPriceMessenger Scroll address for each network
	scrollPriceMessengerAddress map[config.Network]string `yaml:"-"`

	// The Scroll L2 message fee estimator address for each network
	scrollFeeEstimatorAddress map[config.Network]string `yaml:"-"`

	// The UniswapV3 pool address for each network (used for RPL price TWAP info)
	rplTwapPoolAddress map[config.Network]string `yaml:"-"`

	// The multicall contract address
	multicallAddress map[config.Network]string `yaml:"-"`

	// The BalanceChecker contract address
	balancebatcherAddress map[config.Network]string `yaml:"-"`

	// The FlashBots Protect RPC endpoint
	flashbotsProtectUrl map[config.Network]string `yaml:"-"`
}

// Generates a newSmart Node configuration
func NewSmartnodeConfig(cfg *RocketPoolConfig) *SmartnodeConfig {

	return &SmartnodeConfig{
		Title:  "Smartnode Settings",
		parent: cfg,

		ProjectName: config.Parameter{
			ID:                 ProjectNameID,
			Name:               "Project Name",
			Description:        "This is the prefix that will be attached to all of the Docker containers managed by the Smart Node.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: defaultProjectName},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth1, config.ContainerID_Eth2, config.ContainerID_Validator, config.ContainerID_Grafana, config.ContainerID_Prometheus, config.ContainerID_Exporter},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		DataPath: config.Parameter{
			ID:                 "dataPath",
			Name:               "Data Path",
			Description:        "The absolute path of the `data` folder that contains your node wallet's encrypted file, the password for your node wallet, and all of the validator keys for your minipools. You may use environment variables in this string.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: getDefaultDataDir(cfg)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		WatchtowerStatePath: config.Parameter{
			ID:                 "watchtowerPath",
			Name:               "Watchtower Path",
			Description:        "The absolute path of the watchtower state folder that contains persistent state that is used by the watchtower process on trusted nodes. **Only relevant for trusted nodes.**",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: "$HOME/.rocketpool/watchtower"},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		Network: config.Parameter{
			ID:                 NetworkID,
			Name:               "Network",
			Description:        "The Ethereum network you want to use - select Hoodi Testnet to practice with fake ETH, or Mainnet to stake on the real network using real ETH.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.Network_Mainnet},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth1, config.ContainerID_Eth2, config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            getNetworkOptions(),
		},

		ManualMaxFee: config.Parameter{
			ID:                 "manualMaxFee",
			Name:               "Manual Max Fee",
			Description:        "Set this if you want all of the Smart Node's transactions to use this specific max fee value (in gwei), which is the most you'd be willing to pay (*including the priority fee*).\n\nA value of 0 will show you the current suggested max fee based on the current network conditions and let you specify it each time you do a transaction.\n\nAny other value will ignore the recommended max fee and explicitly use this value instead.\n\nThis applies to automated transactions (such as claiming RPL and staking minipools) as well.",
			Type:               config.ParameterType_Float,
			Default:            map[config.Network]interface{}{config.Network_All: float64(0)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		PriorityFee: config.Parameter{
			ID:                 "priorityFee",
			Name:               "Priority Fee",
			Description:        "The default value for the priority fee (in gwei) for all of your transactions. This describes how much you're willing to pay *above the network's current base fee* - the higher this is, the more ETH you give to the validators for including your transaction, which generally means it will be included in a block faster (as long as your max fee is sufficiently high to cover the current network conditions).\n\nMust be larger than 0.",
			Type:               config.ParameterType_Float,
			Default:            map[config.Network]interface{}{config.Network_All: float64(2)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		AutoTxGasThreshold: config.Parameter{
			ID:   "minipoolStakeGasThreshold",
			Name: "Automatic TX Gas Threshold",
			Description: "Occasionally, the Smart Node will attempt to perform some automatic transactions (such as the second `stake` transaction to finish launching a minipool or the `reduce bond` transaction to convert a 16-ETH minipool to an 8-ETH one). During these, your node will use the `Rapid` suggestion from the gas estimator as its max fee.\n\nThis threshold is a limit (in gwei) you can put on that suggestion; your node will not `stake` the new minipool until the suggestion is below this limit.\n\n" +
				"A value of 0 will disable non-essential automatic transactions (such as minipool balance distribution and bond reduction), but essential transactions (such as minipool staking and solo migration promotion) will not be disabled.\n\n" +
				"NOTE: the node will ignore this limit and automatically execute transactions at whatever the suggested fee happens to be once too much time has passed since those transactions were first eligible. You may end up paying more than you wanted to if you set this too low!",
			Type:               config.ParameterType_Float,
			Default:            map[config.Network]interface{}{config.Network_All: float64(150)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		DistributeThreshold: config.Parameter{
			ID:                 "distributeThreshold",
			Name:               "Auto-Distribute Threshold",
			Description:        "The Smart Node will regularly check the balance of each of your minipools on the Execution Layer (**not** the Beacon Chain).\nIf any of them have a balance greater than this threshold (in ETH), the Smart Node will automatically distribute the balance. This will send your share of the balance to your withdrawal address.\n\nMust be less than 8 ETH.\n\nSet this to 0 to disable automatic distributes.\n[orange]WARNING: if you disable automatic distribution, you **must** ensure you distribute your minipool's balance before it reaches 8 ETH or you will no longer be able to distribute your rewards until you exit the minipool!",
			Type:               config.ParameterType_Float,
			Default:            map[config.Network]interface{}{config.Network_All: float64(1)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		VerifyProposals: config.Parameter{
			ID:                 "verifyProposals",
			Name:               "Enable PDAO Proposal Checker",
			Description:        "Check this box to opt into the responsibility for verifying Protocol DAO proposals once the Houston upgrade has been activated. Your node will regularly check for new proposals, verify their correctness, and submit challenges to any that do not match the on-chain data (e.g., if someone tampered with voting power and attempted to cheat).\n\nTo learn more about the PDAO proposal checking duty, including requirements and RPL bonding, please see the documentation at https://docs.rocketpool.net/guides/houston/pdao#challenge-process.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		AutoAssignmentDelay: config.Parameter{
			ID:                 "autoAssignmentDelay",
			Name:               "Automatic queue assignment delay",
			Description:        "The Smart Node will periodically check whether its megapool is next in the queue. It will wait for the number of hours specified by this parameter after the last assignment before performing the assignment automatically.\n\n",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: uint16(48)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		RewardsTreeMode: config.Parameter{
			ID:                 "rewardsTreeMode",
			Name:               "Rewards Tree Mode",
			Description:        "Select how you want to acquire the Merkle Tree files for each rewards interval.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.RewardsMode_Download},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
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

		PriceBalanceSubmissionReferenceTimestamp: config.Parameter{
			ID:                 "priceBalanceSubmissionReferenceTimestamp",
			Name:               "P/B Submission Time Ref",
			Description:        "Prices and balances submission time reference. An Unix timestamp used by oDAO members as an initial reference to calculate when submissions are due based on the onchain stored submission interval value.",
			Type:               config.ParameterType_Int,
			Default:            map[config.Network]interface{}{config.Network_All: int64(config.PBSubmission_6AM)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		RewardsTreeCustomUrl: config.Parameter{
			ID:                 "rewardsTreeCustomUrl",
			Name:               "Rewards Tree Custom Download URLs",
			Description:        "The Smart Node will automatically download missing rewards tree files from trusted sources like IPFS and Rocket Pool's repository on GitHub. Use this field if you would like to manually specify additional sources that host the rewards tree files, so the Smart Node can download from them as well.\nMultiple URLs can be provided using ';' as separator).\n\nUse '%s' to specify the location of the rewards file name in the URL - for example: `https://my-cool-domain.com/rewards-trees/mainnet/%s`.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		ArchiveECUrl: config.Parameter{
			ID:                 "archiveECUrl",
			Name:               "Archive-Mode EC URL",
			Description:        "[orange]**For manual Merkle rewards tree generation only.**[white]\n\nGenerating the Merkle rewards tree files for past rewards intervals typically requires an Execution client with Archive mode enabled, which is usually disabled on your primary and fallback Execution clients to save disk space.\nIf you want to generate your own rewards tree files for intervals from a long time ago, you may enter the URL of an Execution client with Archive access here.\n\nFor a free light client with Archive access, you may use https://www.alchemy.com/supernode.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		WatchtowerMaxFeeOverride: config.Parameter{
			ID:                 "watchtowerMaxFeeOverride",
			Name:               "Watchtower Max Fee Override",
			Description:        fmt.Sprintf("[orange]**For Oracle DAO members only.**\n\n[white]Use this to override the max fee (in gwei) for watchtower transactions. Note that if you set it below %d, the setting will be ignored; it can only be used to set the max fee higher than %d during times of extreme network stress.", WatchtowerMaxFeeDefault, WatchtowerMaxFeeDefault),
			Type:               config.ParameterType_Float,
			Default:            map[config.Network]interface{}{config.Network_All: float64(WatchtowerMaxFeeDefault)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		WatchtowerPrioFeeOverride: config.Parameter{
			ID:                 "watchtowerPrioFeeOverride",
			Name:               "Watchtower Priority Fee Override",
			Description:        fmt.Sprintf("[orange]**For Oracle DAO members only.**\n\n[white]Use this to override the priority fee (in gwei) for watchtower transactions. Note that if you set it below %d, the setting will be ignored; it can only be used to set the priority fee higher than %d during times of extreme network stress.", WatchtowerPrioFeeDefault, WatchtowerPrioFeeDefault),
			Type:               config.ParameterType_Float,
			Default:            map[config.Network]interface{}{config.Network_All: float64(WatchtowerPrioFeeDefault)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		txWatchUrl: map[config.Network]string{
			config.Network_Mainnet: "https://etherscan.io/tx",
			config.Network_Devnet:  "https://hoodi.etherscan.io/tx",
			config.Network_Testnet: "https://hoodi.etherscan.io/tx",
		},

		stakeUrl: map[config.Network]string{
			config.Network_Mainnet: "https://stake.rocketpool.net",
			config.Network_Devnet:  "https://devnet.rocketpool.net",
			config.Network_Testnet: "https://testnet.rocketpool.net",
		},

		chainID: map[config.Network]uint{
			config.Network_Mainnet: 1,      // Mainnet
			config.Network_Devnet:  560048, // Hoodi
			config.Network_Testnet: 560048, // Hoodi
		},

		storageAddress: map[config.Network]string{
			config.Network_Mainnet: "0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46",
			config.Network_Devnet:  "0x990BC2c12d2a39e5FD92111B98A728bf39742478",
			config.Network_Testnet: "0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1",
		},

		rocketSignerRegistryAddress: map[config.Network]string{
			config.Network_Mainnet: "0xc1062617d10Ae99E09D941b60746182A87eAB38F",
			config.Network_Devnet:  "0xE3FbfaD4A11777E6271921E7EC1A5a1345684F4E",
			config.Network_Testnet: "0xE3FbfaD4A11777E6271921E7EC1A5a1345684F4E",
		},

		rplTokenAddress: map[config.Network]string{
			config.Network_Mainnet: "0xD33526068D116cE69F19A9ee46F0bd304F21A51f",
			config.Network_Devnet:  "0xce59520Cbaec3B399a5245e72C0F21df791202FE",
			config.Network_Testnet: "0x1Cc9cF5586522c6F483E84A19c3C2B0B6d027bF0",
		},

		rethAddress: map[config.Network]string{
			config.Network_Mainnet: "0xae78736Cd615f374D3085123A210448E74Fc6393",
			config.Network_Devnet:  "0x3aC886F531BEb95f08F73fDb21528BE3c63AA82F",
			config.Network_Testnet: "0x7322c24752f79c05FFD1E2a6FCB97020C1C264F1",
		},

		v1_0_0_RewardsPoolAddress: map[config.Network]string{
			config.Network_Mainnet: "0xA3a18348e6E2d3897B6f2671bb8c120e36554802",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		v1_0_0_ClaimNodeAddress: map[config.Network]string{
			config.Network_Mainnet: "0x899336A2a86053705E65dB61f52C686dcFaeF548",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		v1_0_0_ClaimTrustedNodeAddress: map[config.Network]string{
			config.Network_Mainnet: "0x6af730deB0463b432433318dC8002C0A4e9315e8",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		v1_0_0_MinipoolManagerAddress: map[config.Network]string{
			config.Network_Mainnet: "0x6293B8abC1F36aFB22406Be5f96D893072A8cF3a",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		v1_1_0_NetworkPricesAddress: map[config.Network]string{
			config.Network_Mainnet: "0xd3f500F550F46e504A4D2153127B47e007e11166",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		v1_1_0_NodeStakingAddress: map[config.Network]string{
			config.Network_Mainnet: "0x0d8D8f8541B12A0e1194B7CC4b6D954b90AB82ec",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		v1_1_0_NodeDepositAddress: map[config.Network]string{
			config.Network_Mainnet: "0x1Cc9cF5586522c6F483E84A19c3C2B0B6d027bF0",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		v1_1_0_MinipoolQueueAddress: map[config.Network]string{
			config.Network_Mainnet: "0x5870dA524635D1310Dc0e6F256Ce331012C9C19E",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		v1_1_0_MinipoolFactoryAddress: map[config.Network]string{
			config.Network_Mainnet: "0x54705f80D7C51Fcffd9C659ce3f3C9a7dCCf5788",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		v1_2_0_NetworkPricesAddress: map[config.Network]string{
			config.Network_Mainnet: "0x751826b107672360b764327631cC5764515fFC37",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		v1_2_0_NetworkBalancesAddress: map[config.Network]string{
			config.Network_Mainnet: "0x07FCaBCbe4ff0d80c2b1eb42855C0131b6cba2F4",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		snapshotApiDomain: map[config.Network]string{
			config.Network_Mainnet: "hub.snapshot.org",
			config.Network_Devnet:  "hub.snapshot.org",
			config.Network_Testnet: "hub.snapshot.org",
		},

		previousRewardsPoolAddresses: map[config.Network][]common.Address{
			config.Network_Mainnet: {
				common.HexToAddress("0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1"),
				common.HexToAddress("0xA805d68b61956BC92d556F2bE6d18747adAeEe82"),
			},
			config.Network_Devnet: {
				common.HexToAddress("0x556791EC1aa443df339E340E6f20d06a1cD21583"),
			},
			config.Network_Testnet: {},
		},

		previousRocketDAOProtocolVerifier: map[config.Network][]common.Address{
			config.Network_Mainnet: {},
			config.Network_Devnet:  {},
			config.Network_Testnet: {},
		},

		optimismPriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0x12759f8Df234f8f2cDdb3d2Ed5604adF9ACCfc9F",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		polygonPriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0xb1029Ac2Be4e08516697093e2AFeC435057f3511",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		arbitrumPriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0x05330300f829AD3fC8f33838BC88CFC4093baD53",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		arbitrumPriceMessengerAddressV2: map[config.Network]string{
			config.Network_Mainnet: "0x312FcFB03eC9B1Ea38CB7BFCd26ee7bC3b505aB1",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		zkSyncEraPriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0x6cf6CB29754aEBf88AF12089224429bD68b0b8c8",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		basePriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0x8aa4afc5a9793433eb37c9919ff49b54903c7cb1",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		scrollPriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0x0f22dc9b9c03757d4676539203d7549c8f22c15c",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		scrollFeeEstimatorAddress: map[config.Network]string{
			config.Network_Mainnet: "0x0d7E906BD9cAFa154b048cFa766Cc1E54E39AF9B",
			config.Network_Devnet:  "",
			config.Network_Testnet: "",
		},

		rplTwapPoolAddress: map[config.Network]string{
			config.Network_Mainnet: "0xe42318ea3b998e8355a3da364eb9d48ec725eb45",
			config.Network_Devnet:  "0x0ca239d8AC5E49E3203d60eaf86Baa6712E5b454",
			config.Network_Testnet: "0x0ca239d8AC5E49E3203d60eaf86Baa6712E5b454",
		},

		multicallAddress: map[config.Network]string{
			config.Network_Mainnet: "0x5BA1e12693Dc8F9c48aAD8770482f4739bEeD696",
			config.Network_Devnet:  "0xc5fA61aA6Ec012d1A2Ea38f31ADAf4D06c8725E7",
			config.Network_Testnet: "0xc5fA61aA6Ec012d1A2Ea38f31ADAf4D06c8725E7",
		},

		balancebatcherAddress: map[config.Network]string{
			config.Network_Mainnet: "0xb1f8e55c7f64d203c1400b9d8555d050f94adf39",
			config.Network_Devnet:  "0xB80b500CF68a956b6f149F1C48E8F07EEF4486Ce",
			config.Network_Testnet: "0xB80b500CF68a956b6f149F1C48E8F07EEF4486Ce",
		},

		flashbotsProtectUrl: map[config.Network]string{
			config.Network_Mainnet: "https://rpc.flashbots.net/",
			config.Network_Devnet:  "https://rpc-hoodi.flashbots.net/",
			config.Network_Testnet: "https://rpc-hoodi.flashbots.net/",
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
		&cfg.VerifyProposals,
		&cfg.AutoAssignmentDelay,
		&cfg.RewardsTreeMode,
		&cfg.PriceBalanceSubmissionReferenceTimestamp,
		&cfg.RewardsTreeCustomUrl,
		&cfg.ArchiveECUrl,
		&cfg.WatchtowerMaxFeeOverride,
		&cfg.WatchtowerPrioFeeOverride,
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

func (cfg *SmartnodeConfig) GetNodeAddressPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "address")
	}

	return filepath.Join(DaemonDataPath, "address")
}

func (cfg *SmartnodeConfig) GetValidatorKeychainPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "validators")
	}

	return filepath.Join(DaemonDataPath, "validators")
}

func (cfg *SmartnodeConfig) GetRecordsPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "records")
	}

	return filepath.Join(DaemonDataPath, "records")
}

func (cfg *SmartnodeConfig) GetVotingPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "voting", string(cfg.Network.Value.(config.Network)))
	}

	return filepath.Join(DaemonDataPath, "voting", string(cfg.Network.Value.(config.Network)))
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

func (cfg *SmartnodeConfig) GetRocketSignerRegistryAddress() string {
	return cfg.rocketSignerRegistryAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetRplTokenAddress() string {
	return cfg.rplTokenAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetSmartnodeContainerTag() string {
	return smartnodeTagPrefix + shared.RocketPoolVersion()
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

// The title for the config
func (cfg *SmartnodeConfig) GetConfigTitle() string {
	return cfg.Title
}

func (cfg *SmartnodeConfig) GetRethAddress() common.Address {
	return common.HexToAddress(cfg.rethAddress[cfg.Network.Value.(config.Network)])
}

func getDefaultDataDir(config *RocketPoolConfig) string {
	if config == nil {
		// Handle tests. Eventually we'll refactor so this isn't necessary.
		return ""
	}
	return filepath.Join(config.RocketPoolDirectory, "data")
}

func (cfg *SmartnodeConfig) GetRewardsTreeDirectory(daemon bool) string {
	if daemon && !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, RewardsTreesFolder)
	}

	return filepath.Join(cfg.DataPath.Value.(string), RewardsTreesFolder)
}

func (cfg *SmartnodeConfig) formatRewardsFilename(f string, interval uint64, extension RewardsExtension) string {
	return fmt.Sprintf(f, string(cfg.Network.Value.(config.Network)), interval, string(extension))
}

func (cfg *SmartnodeConfig) GetRewardsTreeFilename(interval uint64, extension RewardsExtension) string {
	return cfg.formatRewardsFilename(rewardsTreeFilenameFormat, interval, extension)
}

func (cfg *SmartnodeConfig) GetMinipoolPerformanceFilename(interval uint64) string {
	return cfg.formatRewardsFilename(minipoolPerformanceFilenameFormat, interval, RewardsExtensionJSON)
}

func (cfg *SmartnodeConfig) GetPerformanceFilename(interval uint64) string {
	return cfg.formatRewardsFilename(performanceFilenameFormat, interval, RewardsExtensionJSON)
}

func (cfg *SmartnodeConfig) GetRewardsTreePath(interval uint64, daemon bool, extension RewardsExtension) string {
	return filepath.Join(
		cfg.GetRewardsTreeDirectory(daemon),
		cfg.GetRewardsTreeFilename(interval, extension),
	)
}

func (cfg *SmartnodeConfig) GetMinipoolPerformancePath(interval uint64, daemon bool) string {
	return filepath.Join(
		cfg.GetRewardsTreeDirectory(daemon),
		cfg.GetMinipoolPerformanceFilename(interval),
	)
}

func (cfg *SmartnodeConfig) GetPerformancePath(interval uint64) string {
	return filepath.Join(
		cfg.GetRewardsTreeDirectory(true),
		cfg.GetPerformanceFilename(interval),
	)
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

func (cfg *SmartnodeConfig) GetGlobalFeeRecipientFilePath() string {
	if !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, "validators", GlobalFeeRecipientFilename)
	}

	return filepath.Join(cfg.DataPath.Value.(string), "validators", NativeFeeRecipientFilename)
}

func (cfg *SmartnodeConfig) GetPerKeyFeeRecipientFilePath() string {
	if !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, "validators", PerKeyFeeRecipientFilename)
	}

	return filepath.Join(cfg.DataPath.Value.(string), "validators", PerKeyFeeRecipientFilename)
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

func (cfg *SmartnodeConfig) GetV120NetworkPricesAddress() common.Address {
	return common.HexToAddress(cfg.v1_2_0_NetworkPricesAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetV120NetworkBalancesAddress() common.Address {
	return common.HexToAddress(cfg.v1_2_0_NetworkBalancesAddress[cfg.Network.Value.(config.Network)])
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

func (cfg *SmartnodeConfig) GetPreviousRewardsPoolAddresses() []common.Address {
	return cfg.previousRewardsPoolAddresses[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetPreviousRocketDAOProtocolVerifierAddresses() []common.Address {
	return cfg.previousRocketDAOProtocolVerifier[cfg.Network.Value.(config.Network)]
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

func (cfg *SmartnodeConfig) GetArbitrumMessengerAddressV2() string {
	return cfg.arbitrumPriceMessengerAddressV2[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetZkSyncEraMessengerAddress() string {
	return cfg.zkSyncEraPriceMessengerAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetBaseMessengerAddress() string {
	return cfg.basePriceMessengerAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetScrollMessengerAddress() string {
	return cfg.scrollPriceMessengerAddress[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetScrollFeeEstimatorAddress() string {
	return cfg.scrollFeeEstimatorAddress[cfg.Network.Value.(config.Network)]
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

// Utility function to get the state manager contracts
func (cfg *SmartnodeConfig) GetStateManagerContracts() StateManagerContracts {
	return StateManagerContracts{
		Multicaller:    common.HexToAddress(cfg.GetMulticallAddress()),
		BalanceBatcher: common.HexToAddress(cfg.GetBalanceBatcherAddress()),
	}
}

func (cfg *SmartnodeConfig) GetFlashbotsProtectUrl() string {
	return cfg.flashbotsProtectUrl[cfg.Network.Value.(config.Network)]
}

func getNetworkOptions() []config.ParameterOption {
	options := []config.ParameterOption{
		{
			Name:        "Ethereum Mainnet",
			Description: "This is the real Ethereum main network, using real ETH and real RPL to make real validators.",
			Value:       config.Network_Mainnet,
		}, {
			Name:        "Hoodi Testnet",
			Description: "This is the Hoodi test network, which is the next generation of long-lived testnets for Ethereum. It uses free fake ETH and free fake RPL to make fake validators.\nUse this if you want to practice running the Smart Node in a free, safe environment before moving to Mainnet.",
			Value:       config.Network_Testnet,
		},
	}

	if strings.Contains(shared.RocketPoolVersion(), "dev") {
		options = append(options, config.ParameterOption{
			Name:        "Devnet",
			Description: "This is a development network used by Rocket Pool engineers to test new features and contract upgrades before they are promoted to a Testnet for staging. You should not use this network unless invited to do so by the developers.",
			Value:       config.Network_Devnet,
		})
	}

	return options
}
