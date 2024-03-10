package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/nodeset-org/hyperdrive/shared/config/migration"
	"github.com/pbnjay/memory"
	"github.com/rocket-pool/node-manager-core/config"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/config/ids"
	"gopkg.in/yaml.v3"
)

// =========================
// === Smart Node Config ===
// =========================

const (
	// Tags
	smartnodeTag string = "rocketpool/smartnode:v" + shared.RocketPoolVersion
)

// The master configuration struct
type SmartNodeConfig struct {
	// Smart Node settings
	DebugMode                     config.Parameter[bool]
	Network                       config.Parameter[config.Network]
	ClientMode                    config.Parameter[config.ClientMode]
	ProjectName                   config.Parameter[string]
	UserDataPath                  config.Parameter[string]
	WatchtowerStatePath           config.Parameter[string]
	AutoTxMaxFee                  config.Parameter[float64]
	MaxPriorityFee                config.Parameter[float64]
	AutoTxGasThreshold            config.Parameter[float64]
	DistributeThreshold           config.Parameter[float64]
	RewardsTreeMode               config.Parameter[RewardsMode]
	RewardsTreeCustomUrl          config.Parameter[string]
	ArchiveEcUrl                  config.Parameter[string]
	WatchtowerMaxFeeOverride      config.Parameter[float64]
	WatchtowerPriorityFeeOverride config.Parameter[float64]
	UseRollingRecords             config.Parameter[bool]
	RecordCheckpointInterval      config.Parameter[uint64]
	CheckpointRetentionLimit      config.Parameter[uint64]
	RecordsPath                   config.Parameter[string]
	VerifyProposals               config.Parameter[bool]

	// Execution client settings
	LocalExecutionConfig    *config.LocalExecutionConfig
	ExternalExecutionConfig *config.ExternalExecutionConfig

	// Beacon node settings
	LocalBeaconConfig    *config.LocalBeaconConfig
	ExternalBeaconConfig *config.ExternalBeaconConfig

	// Fallback clients
	Fallback *config.FallbackConfig

	// Metrics
	Metrics *MetricsConfig

	// Native mode
	Native *NativeConfig

	// MEV-Boost
	MevBoost *MevBoostConfig

	// Addons
	Addons map[string]any

	// Internal fields
	Version             string
	RocketPoolDirectory string
	IsNativeMode        bool
}

// Load configuration settings from a file
func LoadFromFile(path string) (*SmartNodeConfig, error) {
	// Return nil if the file doesn't exist
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	// Read the file
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read Smart Node settings file at %s: %w", shellescape.Quote(path), err)
	}

	// Attempt to parse it out into a settings map
	var settings map[string]any
	if err := yaml.Unmarshal(configBytes, &settings); err != nil {
		return nil, fmt.Errorf("could not parse settings file: %w", err)
	}

	// Deserialize it into a config object
	cfg := NewSmartNodeConfig(filepath.Dir(path), false)
	err = cfg.Deserialize(settings)
	if err != nil {
		return nil, fmt.Errorf("could not deserialize settings file: %w", err)
	}

	return cfg, nil
}

