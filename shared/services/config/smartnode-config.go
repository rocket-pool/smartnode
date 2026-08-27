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
	smartnodeTagPrefix string = "rocketpool/smartnode:v"
	NetworkID          string = "network"
	ProjectNameID      string = "projectName"
	SnapshotID         string = "rocketpool-dao.eth"

	rewardsTreeFilenameFormat          string = "rp-rewards-%s-%d%s"
	minipoolPerformanceFilenameFormat  string = "rp-minipool-performance-%s-%d%s"
	performanceFilenameFormat          string = "rp-performance-%s-%d%s"
	rewardsEventFilenameFormat         string = "rp-rewards-event-%s-%d.yaml"
	rewardsEventRemoteFilenameFormat   string = "rp-rewards-event-%d.yaml"
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
	WatchtowerMaxFeeDefault  uint64 = 30
	WatchtowerPrioFeeDefault uint64 = 1
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
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth1, config.ContainerID_Eth2, config.ContainerID_Validator, config.ContainerID_Grafana, config.ContainerID_Prometheus, config.ContainerID_Exporter},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		DataPath: config.Parameter{
			ID:                 "dataPath",
			Name:               "Data Path",
			Description:        "The absolute path of the `data` folder that contains your node wallet's encrypted file, the password for your node wallet, and all of the validator keys for your minipools. You may use environment variables in this string.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: getDefaultDataDir(cfg)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Validator},
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
			Default:            map[config.Network]interface{}{config.Network_All: defaultNetwork(cfg)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth1, config.ContainerID_Eth2, config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            getNetworkOptions(cfg),
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
			Description:        "Check this box to opt into the responsibility for verifying Protocol DAO proposals once the Houston upgrade has been activated. Your node will regularly check for new proposals, verify their correctness, and submit challenges to any that do not match the on-chain data (e.g., if someone tampered with voting power and attempted to cheat).\n\nTo learn more about the PDAO proposal checking duty, including requirements and RPL bonding, please see the documentation at https://docs.rocketpool.net/pdao#challenge-process.",
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

func (cfg *SmartnodeConfig) networkInfo() *config.NetworkInfo {
	if cfg.parent == nil || cfg.parent.networks == nil {
		return &config.NetworkInfo{}
	}
	info := cfg.parent.networks.GetNetwork(cfg.Network.Value.(config.Network))
	if info == nil {
		return &config.NetworkInfo{}
	}
	return info
}

func (cfg *SmartnodeConfig) GetTxWatchUrl() string {
	return cfg.networkInfo().TxWatchUrl
}

func (cfg *SmartnodeConfig) GetNodeManagerUrl() string {
	return cfg.networkInfo().NodeManagerUrl
}

func (cfg *SmartnodeConfig) GetChainID() uint {
	return cfg.networkInfo().ChainID
}

func (cfg *SmartnodeConfig) GetBeaconNetwork() string {
	return cfg.networkInfo().BeaconNetwork
}

func (cfg *SmartnodeConfig) GetCustomChainConfigDir() string {
	return cfg.networkInfo().CustomChainConfigDir
}

func (cfg *SmartnodeConfig) GetBeaconExplorerUrl() string {
	return cfg.networkInfo().BeaconExplorerUrl
}

func (cfg *SmartnodeConfig) GetCommitBoostChainName() string {
	return cfg.networkInfo().CommitBoostChainName
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

func (cfg *SmartnodeConfig) GetWatchtowerStatePath() string {
	if cfg.parent.IsNativeMode {
		return filepath.Join(cfg.DataPath.Value.(string), WatchtowerFolder, "state.yml")
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
	return cfg.networkInfo().Addresses.Storage
}

func (cfg *SmartnodeConfig) GetRocketSignerRegistryAddress() string {
	return cfg.networkInfo().Addresses.RocketSignerRegistry
}

func (cfg *SmartnodeConfig) GetRplTokenAddress() string {
	return cfg.networkInfo().Addresses.RplToken
}

func (cfg *SmartnodeConfig) GetSmartnodeContainerTag() string {
	return smartnodeTagPrefix + shared.RocketPoolVersion()
}

func (cfg *SmartnodeConfig) GetSnapshotApiDomain() string {
	return cfg.networkInfo().SnapshotApiDomain
}

func (cfg *SmartnodeConfig) GetVotingSnapshotID() [32]byte {
	// So the contract wants a Keccak'd hash of the voting ID, but Snapshot's service wants ASCII so it can display the ID in plain text; we have to do this to make it play nicely with Snapshot
	buffer := [32]byte{}
	idBytes := []byte(SnapshotID)
	copy(buffer[0:], idBytes)
	return buffer
}

func (cfg *SmartnodeConfig) GetSnapshotID() string {
	return SnapshotID
}

// The title for the config
func (cfg *SmartnodeConfig) GetConfigTitle() string {
	return cfg.Title
}

func (cfg *SmartnodeConfig) GetRethAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.Reth)
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

func (cfg *SmartnodeConfig) GetRewardsEventFilename(interval uint64) string {
	return fmt.Sprintf(rewardsEventFilenameFormat, string(cfg.Network.Value.(config.Network)), interval)
}

func (cfg *SmartnodeConfig) GetRewardsEventRemoteFilename(interval uint64) string {
	return fmt.Sprintf(rewardsEventRemoteFilenameFormat, interval)
}

func (cfg *SmartnodeConfig) GetRewardsEventPath(interval uint64, daemon bool) string {
	return filepath.Join(
		cfg.GetRewardsTreeDirectory(daemon),
		cfg.GetRewardsEventFilename(interval),
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
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_0_0_RewardsPool)
}

func (cfg *SmartnodeConfig) GetV100ClaimNodeAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_0_0_ClaimNode)
}

func (cfg *SmartnodeConfig) GetV100ClaimTrustedNodeAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_0_0_ClaimTrustedNode)
}

