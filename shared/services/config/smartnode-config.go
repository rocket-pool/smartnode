package config

import (
	"fmt"
	"path/filepath"

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
	FeeRecipientFilename               string = "rp-fee-recipient.txt"
	NativeFeeRecipientFilename         string = "rp-fee-recipient-env.txt"
)

// Defaults
const defaultProjectName string = "rocketpool"

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

	// The command for restarting the validator container in native mode
	ValidatorRestartCommand config.Parameter `yaml:"validatorRestartCommand,omitempty"`

	// Which network we're on
	Network config.Parameter `yaml:"network,omitempty"`

	// The terminal total difficulty override for the Merge
	TTD config.Parameter `yaml:"ttd,omitempty"`

	// Manual max fee override
	ManualMaxFee config.Parameter `yaml:"manualMaxFee,omitempty"`

	// Manual priority fee override
	PriorityFee config.Parameter `yaml:"priorityFee,omitempty"`

	// Threshold for auto RPL claims
	RplClaimGasThreshold config.Parameter `yaml:"rplClaimGasThreshold,omitempty"`

	// Threshold for auto minipool stakes
	MinipoolStakeGasThreshold config.Parameter `yaml:"minipoolStakeGasThreshold,omitempty"`

	// Mode for acquiring Merkle rewards trees
	RewardsTreeMode config.Parameter `yaml:"rewardsTreeMode,omitempty"`

	// URL for an EC with archive mode, for manual rewards tree generation
	ArchiveECUrl config.Parameter `yaml:"archiveEcUrl,omitempty"`

	// Token for Oracle DAO members to use when uploading Merkle trees to Web3.Storage
	Web3StorageApiToken config.Parameter `yaml:"web3StorageApiToken,omitempty"`

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
	snapshotApiDomain map[Network]string `yaml:"-"`

	// The contract address of rETH
	rethAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketRewardsPool from v1.0.0
	legacyRewardsPoolAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketClaimNode from v1.0.0
	legacyClaimNodeAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketClaimTrustedNode from v1.0.0
	legacyClaimTrustedNodeAddress map[config.Network]string `yaml:"-"`

	// The contract address of rocketMinipoolManager from v1.0.0
	legacyMinipoolManagerAddress map[config.Network]string `yaml:"-"`

	// Addresses for RocketRewardsPool that have been upgraded during development
	previousRewardsPoolAddresses map[config.Network]map[string][]common.Address `yaml:"-"`

	// The RocketOvmPriceMessenger address for each network
	optimismPriceMessengerAddress map[config.Network]string `yaml:"-"`
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

		TTD: config.Parameter{
			ID:                   "ttd",
			Name:                 "TTD Override",
			Description:          "Use this to manually override the terminal total difficulty value for the network. This is the number used by the Execution and Consensus clients to know when to trigger The Merge.\n\nNOTE: This should only be used in special situations where the Core Developers have felt it necessary to change the TTD from the previously-agreed-upon value.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1, config.ContainerID_Eth2},
			EnvironmentVariables: []string{"TTD_OVERRIDE"},
			CanBeBlank:           true,
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
			Options: []config.ParameterOption{
				{
					Name:        "Ethereum Mainnet",
					Description: "This is the real Ethereum main network, using real ETH and real RPL to make real validators.",
					Value:       config.Network_Mainnet,
				}, {
					Name:        "Prater Testnet",
					Description: "This is the Prater test network, using free fake ETH and free fake RPL to make fake validators.\nUse this if you want to practice running the Smartnode in a free, safe environment before moving to Mainnet.",
					Value:       config.Network_Prater,
				}, /*{
					Name:        "Kiln Testnet",
					Description: "This is the Kiln test network, which uses free \"test\" ETH and free \"test\" RPL.\n\nUse this if you want to practice running a node on a post-merge network to learn how it differs from Mainnet today.",
					Value:       config.Network_Kiln,
				},{
					Name:        "Ropsten Testnet",
					Description: "This is the Ropsten test network, which uses free \"test\" ETH and free \"test\" RPL.\n\nUse this if you want to practice running a node on a post-merge network to learn how it differs from Mainnet today.",
					Value:       config.Network_Ropsten,
				}*/},
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
			Description:          "The default value for the priority fee (in gwei) for all of your transactions. This describes how much you're willing to pay *above the network's current base fee* - the higher this is, the more ETH you give to the miners for including your transaction, which generally means it will be mined faster (as long as your max fee is sufficiently high to cover the current network conditions).\n\nMust be larger than 0.",
			Type:                 config.ParameterType_Float,
			Default:              map[config.Network]interface{}{config.Network_All: float64(2)},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		RplClaimGasThreshold: config.Parameter{
			ID:                   "rplClaimGasThreshold",
			Name:                 "RPL Claim Gas Threshold",
			Description:          "Automatic RPL rewards claims will use the `Rapid` suggestion from the gas estimator, based on current network conditions. This threshold is a limit (in gwei) you can put on that suggestion; your node will not try to claim RPL rewards automatically until the suggestion is below this limit.",
			Type:                 config.ParameterType_Float,
			Default:              map[config.Network]interface{}{config.Network_All: float64(150)},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		MinipoolStakeGasThreshold: config.Parameter{
			ID:   "minipoolStakeGasThreshold",
			Name: "Minipool Stake Gas Threshold",
			Description: "Once a newly created minipool passes the scrub check and is ready to perform its second 16 ETH deposit (the `stake` transaction), your node will try to do so automatically using the `Rapid` suggestion from the gas estimator as its max fee. This threshold is a limit (in gwei) you can put on that suggestion; your node will not `stake` the new minipool until the suggestion is below this limit.\n\n" +
				"Note that to ensure your minipool does not get dissolved, the node will ignore this limit and automatically execute the `stake` transaction at whatever the suggested fee happens to be once too much time has passed since its first deposit (currently 7 days).",
			Type:                 config.ParameterType_Float,
			Default:              map[config.Network]interface{}{config.Network_All: float64(150)},
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

		txWatchUrl: map[config.Network]string{
			config.Network_Mainnet: "https://etherscan.io/tx",
			config.Network_Prater:  "https://goerli.etherscan.io/tx",
			config.Network_Kiln:    "TBD",
			config.Network_Ropsten: "https://ropsten.etherscan.io/tx",
		},

		stakeUrl: map[config.Network]string{
			config.Network_Mainnet: "https://stake.rocketpool.net",
			config.Network_Prater:  "https://testnet.rocketpool.net",
			config.Network_Kiln:    "TBD",
			config.Network_Ropsten: "TBD",
		},

		chainID: map[config.Network]uint{
			config.Network_Mainnet: 1, // Mainnet
			config.Network_Prater:  5, // Goerli
			config.Network_Kiln:    0x1469ca,
			config.Network_Ropsten: 3,
		},

		storageAddress: map[config.Network]string{
			config.Network_Mainnet: "0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46",
			config.Network_Prater:  "0xd8Cd47263414aFEca62d6e2a3917d6600abDceB3",
			config.Network_Kiln:    "0x340b8bB313B9d5d79cA82c7b7416Ee1A8485326c",
			config.Network_Ropsten: "0x2b43C0b8e44FD03aa5C2866ba380fBcB646A1D42",
		},

		oneInchOracleAddress: map[config.Network]string{
			config.Network_Mainnet: "0x07D91f5fb9Bf7798734C3f606dB065549F6893bb",
			config.Network_Prater:  "0x4eDC966Df24264C9C817295a0753804EcC46Dd22",
			config.Network_Kiln:    "0xd46a870139F348C3d2596470c355E4BE26b03071",
			config.Network_Ropsten: "0xd46a870139f348c3d2596470c355e4be26b03071",
		},

		rplTokenAddress: map[config.Network]string{
			config.Network_Mainnet: "0xD33526068D116cE69F19A9ee46F0bd304F21A51f",
			config.Network_Prater:  "0x5e932688e81a182e3de211db6544f98b8e4f89c7",
			config.Network_Kiln:    "0xb4efd85c19999d84251304bda99e90b92300bd93",
			config.Network_Ropsten: "0xb4efd85c19999d84251304bda99e90b92300bd93",
		},

		rplFaucetAddress: map[config.Network]string{
			config.Network_Mainnet: "",
			config.Network_Prater:  "0x95D6b8E2106E3B30a72fC87e2B56ce15E37853F9",
			config.Network_Kiln:    "0xE200d83cE0534fB28F12813617abCE01258b6673",
			config.Network_Ropsten: "0x94e840CD58FF0F387117869A28035903a51502aa",
		},

		rethAddress: map[config.Network]string{
			config.Network_Mainnet: "0xae78736Cd615f374D3085123A210448E74Fc6393",
			config.Network_Prater:  "0x178E141a0E3b34152f73Ff610437A7bf9B83267A",
			config.Network_Kiln:    "0xBCAB713355D265B3e3D64F02bc1Eb4F7717b4B14",
			config.Network_Ropsten: "0x00651FC69eFd13F76fC7dEBC2540F2662A09fa8c",
		},

		legacyRewardsPoolAddress: map[config.Network]string{
			config.Network_Mainnet: "0xA3a18348e6E2d3897B6f2671bb8c120e36554802",
			config.Network_Prater:  "0xf9aE18eB0CE4930Bc3d7d1A5E33e4286d4FB0f8B",
			config.Network_Kiln:    "0xFb62F3B5AF8099Bbd19d5d46084Bb152ECDE25A6",
			config.Network_Ropsten: "0x401e46fA6cBC9e1E6Cc3E9666C10329f938aE1B3",
		},

		legacyClaimNodeAddress: map[config.Network]string{
			config.Network_Mainnet: "0x899336A2a86053705E65dB61f52C686dcFaeF548",
			config.Network_Prater:  "0xc05b7A2a03A6d2736d1D0ebf4d4a0aFE2cc32cE1",
			config.Network_Kiln:    "0xF98086202F8F58dad8120055Fdd6e2f36De2c6Fb",
			config.Network_Ropsten: "0xA55F65219d7254DFde4021E4f534a7a55750C4a1",
		},

		legacyClaimTrustedNodeAddress: map[config.Network]string{
			config.Network_Mainnet: "0x6af730deB0463b432433318dC8002C0A4e9315e8",
			config.Network_Prater:  "0x730982F4439E5AC30292333ff7d0C478907f2219",
			config.Network_Kiln:    "0x2DFD22dD4474C306824d3014C6FC664012577807",
			config.Network_Ropsten: "0x097664F6B68564B3AFEad72662fCA557350aE7Bf",
		},

		legacyMinipoolManagerAddress: map[config.Network]string{
			config.Network_Mainnet: "0x6293B8abC1F36aFB22406Be5f96D893072A8cF3a",
			config.Network_Prater:  "0xB815a94430f08dD2ab61143cE1D5739Ac81D3C6d",
			config.Network_Kiln:    "0x6230e0180bc24cA59D20c56F964C81DcE4fe8df6",
			config.Network_Ropsten: "0xcE2Fc66c58e35bF4c156ab13085826B8d26c4B7A",
		},

		snapshotDelegationAddress: map[config.Network]string{
			config.Network_Mainnet: "0x469788fE6E9E9681C6ebF3bF78e7Fd26Fc015446",
			config.Network_Prater:  "0xD0897D68Cd66A710dDCecDe30F7557972181BEDc",
			config.Network_Kiln:    "",
			config.Network_Ropsten: "0x2588C77829015080C771359eC1C3066d2f1158Db",
		},

		snapshotApiDomain: map[config.Network]string{
			config.Network_Mainnet: "hub.snapshot.org",
			config.Network_Prater:  "testnet.snapshot.org",
			config.Network_Kiln:    "",
			config.Network_Ropsten: "",
		},

		previousRewardsPoolAddresses: map[config.Network]map[string][]common.Address{
			config.Network_Mainnet: {},
			config.Network_Prater: {
				"v1.5.0-rc1": []common.Address{
					common.HexToAddress("0x594Fb75D3dc2DFa0150Ad03F99F97817747dd4E1"),
				},
			},
			config.Network_Kiln: {},
			config.Network_Ropsten: {
				"v1.5.0-rc1": []common.Address{
					common.HexToAddress("0x594fb75d3dc2dfa0150ad03f99f97817747dd4e1"),
					common.HexToAddress("0x762e79b27feE0C0975f9caBEa9E9976006A7aD98"),
				},
			},
		},

		optimismPriceMessengerAddress: map[config.Network]string{
			config.Network_Mainnet: "0xdddcf2c25d50ec22e67218e873d46938650d03a7",
			config.Network_Prater:  "0x87E2deCE7d0A080D579f63cbcD7e1629BEcd7E7d",
			config.Network_Kiln:    "",
			config.Network_Ropsten: "",
		},
	}

}

// Get the parameters for this config
func (cfg *SmartnodeConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.Network,
		&cfg.TTD,
		&cfg.ProjectName,
		&cfg.DataPath,
		&cfg.ManualMaxFee,
		&cfg.PriorityFee,
		&cfg.RplClaimGasThreshold,
		&cfg.MinipoolStakeGasThreshold,
		&cfg.RewardsTreeMode,
		&cfg.ArchiveECUrl,
		&cfg.Web3StorageApiToken,
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
	} else {
		return filepath.Join(DaemonDataPath, "wallet")
	}
}