// Creates a new Smart Node configuration instance
func NewSmartNodeConfig(rpDir string, isNativeMode bool) *SmartNodeConfig {
	cfg := &SmartNodeConfig{
		RocketPoolDirectory: rpDir,
		IsNativeMode:        isNativeMode,

		DebugMode: config.Parameter[bool]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.DebugModeID,
				Name:               "Debug Mode",
				Description:        "Enable debug log printing in the daemon and watchtower.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]bool{
				config.Network_All: false,
			},
		},

		Network: config.Parameter[config.Network]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.NetworkID,
				Name:               "Network",
				Description:        "The Ethereum network you want to use - select Prater Testnet or Holesky Testnet to practice with fake ETH, or Mainnet to stake on the real network using real ETH.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, ContainerID_Watchtower, config.ContainerID_ExecutionClient, config.ContainerID_BeaconNode, config.ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Options: getNetworkOptions(),
			Default: map[config.Network]config.Network{
				config.Network_All: config.Network_Mainnet,
			},
		},

		ClientMode: config.Parameter[config.ClientMode]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.ClientModeID,
				Name:               "Client Mode",
				Description:        "Choose which mode to use for your Execution Client and Beacon Node - locally managed (Docker Mode), or externally managed (Hybrid Mode).",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, config.ContainerID_ExecutionClient, config.ContainerID_BeaconNode},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Options: []*config.ParameterOption[config.ClientMode]{
				{
					ParameterOptionCommon: &config.ParameterOptionCommon{
						Name:        "Locally Managed",
						Description: "Allow the Smart Node to manage the Execution Client and Beacon Node for you (Docker Mode)",
					},
					Value: config.ClientMode_Local,
				}, {
					ParameterOptionCommon: &config.ParameterOptionCommon{
						Name:        "Externally Managed",
						Description: "Use an existing Execution Client and Beacon Node that you manage on your own (Hybrid Mode)",
					},
					Value: config.ClientMode_External,
				}},
			Default: map[config.Network]config.ClientMode{
				config.Network_All: config.ClientMode_Local,
			},
		},

		ProjectName: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.ProjectNameID,
				Name:               "Project Name",
				Description:        "This is the prefix that will be attached to all of the Docker containers managed by the Smart Node.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_BeaconNode, config.ContainerID_Daemon, config.ContainerID_ExecutionClient, config.ContainerID_Exporter, config.ContainerID_Grafana, config.ContainerID_Prometheus, config.ContainerID_ValidatorClient, ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: "rocketpool",
			},
		},

		UserDataPath: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.UserDataPathID,
				Name:               "User Data Path",
				Description:        "The absolute path of your personal `data` folder that contains your node wallet's encrypted file, the password for your node wallet, and all of the validator keys for your minipools. You may use environment variables in this string.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, ContainerID_Watchtower, config.ContainerID_ValidatorClient},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: getDefaultDataDir(rpDir),
			},
		},

		WatchtowerStatePath: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.WatchtowerStatePath,
				Name:               "Watchtower Path",
				Description:        "The absolute path of the watchtower state folder that contains persistent state that is used by the watchtower process on trusted nodes. **Only relevant for trusted nodes.**",
				AffectsContainers:  []config.ContainerID{ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: filepath.Join(rpDir, "watchtower"),
			},
		},

		AutoTxMaxFee: config.Parameter[float64]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.AutoTxMaxFeeID,
				Name:               "Auto TX Max Fee",
				Description:        "Set this if you want all of the Smartnode's transactions to use this specific max fee value (in gwei), which is the most you'd be willing to pay (*including the priority fee*).\n\nA value of 0 will show you the current suggested max fee based on the current network conditions and let you specify it each time you do a transaction.\n\nAny other value will ignore the recommended max fee and explicitly use this value instead.\n\nThis applies to automated transactions (such as claiming RPL and staking minipools) as well.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]float64{
				config.Network_All: float64(0),
			},
		},

		MaxPriorityFee: config.Parameter[float64]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.MaxPriorityFeeID,
				Name:               "Max Priority Fee",
				Description:        "The default value for the priority fee (in gwei) for all of your transactions. This describes how much you're willing to pay *above the network's current base fee* - the higher this is, the more ETH you give to the validators for including your transaction, which generally means it will be included in a block faster (as long as your max fee is sufficiently high to cover the current network conditions).\n\nMust be larger than 0.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]float64{
				config.Network_All: float64(1),
			},
		},

		AutoTxGasThreshold: config.Parameter[float64]{
			ParameterCommon: &config.ParameterCommon{
				ID:   ids.AutoTxGasThresholdID,
				Name: "Automatic TX Gas Threshold",
				Description: "Occasionally, the Smartnode will attempt to perform some automatic transactions (such as the second `stake` transaction to finish launching a minipool or the `reduce bond` transaction to convert a 16-ETH minipool to an 8-ETH one). During these, your node will use the `Rapid` suggestion from the gas estimator as its max fee.\n\nThis threshold is a limit (in gwei) you can put on that suggestion; your node will not `stake` the new minipool until the suggestion is below this limit.\n\n" +
					"A value of 0 will disable non-essential automatic transactions (such as minipool balance distribution and bond reduction), but essential transactions (such as minipool staking and solo migration promotion) will not be disabled.\n\n" +
					"NOTE: the node will ignore this limit and automatically execute transactions at whatever the suggested fee happens to be once too much time has passed since those transactions were first eligible. You may end up paying more than you wanted to if you set this too low!",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]float64{
				config.Network_All: float64(100),
			},
		},

		DistributeThreshold: config.Parameter[float64]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.DistributeThresholdID,
				Name:               "Auto-Distribute Threshold",
				Description:        "The Smartnode will regularly check the balance of each of your minipools on the Execution Layer (**not** the Beacon Chain).\nIf any of them have a balance greater than this threshold (in ETH), the Smartnode will automatically distribute the balance. This will send your share of the balance to your withdrawal address.\n\nMust be less than 8 ETH.\n\nSet this to 0 to disable automatic distributes.\n[orange]WARNING: if you disable automatic distribution, you **must** ensure you distribute your minipool's balance before it reaches 8 ETH or you will no longer be able to distribute your rewards until you exit the minipool!",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]float64{
				config.Network_All: 1.0,
			},
		},

		RewardsTreeMode: config.Parameter[RewardsMode]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.RewardsTreeModeID,
				Name:               "Rewards Tree Mode",
				Description:        "Select how you want to acquire the Merkle Tree files for each rewards interval.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon, ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Options: []*config.ParameterOption[RewardsMode]{{
				ParameterOptionCommon: &config.ParameterOptionCommon{
					Name:        "Download",
					Description: "Automatically download the Merkle Tree rewards files that were published by the Oracle DAO after a rewards checkpoint.",
				},
				Value: RewardsMode_Download,
			}, {
				ParameterOptionCommon: &config.ParameterOptionCommon{
					Name:        "Generate",
					Description: "Use your node to automatically generate the Merkle Tree rewards file once a checkpoint has passed. This option lets you build and verify the file that the Oracle DAO created if you prefer not to trust it and want to generate the tree yourself.\n\n[orange]WARNING: Generating the tree can take a *very long time* if many node operators are opted into the Smoothing Pool, which could impact your attestation performance!",
				},
				Value: RewardsMode_Generate,
			}},
			Default: map[config.Network]RewardsMode{
				config.Network_All: RewardsMode_Download,
			},
		},

		RewardsTreeCustomUrl: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.RewardsTreeCustomUrlID,
				Name:               "Rewards Tree Custom Download URLs",
				Description:        "The Smartnode will automatically download missing rewards tree files from trusted sources like IPFS and Rocket Pool's repository on GitHub. Use this field if you would like to manually specify additional sources that host the rewards tree files, so the Smartnode can download from them as well.\nMultiple URLs can be provided using ';' as separator).\n\nUse '%s' to specify the location of the rewards file name in the URL - for example: `https://my-cool-domain.com/rewards-trees/mainnet/%s`.",
				AffectsContainers:  []config.ContainerID{ContainerID_Watchtower},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: "",
			},
		},

		ArchiveEcUrl: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.ArchiveEcUrlID,
				Name:               "Archive-Mode EC URL",
				Description:        "[orange]**For manual Merkle rewards tree generation only.**[white]\n\nGenerating the Merkle rewards tree files for past rewards intervals typically requires an Execution client with Archive mode enabled, which is usually disabled on your primary and fallback Execution clients to save disk space.\nIf you want to generate your own rewards tree files for intervals from a long time ago, you may enter the URL of an Execution client with Archive access here.\n\nFor a free light client with Archive access, you may use https://www.alchemy.com/supernode.",
				AffectsContainers:  []config.ContainerID{ContainerID_Watchtower},
				CanBeBlank:         true,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: "",
			},
		},

		WatchtowerMaxFeeOverride: config.Parameter[float64]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.WatchtowerMaxFeeOverrideID,
				Name:               "Watchtower Max Fee Override",
				Description:        fmt.Sprintf("[orange]**For Oracle DAO members only.**\n\n[white]Use this to override the max fee (in gwei) for watchtower transactions. Note that if you set it below %d, the setting will be ignored; it can only be used to set the max fee higher than %d during times of extreme network stress.", WatchtowerMaxFeeDefault, WatchtowerMaxFeeDefault),
				AffectsContainers:  []config.ContainerID{ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[config.Network]float64{
				config.Network_All: float64(WatchtowerMaxFeeDefault),
			},
		},

		WatchtowerPriorityFeeOverride: config.Parameter[float64]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.WatchtowerPriorityFeeOverrideID,
				Name:               "Watchtower Priority Fee Override",
				Description:        fmt.Sprintf("[orange]**For Oracle DAO members only.**\n\n[white]Use this to override the priority fee (in gwei) for watchtower transactions. Note that if you set it below %d, the setting will be ignored; it can only be used to set the priority fee higher than %d during times of extreme network stress.", WatchtowerPriorityFeeDefault, WatchtowerPriorityFeeDefault),
				AffectsContainers:  []config.ContainerID{ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: true,
			},
			Default: map[config.Network]float64{
				config.Network_All: float64(WatchtowerPriorityFeeDefault),
			},
		},

		UseRollingRecords: config.Parameter[bool]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.UseRollingRecordsID,
				Name:               "Use Rolling Records",
				Description:        "Enable this to use the new rolling records feature, which stores attestation records for the entire Rocket Pool network in real time instead of collecting them all after a rewards period during tree generation.\n\nOnly useful for the Oracle DAO, or if you generate your own rewards trees.",
				AffectsContainers:  []config.ContainerID{ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]bool{
				config.Network_All: false,
			},
		},

		RecordCheckpointInterval: config.Parameter[uint64]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.RecordCheckpointIntervalID,
				Name:               "Record Checkpoint Interval",
				Description:        "The number of epochs that should pass before saving a new rolling record checkpoint. Used if Rolling Records is enabled.\n\nOnly useful for the Oracle DAO, or if you generate your own rewards trees.",
				AffectsContainers:  []config.ContainerID{ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]uint64{
				config.Network_All: 45,
			},
		},

		CheckpointRetentionLimit: config.Parameter[uint64]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.CheckpointRetentionLimitID,
				Name:               "Checkpoint Retention Limit",
				Description:        "The number of checkpoint files to save on-disk before pruning old ones. Used if Rolling Records is enabled.\n\nOnly useful for the Oracle DAO, or if you generate your own rewards trees.",
				AffectsContainers:  []config.ContainerID{ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]uint64{
				config.Network_All: uint64(200),
			},
		},

		RecordsPath: config.Parameter[string]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.RecordsPathID,
				Name:               "Records Path",
				Description:        "The path of the folder to store rolling record checkpoints in during a rewards interval. Used if Rolling Records is enabled.\n\nOnly useful if you're an Oracle DAO member, or if you generate your own rewards trees.",
				AffectsContainers:  []config.ContainerID{ContainerID_Watchtower},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]string{
				config.Network_All: getDefaultRecordsDir(rpDir),
			},
		},

		VerifyProposals: config.Parameter[bool]{
			ParameterCommon: &config.ParameterCommon{
				ID:                 ids.VerifyProposalsID,
				Name:               "Enable PDAO Proposal Checker",
				Description:        "Check this box to opt into the responsibility for verifying Protocol DAO proposals once the Houston upgrade has been activated. Your node will regularly check for new proposals, verify their correctness, and submit challenges to any that do not match the on-chain data (e.g., if someone tampered with voting power and attempted to cheat).\n\nTo learn more about the PDAO proposal checking duty, including requirements and RPL bonding, please see the documentation at <placeholder>.",
				AffectsContainers:  []config.ContainerID{config.ContainerID_Daemon},
				CanBeBlank:         false,
				OverwriteOnUpgrade: false,
			},
			Default: map[config.Network]bool{
				config.Network_All: false,
			},
		},
	}

	// Create the subconfigs
	cfg.LocalExecutionConfig = NewLocalExecutionConfig()
	cfg.ExternalExecutionConfig = config.NewExternalExecutionConfig()
	cfg.LocalBeaconConfig = NewLocalBeaconConfig()
	cfg.ExternalBeaconConfig = config.NewExternalBeaconConfig()
	cfg.Fallback = config.NewFallbackConfig()
	cfg.Metrics = NewMetricsConfig()
	cfg.Native = NewNativeConfig(cfg)
	cfg.MevBoost = NewMevBoostConfig(cfg)

	// Apply the default values for mainnet
	cfg.Network.Value = config.Network_Mainnet
	cfg.applyAllDefaults()

	return cfg
}