func (cfg *SmartnodeConfig) GetV100MinipoolManagerAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_0_0_MinipoolManager)
}

func (cfg *SmartnodeConfig) GetV110NetworkPricesAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_1_0_NetworkPrices)
}

func (cfg *SmartnodeConfig) GetV120NetworkPricesAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_2_0_NetworkPrices)
}

func (cfg *SmartnodeConfig) GetV120NetworkBalancesAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_2_0_NetworkBalances)
}

func (cfg *SmartnodeConfig) GetV110NodeStakingAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_1_0_NodeStaking)
}

func (cfg *SmartnodeConfig) GetV110NodeDepositAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_1_0_NodeDeposit)
}

func (cfg *SmartnodeConfig) GetV110MinipoolQueueAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_1_0_MinipoolQueue)
}

func (cfg *SmartnodeConfig) GetV110MinipoolFactoryAddress() common.Address {
	return common.HexToAddress(cfg.networkInfo().Addresses.V1_1_0_MinipoolFactory)
}

func (cfg *SmartnodeConfig) GetPreviousRewardsPoolAddresses() []common.Address {
	return hexAddresses(cfg.networkInfo().Addresses.PreviousRewardsPools)
}

func (cfg *SmartnodeConfig) GetPreviousRocketDAOProtocolVerifierAddresses() []common.Address {
	return hexAddresses(cfg.networkInfo().Addresses.PreviousDAOVerifiers)
}

func (cfg *SmartnodeConfig) GetOptimismMessengerAddress() string {
	return cfg.networkInfo().Addresses.OptimismPriceMessenger
}

func (cfg *SmartnodeConfig) GetPolygonMessengerAddress() string {
	return cfg.networkInfo().Addresses.PolygonPriceMessenger
}

func (cfg *SmartnodeConfig) GetArbitrumMessengerAddress() string {
	return cfg.networkInfo().Addresses.ArbitrumPriceMessenger
}

func (cfg *SmartnodeConfig) GetArbitrumMessengerAddressV2() string {
	return cfg.networkInfo().Addresses.ArbitrumPriceMessengerV2
}

func (cfg *SmartnodeConfig) GetZkSyncEraMessengerAddress() string {
	return cfg.networkInfo().Addresses.ZkSyncEraPriceMessenger
}

func (cfg *SmartnodeConfig) GetBaseMessengerAddress() string {
	return cfg.networkInfo().Addresses.BasePriceMessenger
}

func (cfg *SmartnodeConfig) GetScrollMessengerAddress() string {
	return cfg.networkInfo().Addresses.ScrollPriceMessenger
}

func (cfg *SmartnodeConfig) GetScrollFeeEstimatorAddress() string {
	return cfg.networkInfo().Addresses.ScrollFeeEstimator
}

func (cfg *SmartnodeConfig) GetRplTwapPoolAddress() string {
	return cfg.networkInfo().Addresses.RplTwapPool
}

func (cfg *SmartnodeConfig) GetMulticallAddress() string {
	return cfg.networkInfo().Addresses.Multicall
}

func (cfg *SmartnodeConfig) GetBalanceBatcherAddress() string {
	return cfg.networkInfo().Addresses.BalanceBatcher
}

// Utility function to get the state manager contracts
func (cfg *SmartnodeConfig) GetStateManagerContracts() StateManagerContracts {
	return StateManagerContracts{
		Multicaller:    common.HexToAddress(cfg.GetMulticallAddress()),
		BalanceBatcher: common.HexToAddress(cfg.GetBalanceBatcherAddress()),
	}
}

func (cfg *SmartnodeConfig) GetFlashbotsProtectUrl() string {
	return cfg.networkInfo().FlashbotsProtectUrl
}

func (cfg *SmartnodeConfig) GetFlashbotsRelayUrl() string {
	return cfg.networkInfo().FlashbotsRelayUrl
}

func defaultNetwork(cfg *RocketPoolConfig) config.Network {
	if cfg == nil || cfg.networks == nil {
		return config.Network_Unknown
	}
	return cfg.networks.DefaultNetwork()
}

func getNetworkOptions(cfg *RocketPoolConfig) []config.ParameterOption {
	if cfg == nil || cfg.networks == nil {
		return nil
	}
	options := make([]config.ParameterOption, 0, len(cfg.networks.AllNetworks()))
	for _, n := range cfg.networks.AllNetworks() {
		options = append(options, config.ParameterOption{
			Name:        n.Label,
			Description: n.Description,
			Value:       n.ID(),
		})
	}
	return options
}