func (cfg *SmartnodeConfig) GetPasswordPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "password")
	} else {
		return filepath.Join(DaemonDataPath, "password")
	}
}

func (cfg *SmartnodeConfig) GetValidatorKeychainPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "validators")
	} else {
		return filepath.Join(DaemonDataPath, "validators")
	}
}

func (config *SmartnodeConfig) GetWatchtowerStatePath() string {
	if config.parent.IsNativeMode {
		return filepath.Join(config.DataPath.Value.(string), WatchtowerFolder, "state.yml")
	} else {
		return filepath.Join(DaemonDataPath, WatchtowerFolder, "state.yml")
	}
}

func (cfg *SmartnodeConfig) GetCustomKeyPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "custom-keys")
	} else {
		return filepath.Join(DaemonDataPath, "custom-keys")
	}
}

func (cfg *SmartnodeConfig) GetCustomKeyPasswordFilePath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "custom-key-passwords")
	} else {
		return filepath.Join(DaemonDataPath, "custom-key-passwords")
	}
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
	} else {
		return filepath.Join(cfg.DataPath.Value.(string), RewardsTreesFolder, fmt.Sprintf(RewardsTreeFilenameFormat, string(cfg.Network.Value.(config.Network)), interval))
	}
}

func (cfg *SmartnodeConfig) GetMinipoolPerformancePath(interval uint64, daemon bool) string {
	if daemon && !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, RewardsTreesFolder, fmt.Sprintf(MinipoolPerformanceFilenameFormat, string(cfg.Network.Value.(config.Network)), interval))
	} else {
		return filepath.Join(cfg.DataPath.Value.(string), RewardsTreesFolder, fmt.Sprintf(MinipoolPerformanceFilenameFormat, string(cfg.Network.Value.(config.Network)), interval))
	}
}