// Get the title for this config
func (cfg *SmartNodeConfig) GetTitle() string {
	return "Hyperdrive"
}

// Get the config.Parameters for this config
func (cfg *SmartNodeConfig) GetParameters() []config.IParameter {
	return []config.IParameter{
		&cfg.ProjectName,
		&cfg.UserDataPath,
		&cfg.Network,
		&cfg.ClientMode,
		&cfg.VerifyProposals,
		&cfg.AutoTxMaxFee,
		&cfg.MaxPriorityFee,
		&cfg.AutoTxGasThreshold,
		&cfg.DistributeThreshold,
		&cfg.RewardsTreeMode,
		&cfg.RewardsTreeCustomUrl,
		&cfg.WatchtowerMaxFeeOverride,
		&cfg.WatchtowerPriorityFeeOverride,
		&cfg.ArchiveEcUrl,
		&cfg.UseRollingRecords,
		&cfg.RecordCheckpointInterval,
		&cfg.CheckpointRetentionLimit,
		&cfg.WatchtowerStatePath,
		&cfg.RecordsPath,

		&cfg.DebugMode,
	}
}

// Get the subconfigurations for this config
func (cfg *SmartNodeConfig) GetSubconfigs() map[string]config.IConfigSection {
	return map[string]config.IConfigSection{
		"fallback":          cfg.Fallback,
		"localExecution":    cfg.LocalExecutionConfig,
		"externalExecution": cfg.ExternalExecutionConfig,
		"localBeacon":       cfg.LocalBeaconConfig,
		"externalBeacon":    cfg.ExternalBeaconConfig,
		"metrics":           cfg.Metrics,
		"native":            cfg.Native,
		"mevBoost":          cfg.MevBoost,
	}
}

