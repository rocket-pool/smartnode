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

	// Custom URL to download a rewards tree
	RewardsTreeCustomUrl config.Parameter `yaml:"rewardsTreeCustomUrl,omitempty"`

	// URL for an EC with archive mode, for manual rewards tree generation
	ArchiveECUrl config.Parameter `yaml:"archiveEcUrl,omitempty"`

	// Manual override for the watchtower's max fee
	WatchtowerMaxFeeOverride config.Parameter `yaml:"watchtowerMaxFeeOverride,omitempty"`

	// Manual override for the watchtower's priority fee
	WatchtowerPrioFeeOverride config.Parameter `yaml:"watchtowerPrioFeeOverride,omitempty"`

	// The toggle for rolling records
	UseRollingRecords config.Parameter `yaml:"useRollingRecords,omitempty"`

	// The rolling record checkpoint interval
	RecordCheckpointInterval config.Parameter `yaml:"recordCheckpointInterval,omitempty"`

	// The checkpoint retention limit
	CheckpointRetentionLimit config.Parameter `yaml:"checkpointRetentionLimit,omitempty"`

	// The path of the records folder where snapshots of rolling record info is stored during a rewards interval
	RecordsPath config.Parameter `yaml:"recordsPath,omitempty"`
}

// Generates a new Smartnode configuration
func NewSmartnodeConfig(cfg *RocketPoolConfig) *SmartnodeConfig {

	return &SmartnodeConfig{
		Title:  "Smartnode Settings",
		parent: cfg,

		ProjectName: config.Parameter{
			ID:                 ProjectNameID,
			Name:               "Project Name",
			Description:        "This is the prefix that will be attached to all of the Docker containers managed by the Smartnode.",
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
			Description:        "The Ethereum network you want to use - select Holesky Testnet to practice with fake ETH, or Mainnet to stake on the real network using real ETH.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.Network_Mainnet},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth1, config.ContainerID_Eth2, config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            getNetworkOptions(cfg.networks.GetNetworks()),
		},

		ManualMaxFee: config.Parameter{
			ID:                 "manualMaxFee",
			Name:               "Manual Max Fee",
			Description:        "Set this if you want all of the Smartnode's transactions to use this specific max fee value (in gwei), which is the most you'd be willing to pay (*including the priority fee*).\n\nA value of 0 will show you the current suggested max fee based on the current network conditions and let you specify it each time you do a transaction.\n\nAny other value will ignore the recommended max fee and explicitly use this value instead.\n\nThis applies to automated transactions (such as claiming RPL and staking minipools) as well.",
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
			Description: "Occasionally, the Smartnode will attempt to perform some automatic transactions (such as the second `stake` transaction to finish launching a minipool or the `reduce bond` transaction to convert a 16-ETH minipool to an 8-ETH one). During these, your node will use the `Rapid` suggestion from the gas estimator as its max fee.\n\nThis threshold is a limit (in gwei) you can put on that suggestion; your node will not `stake` the new minipool until the suggestion is below this limit.\n\n" +
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
			Description:        "The Smartnode will regularly check the balance of each of your minipools on the Execution Layer (**not** the Beacon Chain).\nIf any of them have a balance greater than this threshold (in ETH), the Smartnode will automatically distribute the balance. This will send your share of the balance to your withdrawal address.\n\nMust be less than 8 ETH.\n\nSet this to 0 to disable automatic distributes.\n[orange]WARNING: if you disable automatic distribution, you **must** ensure you distribute your minipool's balance before it reaches 8 ETH or you will no longer be able to distribute your rewards until you exit the minipool!",
			Type:               config.ParameterType_Float,
			Default:            map[config.Network]interface{}{config.Network_All: float64(1)},
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

		RewardsTreeCustomUrl: config.Parameter{
			ID:                 "rewardsTreeCustomUrl",
			Name:               "Rewards Tree Custom Download URLs",
			Description:        "The Smartnode will automatically download missing rewards tree files from trusted sources like IPFS and Rocket Pool's repository on GitHub. Use this field if you would like to manually specify additional sources that host the rewards tree files, so the Smartnode can download from them as well.\nMultiple URLs can be provided using ';' as separator).\n\nUse '%s' to specify the location of the rewards file name in the URL - for example: `https://my-cool-domain.com/rewards-trees/mainnet/%s`.",
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

		UseRollingRecords: config.Parameter{
			ID:                 "useRollingRecords",
			Name:               "Use Rolling Records",
			Description:        "[orange]**WARNING: EXPERIMENTAL**\n\n[white]Enable this to use the new rolling records feature, which stores attestation records for the entire Rocket Pool network in real time instead of collecting them all after a rewards period during tree generation.\n\nOnly useful for the Oracle DAO, or if you generate your own rewards trees.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		RecordCheckpointInterval: config.Parameter{
			ID:                 "recordCheckpointInterval",
			Name:               "Record Checkpoint Interval",
			Description:        "The number of epochs that should pass before saving a new rolling record checkpoint. Used if Rolling Records is enabled.\n\nOnly useful for the Oracle DAO, or if you generate your own rewards trees.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: uint64(45)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		CheckpointRetentionLimit: config.Parameter{
			ID:                 "checkpointRetentionLimit",
			Name:               "Checkpoint Retention Limit",
			Description:        "The number of checkpoint files to save on-disk before pruning old ones. Used if Rolling Records is enabled.\n\nOnly useful for the Oracle DAO, or if you generate your own rewards trees.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: uint64(200)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		RecordsPath: config.Parameter{
			ID:                 "recordsPath",
			Name:               "Records Path",
			Description:        "The path of the folder to store rolling record checkpoints in during a rewards interval. Used if Rolling Records is enabled.\n\nOnly useful if you're an Oracle DAO member, or if you generate your own rewards trees.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: getDefaultRecordsDir(cfg)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
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
		&cfg.RewardsTreeCustomUrl,
		&cfg.ArchiveECUrl,
		&cfg.WatchtowerMaxFeeOverride,
		&cfg.WatchtowerPrioFeeOverride,
		&cfg.UseRollingRecords,
		&cfg.RecordCheckpointInterval,
		&cfg.CheckpointRetentionLimit,
		&cfg.RecordsPath,
	}
}

// Getters for the non-editable parameters

// Gets the currently chosen network's configuration
func (cfg *SmartnodeConfig) GetNetworkInfo() *config.NetworkInfo {
	for _, network := range cfg.parent.networks.GetNetworks() {
		networkName := fmt.Sprintf("%s", cfg.Network.Value)
		if network.Name == networkName {
			return network
		}
	}
	fmt.Printf("Network info not found for network %s\n", cfg.Network.Value.(string))
	return nil
}

func (cfg *SmartnodeConfig) GetTxWatchUrl() string {
	return cfg.GetNetworkInfo().TxWatchUrl
}

func (cfg *SmartnodeConfig) GetStakeUrl() string {
	return cfg.GetNetworkInfo().StakeUrl
}

func (cfg *SmartnodeConfig) GetChainID() uint {
	return uint(cfg.GetNetworkInfo().ChainID)
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

func (cfg *SmartnodeConfig) GetRecordsPath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), "records")
	}

	return filepath.Join(DaemonDataPath, "records")
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
	return cfg.GetNetworkInfo().StorageAddress
}

