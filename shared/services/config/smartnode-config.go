package config

import (
	"fmt"
	"path/filepath"

	"github.com/rocket-pool/smartnode/shared"
)

// Constants
const (
	smartnodeTag              string = "rocketpool/smartnode:v" + shared.RocketPoolVersion
	powProxyTag               string = "rocketpool/smartnode-pow-proxy:v" + shared.RocketPoolVersion
	pruneProvisionerTag       string = "rocketpool/eth1-prune-provision:v0.0.1"
	NetworkID                 string = "network"
	ProjectNameID             string = "projectName"
	RewardsTreeFilenameFormat string = "rp-rewards-%d.json"
)

// Defaults
const defaultProjectName string = "rocketpool"

// Configuration for the Smartnode
type SmartnodeConfig struct {
	Title string `yaml:"title,omitempty"`

	// The parent config
	parent *RocketPoolConfig `yaml:"-"`

	////////////////////////////
	// User-editable settings //
	////////////////////////////

	// Docker container prefix
	ProjectName Parameter `yaml:"projectName,omitempty"`

	// The path of the data folder where everything is stored
	DataPath Parameter `yaml:"dataPath,omitempty"`

	// Which network we're on
	Network Parameter `yaml:"network,omitempty"`

	// Manual max fee override
	ManualMaxFee Parameter `yaml:"manualMaxFee,omitempty"`

	// Manual priority fee override
	PriorityFee Parameter `yaml:"priorityFee,omitempty"`

	// Threshold for auto RPL claims
	RplClaimGasThreshold Parameter `yaml:"rplClaimGasThreshold,omitempty"`

	// Threshold for auto minipool stakes
	MinipoolStakeGasThreshold Parameter `yaml:"minipoolStakeGasThreshold,omitempty"`

	///////////////////////////
	// Non-editable settings //
	///////////////////////////

	// The URL to provide the user so they can follow pending transactions
	txWatchUrl map[Network]string `yaml:"txWatchUrl,omitempty"`

	// The URL to use for staking rETH
	stakeUrl map[Network]string `yaml:"stakeUrl,omitempty"`

	// The map of networks to execution chain IDs
	chainID map[Network]uint `yaml:"chainID"`

	// The path within the daemon Docker container of the wallet file
	walletPath string `yaml:"walletPath"`

	// The path within the daemon Docker container of the wallet's password file
	passwordPath string `yaml:"passwordPath"`

	// The path within the daemon Docker container of the validator key folder
	validatorKeychainPath string `yaml:"validatorKeychainPath"`

	// The path within the daemon Docker container of the rewards merkle tree folder
	rewardsTreePath string `yaml:"rewardsTreePath"`

	// The contract address of RocketStorage
	storageAddress map[Network]string `yaml:"storageAddress"`

	// The contract address of the 1inch oracle
	oneInchOracleAddress map[Network]string `yaml:"oneInchOracleAddress"`

	// The contract address of the RPL token
	rplTokenAddress map[Network]string `yaml:"rplTokenAddress"`

	// The contract address of the RPL faucet
	rplFaucetAddress map[Network]string `yaml:"rplFaucetAddress"`

	// The contract address of rETH
	rethAddress map[Network]string `yaml:"rethAddress"`
}