func (cfg *SmartnodeConfig) GetRegenerateRewardsTreeRequestPath(interval uint64, daemon bool) string {
	if daemon && !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, WatchtowerFolder, fmt.Sprintf(RegenerateRewardsTreeRequestFormat, interval))
	} else {
		return filepath.Join(cfg.DataPath.Value.(string), WatchtowerFolder, fmt.Sprintf(RegenerateRewardsTreeRequestFormat, interval))
	}
}

func (cfg *SmartnodeConfig) GetWatchtowerFolder(daemon bool) string {
	if daemon && !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, WatchtowerFolder)
	} else {
		return filepath.Join(cfg.DataPath.Value.(string), WatchtowerFolder)
	}
}

func (cfg *SmartnodeConfig) GetFeeRecipientFilePath() string {
	if !cfg.parent.IsNativeMode {
		return filepath.Join(DaemonDataPath, "validators", FeeRecipientFilename)
	} else {
		return filepath.Join(cfg.DataPath.Value.(string), "validators", NativeFeeRecipientFilename)
	}
}

func (cfg *SmartnodeConfig) GetLegacyRewardsPoolAddress() common.Address {
	return common.HexToAddress(cfg.legacyRewardsPoolAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetLegacyClaimNodeAddress() common.Address {
	return common.HexToAddress(cfg.legacyClaimNodeAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetLegacyClaimTrustedNodeAddress() common.Address {
	return common.HexToAddress(cfg.legacyClaimTrustedNodeAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetLegacyMinipoolManagerAddress() common.Address {
	return common.HexToAddress(cfg.legacyMinipoolManagerAddress[cfg.Network.Value.(config.Network)])
}

func (cfg *SmartnodeConfig) GetPreviousRewardsPoolAddresses() map[string][]common.Address {
	return cfg.previousRewardsPoolAddresses[cfg.Network.Value.(config.Network)]
}

func (cfg *SmartnodeConfig) GetOptimismMessengerAddress() string {
	return cfg.optimismPriceMessengerAddress[cfg.Network.Value.(config.Network)]
}