// Serializes the configuration into a map of maps, compatible with a settings file
func (cfg *SmartNodeConfig) Serialize(modules []IAddonConfig) map[string]any {
	masterMap := map[string]any{}

	hdMap := config.Serialize(cfg)
	masterMap[ids.UserDirectoryKey] = cfg.HyperdriveUserDirectory
	masterMap[ids.VersionID] = fmt.Sprintf("v%s", shared.HyperdriveVersion)
	masterMap[ids.RootConfigID] = hdMap

	// Handle modules
	modulesMap := map[string]any{}
	for modName, value := range cfg.Addons {
		// Copy the module configs already on-board
		modulesMap[modName] = value
	}
	for _, module := range modules {
		// Serialize / overwrite them with explictly provided ones
		modMap := config.Serialize(module)
		modulesMap[module.GetModuleName()] = modMap
	}
	masterMap[ModulesName] = modulesMap
	return masterMap
}

// Deserializes a settings file into this config
func (cfg *SmartNodeConfig) Deserialize(masterMap map[string]any) error {
	// Upgrade the config to the latest version
	err := migration.UpdateConfig(masterMap)
	if err != nil {
		return fmt.Errorf("error upgrading configuration to v%s: %w", shared.RocketPoolVersion, err)
	}

	// Get the network
	network := config.Network_Mainnet
	hyperdriveParams, exists := masterMap[ids.RootConfigID]
	if !exists {
		return fmt.Errorf("config is missing the [%s] section", ids.RootConfigID)
	}
	hdMap, isMap := hyperdriveParams.(map[string]any)
	if !isMap {
		return fmt.Errorf("config has an entry named [%s] but it is not a map, it's a %s", ids.RootConfigID, reflect.TypeOf(hyperdriveParams))
	}
	networkVal, exists := hdMap[cfg.Network.ID]
	if exists {
		networkString, isString := networkVal.(string)
		if !isString {
			return fmt.Errorf("expected [%s - %s] to be a string but it is not", ids.RootConfigID, cfg.Network.ID)
		}
		network = config.Network(networkString)
	}

	// Deserialize the params and subconfigs
	err = config.Deserialize(cfg, hdMap, network)
	if err != nil {
		return fmt.Errorf("error deserializing [%s]: %w", ids.RootConfigID, err)
	}

	// Get the special fields
	udKey, exists := masterMap[ids.UserDirectoryKey]
	if !exists {
		return fmt.Errorf("expected a user directory config.Parameter named [%s] but it was not found", ids.UserDirectoryKey)
	}
	cfg.HyperdriveUserDirectory = udKey.(string)
	version, exists := masterMap[ids.VersionID]
	if !exists {
		return fmt.Errorf("expected a version config.Parameter named [%s] but it was not found", ids.VersionID)
	}
	cfg.Version = version.(string)

	// Handle modules
	modules, exists := masterMap[ModulesName]
	if exists {
		if modMap, ok := modules.(map[string]any); ok {
			cfg.Modules = modMap
		} else {
			return fmt.Errorf("config has an entry named [%s] but it is not a map, it's a %s", ModulesName, reflect.TypeOf(modules))
		}
	} else {
		cfg.Modules = map[string]any{}
	}

	return nil
}