func (cfg *SmartnodeConfig) GetRplTokenAddress() string {
	return cfg.GetNetworkInfo().RplTokenAddress
}

func (cfg *SmartnodeConfig) GetRplFaucetAddress() string {
	return cfg.GetNetworkInfo().RplFaucetAddress
}

func (cfg *SmartnodeConfig) GetSnapshotDelegationAddress() string {
	return cfg.GetNetworkInfo().SnapshotDelegationAddress
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
	return cfg.GetNetworkInfo().SnapshotApiDomain
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
	return common.HexToAddress(cfg.GetNetworkInfo().RethAddress)
}

func getDefaultDataDir(config *RocketPoolConfig) string {
	return filepath.Join(config.RocketPoolDirectory, "data")
}

func getDefaultRecordsDir(config *RocketPoolConfig) string {
	return filepath.Join(getDefaultDataDir(config), "records")
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
	return common.HexToAddress(cfg.GetNetworkInfo().V1_0_0_RewardsPoolAddress)
}

func (cfg *SmartnodeConfig) GetV100ClaimNodeAddress() common.Address {
	return common.HexToAddress(cfg.GetNetworkInfo().V1_0_0_ClaimNodeAddress)
}

func (cfg *SmartnodeConfig) GetV100ClaimTrustedNodeAddress() common.Address {
	return common.HexToAddress(cfg.GetNetworkInfo().V1_0_0_ClaimTrustedNodeAddress)
}