// Generates a new Smartnode configuration
func NewSmartnodeConfig(config *RocketPoolConfig) *SmartnodeConfig {

	return &SmartnodeConfig{
		Title:  "Smartnode Settings",
		parent: config,

		ProjectName: Parameter{
			ID:                   ProjectNameID,
			Name:                 "Project Name",
			Description:          "This is the prefix that will be attached to all of the Docker containers managed by the Smartnode.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: defaultProjectName},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2, ContainerID_Validator, ContainerID_Grafana, ContainerID_Prometheus, ContainerID_Exporter},
			EnvironmentVariables: []string{"COMPOSE_PROJECT_NAME"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		DataPath: Parameter{
			ID:                   "dataPath",
			Name:                 "Data Path",
			Description:          "The absolute path of the `data` folder that contains your node wallet's encrypted file, the password for your node wallet, and all of the validator keys for your minipools. You may use environment variables in this string.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: getDefaultDataDir(config)},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Validator},
			EnvironmentVariables: []string{"ROCKETPOOL_DATA_FOLDER"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Network: Parameter{
			ID:                   NetworkID,
			Name:                 "Network",
			Description:          "The Ethereum network you want to use - select Prater Testnet to practice with fake ETH, or Mainnet to stake on the real network using real ETH.",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: Network_Mainnet},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{"NETWORK"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				Name:        "Ethereum Mainnet",
				Description: "This is the real Ethereum main network, using real ETH and real RPL to make real validators.",
				Value:       Network_Mainnet,
			}, {
				Name:        "Prater Testnet",
				Description: "This is the Prater test network, using free fake ETH and free fake RPL to make fake validators.\nUse this if you want to practice running the Smartnode in a free, safe environment before moving to mainnet.",
				Value:       Network_Prater,
			}, {
				Name:        "Kiln Testnet",
				Description: "This is the Kiln test network, which uses free \"test\" ETH and free \"test\" RPL.\n\nUse this if you want to practice running a node on a post-merge network to learn how it differs from Mainnet today.",
				Value:       Network_Kiln,
			}},
		},

		ManualMaxFee: Parameter{
			ID:                   "manualMaxFee",
			Name:                 "Manual Max Fee",
			Description:          "Set this if you want all of the Smartnode's transactions to use this specific max fee value (in gwei), which is the most you'd be willing to pay (*including the priority fee*).\n\nA value of 0 will show you the current suggested max fee based on the current network conditions and let you specify it each time you do a transaction.\n\nAny other value will ignore the recommended max fee and explicitly use this value instead.\n\nThis applies to automated transactions (such as claiming RPL and staking minipools) as well.",
			Type:                 ParameterType_Float,
			Default:              map[Network]interface{}{Network_All: float64(0)},
			AffectsContainers:    []ContainerID{ContainerID_Node, ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		PriorityFee: Parameter{
			ID:                   "priorityFee",
			Name:                 "Priority Fee",
			Description:          "The default value for the priority fee (in gwei) for all of your transactions. This describes how much you're willing to pay *above the network's current base fee* - the higher this is, the more ETH you give to the miners for including your transaction, which generally means it will be mined faster (as long as your max fee is sufficiently high to cover the current network conditions).\n\nMust be larger than 0.",
			Type:                 ParameterType_Float,
			Default:              map[Network]interface{}{Network_All: float64(2)},
			AffectsContainers:    []ContainerID{ContainerID_Node, ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		RplClaimGasThreshold: Parameter{
			ID:                   "rplClaimGasThreshold",
			Name:                 "RPL Claim Gas Threshold",
			Description:          "Automatic RPL rewards claims will use the `Rapid` suggestion from the gas estimator, based on current network conditions. This threshold is a limit (in gwei) you can put on that suggestion; your node will not try to claim RPL rewards automatically until the suggestion is below this limit.",
			Type:                 ParameterType_Float,
			Default:              map[Network]interface{}{Network_All: float64(150)},
			AffectsContainers:    []ContainerID{ContainerID_Node, ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		MinipoolStakeGasThreshold: Parameter{
			ID:   "minipoolStakeGasThreshold",
			Name: "Minipool Stake Gas Threshold",
			Description: "Once a newly created minipool passes the scrub check and is ready to perform its second 16 ETH deposit (the `stake` transaction), your node will try to do so automatically using the `Rapid` suggestion from the gas estimator as its max fee. This threshold is a limit (in gwei) you can put on that suggestion; your node will not `stake` the new minipool until the suggestion is below this limit.\n\n" +
				"Note that to ensure your minipool does not get dissolved, the node will ignore this limit and automatically execute the `stake` transaction at whatever the suggested fee happens to be once too much time has passed since its first deposit (currently 7 days).",
			Type:                 ParameterType_Float,
			Default:              map[Network]interface{}{Network_All: float64(150)},
			AffectsContainers:    []ContainerID{ContainerID_Node},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		txWatchUrl: map[Network]string{
			Network_Mainnet: "https://etherscan.io/tx",
			Network_Prater:  "https://goerli.etherscan.io/tx",
			Network_Kiln:    "TBD",
		},

		stakeUrl: map[Network]string{
			Network_Mainnet: "https://stake.rocketpool.net",
			Network_Prater:  "https://testnet.rocketpool.net",
			Network_Kiln:    "TBD",
		},

		chainID: map[Network]uint{
			Network_Mainnet: 1, // Mainnet
			Network_Prater:  5, // Goerli
			Network_Kiln:    0x1469ca,
		},

		walletPath: "/.rocketpool/data/wallet",

		passwordPath: "/.rocketpool/data/password",

		validatorKeychainPath: "/.rocketpool/data/validators",

		rewardsTreePath: "/.rocketpool/data/rewards-trees",

		storageAddress: map[Network]string{
			Network_Mainnet: "0x1d8f8f00cfa6758d7bE78336684788Fb0ee0Fa46",
			Network_Prater:  "0xd8Cd47263414aFEca62d6e2a3917d6600abDceB3",
			Network_Kiln:    "0x93c769b239c5dBb383683869FaE2667623055420",
		},

		oneInchOracleAddress: map[Network]string{
			Network_Mainnet: "0x07D91f5fb9Bf7798734C3f606dB065549F6893bb",
			Network_Prater:  "0x4eDC966Df24264C9C817295a0753804EcC46Dd22",
			Network_Kiln:    "0xd46a870139F348C3d2596470c355E4BE26b03071",
		},

		rplTokenAddress: map[Network]string{
			Network_Mainnet: "0xb4efd85c19999d84251304bda99e90b92300bd93",
			Network_Prater:  "0xb4efd85c19999d84251304bda99e90b92300bd93",
			Network_Kiln:    "0x50243dc12c1718E85b1A34ddF66F2c70bC13DF09",
		},

		rplFaucetAddress: map[Network]string{
			Network_Mainnet: "",
			Network_Prater:  "0x95D6b8E2106E3B30a72fC87e2B56ce15E37853F9",
			Network_Kiln:    "0xC066e113cD3a568EdcF18D2Fd502f399E63Bc7B7",
		},

		rethAddress: map[Network]string{
			Network_Mainnet: "0xae78736Cd615f374D3085123A210448E74Fc6393",
			Network_Prater:  "0x178E141a0E3b34152f73Ff610437A7bf9B83267A",
			Network_Kiln:    "0xfD64e5461e790b2108Fcf1Bcf9fA6151E7753Ff7",
		},
	}

}

// Get the parameters for this config
func (config *SmartnodeConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.Network,
		&config.ProjectName,
		&config.DataPath,
		&config.ManualMaxFee,
		&config.PriorityFee,
		&config.RplClaimGasThreshold,
		&config.MinipoolStakeGasThreshold,
	}
}

// Getters for the non-editable parameters

func (config *SmartnodeConfig) GetTxWatchUrl() string {
	return config.txWatchUrl[config.Network.Value.(Network)]
}

func (config *SmartnodeConfig) GetStakeUrl() string {
	return config.stakeUrl[config.Network.Value.(Network)]
}

func (config *SmartnodeConfig) GetChainID() uint {
	return config.chainID[config.Network.Value.(Network)]
}

func (config *SmartnodeConfig) GetWalletPath() string {
	if config.parent.IsNativeMode {
		return filepath.Join(config.DataPath.Value.(string), "wallet")
	} else {
		return config.walletPath
	}
}

func (config *SmartnodeConfig) GetPasswordPath() string {
	if config.parent.IsNativeMode {
		return filepath.Join(config.DataPath.Value.(string), "password")
	} else {
		return config.passwordPath
	}
}

func (config *SmartnodeConfig) GetValidatorKeychainPath() string {
	if config.parent.IsNativeMode {
		return filepath.Join(config.DataPath.Value.(string), "validators")
	} else {
		return config.validatorKeychainPath
	}
}

func (config *SmartnodeConfig) GetStorageAddress() string {
	return config.storageAddress[config.Network.Value.(Network)]
}

func (config *SmartnodeConfig) GetOneInchOracleAddress() string {
	return config.oneInchOracleAddress[config.Network.Value.(Network)]
}

func (config *SmartnodeConfig) GetRplTokenAddress() string {
	return config.rplTokenAddress[config.Network.Value.(Network)]
}

func (config *SmartnodeConfig) GetRplFaucetAddress() string {
	return config.rplFaucetAddress[config.Network.Value.(Network)]
}

func (config *SmartnodeConfig) GetSmartnodeContainerTag() string {
	return smartnodeTag
}

func (config *SmartnodeConfig) GetPowProxyContainerTag() string {
	return powProxyTag
}

func (config *SmartnodeConfig) GetPruneProvisionerContainerTag() string {
	return pruneProvisionerTag
}

// The the title for the config
func (config *SmartnodeConfig) GetConfigTitle() string {
	return config.Title
}

func (config *SmartnodeConfig) GetRethAddress() string {
	return config.rethAddress[config.Network.Value.(Network)]
}

func getDefaultDataDir(config *RocketPoolConfig) string {
	return filepath.Join(config.RocketPoolDirectory, "data")
}

func (config *SmartnodeConfig) GetRewardsTreePath(interval uint64) string {
	return filepath.Join(config.rewardsTreePath, fmt.Sprintf(RewardsTreeFilenameFormat, interval))
}