// =====================
// === Field Helpers ===
// =====================

// Applies all of the defaults to all of the settings that have them defined
func (cfg *SmartNodeConfig) applyAllDefaults() {
	network := cfg.Network.Value
	config.ApplyDefaults(cfg, network)
}

// Get the list of options for networks to run on
func getNetworkOptions() []*config.ParameterOption[config.Network] {
	options := []*config.ParameterOption[config.Network]{
		{
			ParameterOptionCommon: &config.ParameterOptionCommon{
				Name:        "Ethereum Mainnet",
				Description: "This is the real Ethereum main network, using real ETH and real RPL to make real validators.",
			},
			Value: config.Network_Mainnet,
		}, {
			ParameterOptionCommon: &config.ParameterOptionCommon{
				Name:        "Holesky Testnet",
				Description: "This is the Holešky (Holešovice) test network, which is the next generation of long-lived testnets for Ethereum. It uses free fake ETH and free fake RPL to make fake validators.\nUse this if you want to practice running the Smartnode in a free, safe environment before moving to Mainnet.",
			},
			Value: config.Network_Holesky,
		},
	}

	if strings.HasSuffix(shared.RocketPoolVersion, "-dev") {
		options = append(options, &config.ParameterOption[config.Network]{
			ParameterOptionCommon: &config.ParameterOptionCommon{
				Name:        "Devnet",
				Description: "This is a development network used by Rocket Pool engineers to test new features and contract upgrades before they are promoted to Holesky for staging. You should not use this network unless invited to do so by the developers.",
			},
			Value: Network_Devnet,
		})
	}

	return options
}