func (cfg *SmartnodeConfig) GetV100MinipoolManagerAddress() common.Address {
	return common.HexToAddress(cfg.GetNetworkInfo().V1_0_0_MinipoolManagerAddress)
}

func (cfg *SmartnodeConfig) GetV110NetworkPricesAddress() common.Address {
	return common.HexToAddress(cfg.GetNetworkInfo().V1_1_0_NetworkPricesAddress)
}

func (cfg *SmartnodeConfig) GetV110NodeStakingAddress() common.Address {
	return common.HexToAddress(cfg.GetNetworkInfo().V1_1_0_NodeStakingAddress)
}

func (cfg *SmartnodeConfig) GetV110NodeDepositAddress() common.Address {
	return common.HexToAddress(cfg.GetNetworkInfo().V1_1_0_NodeDepositAddress)
}

func (cfg *SmartnodeConfig) GetV110MinipoolQueueAddress() common.Address {
	return common.HexToAddress(cfg.GetNetworkInfo().V1_1_0_MinipoolQueueAddress)
}

func (cfg *SmartnodeConfig) GetV110MinipoolFactoryAddress() common.Address {
	return common.HexToAddress(cfg.GetNetworkInfo().V1_1_0_MinipoolFactoryAddress)
}

func (cfg *SmartnodeConfig) GetPreviousRewardsPoolAddresses() []common.Address {
	addresses := []common.Address{}
	for _, address := range cfg.GetNetworkInfo().PreviousRewardsPoolAddresses {
		addresses = append(addresses, common.HexToAddress(address))
	}
	return addresses
}

func (cfg *SmartnodeConfig) GetOptimismMessengerAddress() string {
	return cfg.GetNetworkInfo().OptimismPriceMessengerAddress
}

func (cfg *SmartnodeConfig) GetPolygonMessengerAddress() string {
	return cfg.GetNetworkInfo().PolygonPriceMessengerAddress
}

func (cfg *SmartnodeConfig) GetArbitrumMessengerAddress() string {
	return cfg.GetNetworkInfo().ArbitrumPriceMessengerAddress
}

func (cfg *SmartnodeConfig) GetArbitrumMessengerAddressV2() string {
	return cfg.GetNetworkInfo().ArbitrumPriceMessengerAddressV2
}

func (cfg *SmartnodeConfig) GetZkSyncEraMessengerAddress() string {
	return cfg.GetNetworkInfo().ZkSyncEraPriceMessengerAddress
}

func (cfg *SmartnodeConfig) GetBaseMessengerAddress() string {
	return cfg.GetNetworkInfo().BasePriceMessengerAddress
}

func (cfg *SmartnodeConfig) GetScrollMessengerAddress() string {
	return cfg.GetNetworkInfo().ScrollPriceMessengerAddress
}

func (cfg *SmartnodeConfig) GetScrollFeeEstimatorAddress() string {
	return cfg.GetNetworkInfo().ScrollFeeEstimatorAddress
}

func (cfg *SmartnodeConfig) GetRplTwapPoolAddress() string {
	return cfg.GetNetworkInfo().RplTwapPoolAddress
}

func (cfg *SmartnodeConfig) GetMulticallAddress() string {
	return cfg.GetNetworkInfo().MulticallAddress
}

func (cfg *SmartnodeConfig) GetBalanceBatcherAddress() string {
	return cfg.GetNetworkInfo().BalancebatcherAddress
}

func (cfg *SmartnodeConfig) GetFlashbotsProtectUrl() string {
	return cfg.GetNetworkInfo().FlashbotsProtectUrl
}

func getNetworkOptions(networks []*config.NetworkInfo) []config.ParameterOption {
	var options []config.ParameterOption
	for _, network := range networks {
		options = append(options, config.ParameterOption{
			Name:        network.Label,
			Description: network.Description,
			Value:       config.Network(network.Name),
		})
	}
	return options
}