// Get a more verbose client description, including warnings
func getAugmentedEcDescription(client config.ExecutionClient, originalDescription string) string {
	switch client {
	case config.ExecutionClient_Nethermind:
		totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
		if totalMemoryGB < 9 {
			return fmt.Sprintf("%s\n\n[red]WARNING: Nethermind currently requires over 8 GB of RAM to run smoothly. We do not recommend it for your system. This may be improved in a future release.", originalDescription)
		}
	case config.ExecutionClient_Besu:
		totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
		if totalMemoryGB < 9 {
			return fmt.Sprintf("%s\n\n[red]WARNING: Besu currently requires over 8 GB of RAM to run smoothly. We do not recommend it for your system. This may be improved in a future release.", originalDescription)
		}
	}

	return originalDescription
}

// Get the default data directory
func getDefaultDataDir(rpDir string) string {
	return filepath.Join(rpDir, "data")
}

// Get the default Watchtower records directory
func getDefaultRecordsDir(rpDir string) string {
	return filepath.Join(getDefaultDataDir(rpDir), "records")
}

// ==============================
// === IConfig Implementation ===
// ==============================

func (cfg *SmartNodeConfig) GetNodeAddressFilePath() string {
	return filepath.Join(cfg.UserDataPath.Value, UserAddressFilename)
}

func (cfg *SmartNodeConfig) GetWalletFilePath() string {
	return filepath.Join(cfg.UserDataPath.Value, UserWalletDataFilename)
}

func (cfg *SmartNodeConfig) GetPasswordFilePath() string {
	return filepath.Join(cfg.UserDataPath.Value, UserPasswordFilename)
}

func (cfg *SmartNodeConfig) GetNetworkResources() *config.NetworkResources {
	switch cfg.Network.Value {
	case Network_Devnet:
		return config.NewResources(config.Network_Holesky)
	default:
		return config.NewResources(cfg.Network.Value)
	}
}

func (cfg *SmartNodeConfig) GetExecutionClientUrls() (string, string) {
	primaryEcUrl := cfg.GetEcHttpEndpoint()
	var fallbackEcUrl string
	if cfg.Fallback.UseFallbackClients.Value {
		fallbackEcUrl = cfg.Fallback.EcHttpUrl.Value
	}
	return primaryEcUrl, fallbackEcUrl
}

func (cfg *SmartNodeConfig) GetBeaconNodeUrls() (string, string) {
	primaryBnUrl := cfg.GetBnHttpEndpoint()
	var fallbackBnUrl string
	if cfg.Fallback.UseFallbackClients.Value {
		fallbackBnUrl = cfg.Fallback.BnHttpUrl.Value
	}
	return primaryBnUrl, fallbackBnUrl
}