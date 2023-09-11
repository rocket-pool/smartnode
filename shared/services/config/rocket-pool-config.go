package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/pbnjay/memory"
	"github.com/rocket-pool/smartnode/addons"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config/migration"
	addontypes "github.com/rocket-pool/smartnode/shared/types/addons"
	"github.com/rocket-pool/smartnode/shared/types/config"
	"gopkg.in/yaml.v2"
)

// Constants
const (
	rootConfigName string = "root"

	ApiContainerName          string = "api"
	Eth1ContainerName         string = "eth1"
	Eth1FallbackContainerName string = "eth1-fallback"
	Eth2ContainerName         string = "eth2"
	ExporterContainerName     string = "exporter"
	GrafanaContainerName      string = "grafana"
	MevBoostContainerName     string = "mev-boost"
	NodeContainerName         string = "node"
	PrometheusContainerName   string = "prometheus"
	ValidatorContainerName    string = "validator"
	WatchtowerContainerName   string = "watchtower"

	FeeRecipientFileEnvVar string = "FEE_RECIPIENT_FILE"
	FeeRecipientEnvVar     string = "FEE_RECIPIENT"
)

// Defaults
const defaultBnMetricsPort uint16 = 9100
const defaultVcMetricsPort uint16 = 9101
const defaultNodeMetricsPort uint16 = 9102
const defaultExporterMetricsPort uint16 = 9103
const defaultWatchtowerMetricsPort uint16 = 9104
const defaultEcMetricsPort uint16 = 9105

// The master configuration struct
type RocketPoolConfig struct {
	Title string `yaml:"-"`

	Version string `yaml:"-"`

	RocketPoolDirectory string `yaml:"-"`

	IsNativeMode bool `yaml:"-"`

	// Execution client settings
	ExecutionClientMode config.Parameter `yaml:"executionClientMode,omitempty"`
	ExecutionClient     config.Parameter `yaml:"executionClient,omitempty"`

	// Fallback settings
	UseFallbackClients config.Parameter `yaml:"useFallbackClients,omitempty"`
	ReconnectDelay     config.Parameter `yaml:"reconnectDelay,omitempty"`

	// Consensus client settings
	ConsensusClientMode     config.Parameter `yaml:"consensusClientMode,omitempty"`
	ConsensusClient         config.Parameter `yaml:"consensusClient,omitempty"`
	ExternalConsensusClient config.Parameter `yaml:"externalConsensusClient,omitempty"`

	// Metrics settings
	EnableMetrics           config.Parameter `yaml:"enableMetrics,omitempty"`
	EnableODaoMetrics       config.Parameter `yaml:"enableODaoMetrics,omitempty"`
	EcMetricsPort           config.Parameter `yaml:"ecMetricsPort,omitempty"`
	BnMetricsPort           config.Parameter `yaml:"bnMetricsPort,omitempty"`
	VcMetricsPort           config.Parameter `yaml:"vcMetricsPort,omitempty"`
	NodeMetricsPort         config.Parameter `yaml:"nodeMetricsPort,omitempty"`
	ExporterMetricsPort     config.Parameter `yaml:"exporterMetricsPort,omitempty"`
	WatchtowerMetricsPort   config.Parameter `yaml:"watchtowerMetricsPort,omitempty"`
	EnableBitflyNodeMetrics config.Parameter `yaml:"enableBitflyNodeMetrics,omitempty"`

	// The Smartnode configuration
	Smartnode *SmartnodeConfig `yaml:"smartnode,omitempty"`

	// Execution client configurations
	ExecutionCommon   *ExecutionCommonConfig   `yaml:"executionCommon,omitempty"`
	Geth              *GethConfig              `yaml:"geth,omitempty"`
	Nethermind        *NethermindConfig        `yaml:"nethermind,omitempty"`
	Besu              *BesuConfig              `yaml:"besu,omitempty"`
	ExternalExecution *ExternalExecutionConfig `yaml:"externalExecution,omitempty"`

	// Consensus client configurations
	ConsensusCommon    *ConsensusCommonConfig    `yaml:"consensusCommon,omitempty"`
	Lighthouse         *LighthouseConfig         `yaml:"lighthouse,omitempty"`
	Lodestar           *LodestarConfig           `yaml:"lodestar,omitempty"`
	Nimbus             *NimbusConfig             `yaml:"nimbus,omitempty"`
	Prysm              *PrysmConfig              `yaml:"prysm,omitempty"`
	Teku               *TekuConfig               `yaml:"teku,omitempty"`
	ExternalLighthouse *ExternalLighthouseConfig `yaml:"externalLighthouse,omitempty"`
	ExternalNimbus     *ExternalNimbusConfig     `yaml:"externalNimbus,omitempty"`
	ExternalLodestar   *ExternalLodestarConfig   `yaml:"externalLodestar,omitempty"`
	ExternalPrysm      *ExternalPrysmConfig      `yaml:"externalPrysm,omitempty"`
	ExternalTeku       *ExternalTekuConfig       `yaml:"externalTeku,omitempty"`

	// Fallback client configurations
	FallbackNormal *FallbackNormalConfig `yaml:"fallbackNormal,omitempty"`
	FallbackPrysm  *FallbackPrysmConfig  `yaml:"fallbackPrysm,omitempty"`

	// Metrics
	Grafana           *GrafanaConfig           `yaml:"grafana,omitempty"`
	Prometheus        *PrometheusConfig        `yaml:"prometheus,omitempty"`
	Exporter          *ExporterConfig          `yaml:"exporter,omitempty"`
	BitflyNodeMetrics *BitflyNodeMetricsConfig `yaml:"bitflyNodeMetrics,omitempty"`

	// Native mode
	Native *NativeConfig `yaml:"native,omitempty"`

	// MEV-Boost
	EnableMevBoost config.Parameter `yaml:"enableMevBoost,omitempty"`
	MevBoost       *MevBoostConfig  `yaml:"mevBoost,omitempty"`

	// Addons
	GraffitiWallWriter addontypes.SmartnodeAddon `yaml:"addon-gww,omitempty"`
}

// Load configuration settings from a file
func LoadFromFile(path string) (*RocketPoolConfig, error) {

	// Return nil if the file doesn't exist
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	// Read the file
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read Rocket Pool settings file at %s: %w", shellescape.Quote(path), err)
	}

	// Attempt to parse it out into a settings map
	var settings map[string]map[string]string
	if err := yaml.Unmarshal(configBytes, &settings); err != nil {
		return nil, fmt.Errorf("could not parse settings file: %w", err)
	}

	// Deserialize it into a config object
	cfg := NewRocketPoolConfig(filepath.Dir(path), false)
	err = cfg.Deserialize(settings)
	if err != nil {
		return nil, fmt.Errorf("could not deserialize settings file: %w", err)
	}

	return cfg, nil

}

// Creates a new Rocket Pool configuration instance
func NewRocketPoolConfig(rpDir string, isNativeMode bool) *RocketPoolConfig {

	clientModes := []config.ParameterOption{{
		Name:        "Locally Managed",
		Description: "Allow the Smartnode to manage the Execution and Consensus clients for you (Docker Mode)",
		Value:       config.Mode_Local,
	}, {
		Name:        "Externally Managed",
		Description: "Use existing Execution and Consensus clients that you manage on your own (Hybrid Mode)",
		Value:       config.Mode_External,
	}}

	cfg := &RocketPoolConfig{
		Title:               "Top-level Settings",
		RocketPoolDirectory: rpDir,
		IsNativeMode:        isNativeMode,

		ExecutionClientMode: config.Parameter{
			ID:                   "executionClientMode",
			Name:                 "Execution Client Mode",
			Description:          "Choose which mode to use for your Execution client - locally managed (Docker Mode), or externally managed (Hybrid Mode).",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Eth1, config.ContainerID_Eth2, config.ContainerID_Node, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options:              clientModes,
		},

		ExecutionClient: config.Parameter{
			ID:                   "executionClient",
			Name:                 "Execution Client",
			Description:          "Select which Execution client you would like to run.",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: config.ExecutionClient_Geth},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []config.ParameterOption{{
				Name:        "Geth",
				Description: "Geth is one of the three original implementations of the Ethereum protocol. It is written in Go, fully open source and licensed under the GNU LGPL v3.",
				Value:       config.ExecutionClient_Geth,
			}, {
				Name:        "Nethermind",
				Description: getAugmentedEcDescription(config.ExecutionClient_Nethermind, "Nethermind is a high-performance full Ethereum protocol client with very fast sync speeds. Nethermind is built with proven industrial technologies such as .NET 6 and the Kestrel web server. It is fully open source."),
				Value:       config.ExecutionClient_Nethermind,
			}, {
				Name:        "Besu",
				Description: getAugmentedEcDescription(config.ExecutionClient_Besu, "Hyperledger Besu is a robust full Ethereum protocol client. It uses a novel system called \"Bonsai Trees\" to store its chain data efficiently, which allows it to access block states from the past and does not require pruning. Besu is fully open source and written in Java."),
				Value:       config.ExecutionClient_Besu,
			}},
		},

		UseFallbackClients: config.Parameter{
			ID:                   "useFallbackClients",
			Name:                 "Use Fallback Clients",
			Description:          "Enable this if you would like to specify a fallback Execution and Consensus Client, which will temporarily be used by the Smartnode and your Validator Client if your primary Execution / Consensus client pair ever go offline (e.g. if you switch, prune, or resync your clients).",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Validator, config.ContainerID_Node, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ReconnectDelay: config.Parameter{
			ID:                   "reconnectDelay",
			Name:                 "Reconnect Delay",
			Description:          "The delay to wait after your primary Execution or Consensus clients fail before trying to reconnect to them. An example format is \"10h20m30s\" - this would make it 10 hours, 20 minutes, and 30 seconds.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: "60s"},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ConsensusClientMode: config.Parameter{
			ID:                   "consensusClientMode",
			Name:                 "Consensus Client Mode",
			Description:          "Choose which mode to use for your Consensus client - locally managed (Docker Mode), or externally managed (Hybrid Mode).",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: config.Mode_Local},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Eth2, config.ContainerID_Node, config.ContainerID_Prometheus, config.ContainerID_Validator, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options:              clientModes,
		},

		ConsensusClient: config.Parameter{
			ID:                   "consensusClient",
			Name:                 "Consensus Client",
			Description:          "Select which Consensus client you would like to use.",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: config.ConsensusClient_Nimbus},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth2, config.ContainerID_Validator},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []config.ParameterOption{{
				Name:        "Lighthouse",
				Description: "Lighthouse is a Consensus client with a heavy focus on speed and security. The team behind it, Sigma Prime, is an information security and software engineering firm who have funded Lighthouse along with the Ethereum Foundation, Consensys, and private individuals. Lighthouse is built in Rust and offered under an Apache 2.0 License.",
				Value:       config.ConsensusClient_Lighthouse,
			}, {
				Name:        "Lodestar",
				Description: "Lodestar is the fifth open-source Ethereum consensus client. It is written in Typescript maintained by ChainSafe Systems. Lodestar, their flagship product, is a production-capable Beacon Chain and Validator Client uniquely situated as the go-to for researchers and developers for rapid prototyping and browser usage.",
				Value:       config.ConsensusClient_Lodestar,
			}, {
				Name:        "Nimbus",
				Description: "Nimbus is a Consensus client implementation that strives to be as lightweight as possible in terms of resources used. This allows it to perform well on embedded systems, resource-restricted devices -- including Raspberry Pis and mobile devices -- and multi-purpose servers.",
				Value:       config.ConsensusClient_Nimbus,
			}, {
				Name:        "Prysm",
				Description: "Prysm is a Go implementation of Ethereum Consensus protocol with a focus on usability, security, and reliability. Prysm is developed by Prysmatic Labs, a company with the sole focus on the development of their client. Prysm is written in Go and released under a GPL-3.0 license.",
				Value:       config.ConsensusClient_Prysm,
			}, {
				Name:        "Teku",
				Description: "PegaSys Teku (formerly known as Artemis) is a Java-based Ethereum 2.0 client designed & built to meet institutional needs and security requirements. PegaSys is an arm of ConsenSys dedicated to building enterprise-ready clients and tools for interacting with the core Ethereum platform. Teku is Apache 2 licensed and written in Java, a language notable for its maturity & ubiquity.",
				Value:       config.ConsensusClient_Teku,
			}},
		},

		ExternalConsensusClient: config.Parameter{
			ID:                   "externalConsensusClient",
			Name:                 "Consensus Client",
			Description:          "Select which Consensus client your externally managed client is.",
			Type:                 config.ParameterType_Choice,
			Default:              map[config.Network]interface{}{config.Network_All: config.ConsensusClient_Lighthouse},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth2, config.ContainerID_Validator},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []config.ParameterOption{{
				Name:        "Lighthouse",
				Description: "Select this if you will use Lighthouse as your Consensus client.",
				Value:       config.ConsensusClient_Lighthouse,
			}, {
				Name:        "Lodestar",
				Description: "Select this if you will use Lodestar as your Consensus client.",
				Value:       config.ConsensusClient_Lodestar,
			}, {
				Name:        "Nimbus",
				Description: "Select this if you will use Nimbus as your Consensus client.",
				Value:       config.ConsensusClient_Nimbus,
			}, {
				Name:        "Prysm",
				Description: "Select this if you will use Prysm as your Consensus client.",
				Value:       config.ConsensusClient_Prysm,
			}, {
				Name:        "Teku",
				Description: "Select this if you will use Teku as your Consensus client.",
				Value:       config.ConsensusClient_Teku,
			}},
		},

		EnableMetrics: config.Parameter{
			ID:                   "enableMetrics",
			Name:                 "Enable Metrics",
			Description:          "Enable the Smartnode's performance and status metrics system. This will provide you with the node operator's Grafana dashboard.",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: true},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth2, config.ContainerID_Grafana, config.ContainerID_Prometheus, config.ContainerID_Exporter},
			EnvironmentVariables: []string{"ENABLE_METRICS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		EnableODaoMetrics: config.Parameter{
			ID:                   "enableODaoMetrics",
			Name:                 "Enable Oracle DAO Metrics",
			Description:          "Enable the tracking of Oracle DAO performance metrics, such as prices and balances submission participation.",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Node},
			EnvironmentVariables: []string{"ENABLE_ODAO_METRICS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		EnableBitflyNodeMetrics: config.Parameter{
			ID:                   "enableBitflyNodeMetrics",
			Name:                 "Enable Beaconcha.in Node Metrics",
			Description:          "Enable the Beaconcha.in node metrics integration. This will allow you to track your node's metrics from your phone using the Beaconcha.in App.\n\nFor more information on setting up an account and the app, please visit https://beaconcha.in/mobile.",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator, config.ContainerID_Eth2},
			EnvironmentVariables: []string{"ENABLE_BITFLY_NODE_METRICS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		EcMetricsPort: config.Parameter{
			ID:                   "ecMetricsPort",
			Name:                 "Execution Client Metrics Port",
			Description:          "The port your Execution client should expose its metrics on.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: defaultEcMetricsPort},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1, config.ContainerID_Prometheus},
			EnvironmentVariables: []string{"EC_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		BnMetricsPort: config.Parameter{
			ID:                   "bnMetricsPort",
			Name:                 "Beacon Node Metrics Port",
			Description:          "The port your Consensus client's Beacon Node should expose its metrics on.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: defaultBnMetricsPort},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_Prometheus},
			EnvironmentVariables: []string{"BN_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		VcMetricsPort: config.Parameter{
			ID:                   "vcMetricsPort",
			Name:                 "Validator Client Metrics Port",
			Description:          "The port your validator client should expose its metrics on.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: defaultVcMetricsPort},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator, config.ContainerID_Prometheus},
			EnvironmentVariables: []string{"VC_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		NodeMetricsPort: config.Parameter{
			ID:                   "nodeMetricsPort",
			Name:                 "Node Metrics Port",
			Description:          "The port your Node container should expose its metrics on.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: defaultNodeMetricsPort},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Node, config.ContainerID_Prometheus},
			EnvironmentVariables: []string{"NODE_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ExporterMetricsPort: config.Parameter{
			ID:                   "exporterMetricsPort",
			Name:                 "Exporter Metrics Port",
			Description:          "The port that Prometheus's Node Exporter should expose its metrics on.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: defaultExporterMetricsPort},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Exporter, config.ContainerID_Prometheus},
			EnvironmentVariables: []string{"EXPORTER_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		WatchtowerMetricsPort: config.Parameter{
			ID:                   "watchtowerMetricsPort",
			Name:                 "Watchtower Metrics Port",
			Description:          "The port your Watchtower container should expose its metrics on.\nThis is only relevant for Oracle Nodes.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: defaultWatchtowerMetricsPort},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Watchtower, config.ContainerID_Prometheus},
			EnvironmentVariables: []string{"WATCHTOWER_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		EnableMevBoost: config.Parameter{
			ID:                   "enableMevBoost",
			Name:                 "Enable MEV-Boost",
			Description:          "Enable MEV-Boost, which connects your validator to one or more relays of your choice. The relays act as intermediaries between you and professional block builders that find and extract MEV opportunities. The builders will give you a healthy tip in return, which tends to be worth more than blocks you built on your own.\n\n[orange]NOTE: This toggle is temporary during the early Merge days while relays are still being created. It will be removed in the future.",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: true},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_MevBoost},
			EnvironmentVariables: []string{"ENABLE_MEV_BOOST"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},
	}

	// Set the defaults for choices
	cfg.ExecutionClientMode.Default[config.Network_All] = cfg.ExecutionClientMode.Options[0].Value
	cfg.ConsensusClientMode.Default[config.Network_All] = cfg.ConsensusClientMode.Options[0].Value

	cfg.Smartnode = NewSmartnodeConfig(cfg)
	cfg.ExecutionCommon = NewExecutionCommonConfig(cfg)
	cfg.Geth = NewGethConfig(cfg)
	cfg.Nethermind = NewNethermindConfig(cfg)
	cfg.Besu = NewBesuConfig(cfg)
	cfg.ExternalExecution = NewExternalExecutionConfig(cfg)
	cfg.FallbackNormal = NewFallbackNormalConfig(cfg)
	cfg.FallbackPrysm = NewFallbackPrysmConfig(cfg)
	cfg.ConsensusCommon = NewConsensusCommonConfig(cfg)
	cfg.Lighthouse = NewLighthouseConfig(cfg)
	cfg.Lodestar = NewLodestarConfig(cfg)
	cfg.Nimbus = NewNimbusConfig(cfg)
	cfg.Prysm = NewPrysmConfig(cfg)
	cfg.Teku = NewTekuConfig(cfg)
	cfg.ExternalLighthouse = NewExternalLighthouseConfig(cfg)
	cfg.ExternalLodestar = NewExternalLodestarConfig(cfg)
	cfg.ExternalNimbus = NewExternalNimbusConfig(cfg)
	cfg.ExternalPrysm = NewExternalPrysmConfig(cfg)
	cfg.ExternalTeku = NewExternalTekuConfig(cfg)
	cfg.Grafana = NewGrafanaConfig(cfg)
	cfg.Prometheus = NewPrometheusConfig(cfg)
	cfg.Exporter = NewExporterConfig(cfg)
	cfg.BitflyNodeMetrics = NewBitflyNodeMetricsConfig(cfg)
	cfg.Native = NewNativeConfig(cfg)
	cfg.MevBoost = NewMevBoostConfig(cfg)

	// Addons
	cfg.GraffitiWallWriter = addons.NewGraffitiWallWriter()

	// Apply the default values for mainnet
	cfg.Smartnode.Network.Value = cfg.Smartnode.Network.Options[0].Value
	cfg.applyAllDefaults()

	return cfg
}

// Get a more verbose client description, including warnings
func getAugmentedEcDescription(client config.ExecutionClient, originalDescription string) string {

	switch client {
	case config.ExecutionClient_Nethermind:
		totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
		if totalMemoryGB < 9 {
			return fmt.Sprintf("%s\n\n[red]WARNING: Nethermind currently requires over 8 GB of RAM to run smoothly. We do not recommend it for your system. This may be improved in a future release.", originalDescription)
		}
	}

	return originalDescription

}

// Create a copy of this configuration.
func (cfg *RocketPoolConfig) CreateCopy() *RocketPoolConfig {
	newConfig := NewRocketPoolConfig(cfg.RocketPoolDirectory, cfg.IsNativeMode)

	// Set the network
	network := cfg.Smartnode.Network.Value.(config.Network)
	newConfig.Smartnode.Network.Value = network

	newParams := newConfig.GetParameters()
	for i, param := range cfg.GetParameters() {
		newParams[i].Value = param.Value
		newParams[i].UpdateDescription(network)
	}

	newSubconfigs := newConfig.GetSubconfigs()
	for name, subConfig := range cfg.GetSubconfigs() {
		newParams := newSubconfigs[name].GetParameters()
		for i, param := range subConfig.GetParameters() {
			newParams[i].Value = param.Value
			newParams[i].UpdateDescription(network)
		}
	}

	return newConfig
}

// Get the parameters for this config
func (cfg *RocketPoolConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.ExecutionClientMode,
		&cfg.ExecutionClient,
		&cfg.UseFallbackClients,
		&cfg.ReconnectDelay,
		&cfg.ConsensusClientMode,
		&cfg.ConsensusClient,
		&cfg.ExternalConsensusClient,
		&cfg.EnableMetrics,
		&cfg.EnableODaoMetrics,
		&cfg.EnableBitflyNodeMetrics,
		&cfg.EcMetricsPort,
		&cfg.BnMetricsPort,
		&cfg.VcMetricsPort,
		&cfg.NodeMetricsPort,
		&cfg.ExporterMetricsPort,
		&cfg.WatchtowerMetricsPort,
		&cfg.EnableMevBoost,
	}
}

// Get the subconfigurations for this config
func (cfg *RocketPoolConfig) GetSubconfigs() map[string]config.Config {
	return map[string]config.Config{
		"smartnode":          cfg.Smartnode,
		"executionCommon":    cfg.ExecutionCommon,
		"geth":               cfg.Geth,
		"nethermind":         cfg.Nethermind,
		"besu":               cfg.Besu,
		"externalExecution":  cfg.ExternalExecution,
		"consensusCommon":    cfg.ConsensusCommon,
		"lighthouse":         cfg.Lighthouse,
		"lodestar":           cfg.Lodestar,
		"nimbus":             cfg.Nimbus,
		"prysm":              cfg.Prysm,
		"teku":               cfg.Teku,
		"externalLighthouse": cfg.ExternalLighthouse,
		"externalLodestar":   cfg.ExternalLodestar,
		"externalNimbus":     cfg.ExternalNimbus,
		"externalPrysm":      cfg.ExternalPrysm,
		"externalTeku":       cfg.ExternalTeku,
		"fallbackNormal":     cfg.FallbackNormal,
		"fallbackPrysm":      cfg.FallbackPrysm,
		"grafana":            cfg.Grafana,
		"prometheus":         cfg.Prometheus,
		"exporter":           cfg.Exporter,
		"bitflyNodeMetrics":  cfg.BitflyNodeMetrics,
		"native":             cfg.Native,
		"mevBoost":           cfg.MevBoost,
		"addons-gww":         cfg.GraffitiWallWriter.GetConfig(),
	}
}

// Handle a network change on all of the parameters
func (cfg *RocketPoolConfig) ChangeNetwork(newNetwork config.Network) {

	// Get the current network
	oldNetwork, ok := cfg.Smartnode.Network.Value.(config.Network)
	if !ok {
		oldNetwork = config.Network_Unknown
	}
	if oldNetwork == newNetwork {
		return
	}
	cfg.Smartnode.Network.Value = newNetwork

	// Update the master parameters
	rootParams := cfg.GetParameters()
	for _, param := range rootParams {
		param.ChangeNetwork(oldNetwork, newNetwork)
	}

	// Update all of the child config objects
	subconfigs := cfg.GetSubconfigs()
	for _, subconfig := range subconfigs {
		for _, param := range subconfig.GetParameters() {
			param.ChangeNetwork(oldNetwork, newNetwork)
		}
	}

}

// Get the configuration for the selected execution client
func (cfg *RocketPoolConfig) GetEventLogInterval() (int, error) {
	if cfg.IsNativeMode {
		return gethEventLogInterval, nil
	}

	mode := cfg.ExecutionClientMode.Value.(config.Mode)
	switch mode {
	case config.Mode_Local:
		client := cfg.ExecutionClient.Value.(config.ExecutionClient)
		switch client {
		case config.ExecutionClient_Besu:
			return cfg.Besu.EventLogInterval, nil
		case config.ExecutionClient_Geth:
			return cfg.Geth.EventLogInterval, nil
		case config.ExecutionClient_Nethermind:
			return cfg.Nethermind.EventLogInterval, nil
		default:
			return 0, fmt.Errorf("can't get event log interval of unknown execution client [%v]", client)
		}

	case config.Mode_External:
		return gethEventLogInterval, nil

	default:
		return 0, fmt.Errorf("can't get event log interval of unknown execution client mode [%v]", mode)
	}
}

// Get the selected CC and mode
func (cfg *RocketPoolConfig) GetSelectedConsensusClient() (config.ConsensusClient, config.Mode) {
	mode := cfg.ConsensusClientMode.Value.(config.Mode)
	var cc config.ConsensusClient
	if mode == config.Mode_Local {
		cc = cfg.ConsensusClient.Value.(config.ConsensusClient)
	} else {
		cc = cfg.ExternalConsensusClient.Value.(config.ConsensusClient)
	}
	return cc, mode
}

// Get the configuration for the selected consensus client
func (cfg *RocketPoolConfig) GetSelectedConsensusClientConfig() (config.ConsensusConfig, error) {
	if cfg.IsNativeMode {
		return nil, fmt.Errorf("consensus config is not available in native mode")
	}

	mode := cfg.ConsensusClientMode.Value.(config.Mode)
	switch mode {
	case config.Mode_Local:
		client := cfg.ConsensusClient.Value.(config.ConsensusClient)
		switch client {
		case config.ConsensusClient_Lighthouse:
			return cfg.Lighthouse, nil
		case config.ConsensusClient_Lodestar:
			return cfg.Lodestar, nil
		case config.ConsensusClient_Nimbus:
			return cfg.Nimbus, nil
		case config.ConsensusClient_Prysm:
			return cfg.Prysm, nil
		case config.ConsensusClient_Teku:
			return cfg.Teku, nil
		default:
			return nil, fmt.Errorf("unknown consensus client [%v] selected", client)
		}

	case config.Mode_External:
		client := cfg.ExternalConsensusClient.Value.(config.ConsensusClient)
		switch client {
		case config.ConsensusClient_Lighthouse:
			return cfg.ExternalLighthouse, nil
		case config.ConsensusClient_Lodestar:
			return cfg.ExternalLodestar, nil
		case config.ConsensusClient_Nimbus:
			return cfg.ExternalNimbus, nil
		case config.ConsensusClient_Prysm:
			return cfg.ExternalPrysm, nil
		case config.ConsensusClient_Teku:
			return cfg.ExternalTeku, nil
		default:
			return nil, fmt.Errorf("unknown external consensus client [%v] selected", client)
		}

	default:
		return nil, fmt.Errorf("unknown consensus client mode [%v]", mode)
	}
}

// Check if doppelganger protection is enabled
func (cfg *RocketPoolConfig) IsDoppelgangerEnabled() (bool, error) {
	if cfg.IsNativeMode {
		return false, fmt.Errorf("consensus config is not available in native mode")
	}

	mode := cfg.ConsensusClientMode.Value.(config.Mode)
	switch mode {
	case config.Mode_Local:
		client := cfg.ConsensusClient.Value.(config.ConsensusClient)
		switch client {
		case config.ConsensusClient_Lighthouse, config.ConsensusClient_Lodestar, config.ConsensusClient_Nimbus, config.ConsensusClient_Prysm:
			return cfg.ConsensusCommon.DoppelgangerDetection.Value.(bool), nil
		case config.ConsensusClient_Teku:
			return false, nil
		default:
			return false, fmt.Errorf("unknown consensus client [%v] selected", client)
		}

	case config.Mode_External:
		client := cfg.ExternalConsensusClient.Value.(config.ConsensusClient)
		switch client {
		case config.ConsensusClient_Lighthouse:
			return cfg.ExternalLighthouse.DoppelgangerDetection.Value.(bool), nil
		case config.ConsensusClient_Lodestar:
			return cfg.ExternalLodestar.DoppelgangerDetection.Value.(bool), nil
		case config.ConsensusClient_Nimbus:
			return cfg.ExternalNimbus.DoppelgangerDetection.Value.(bool), nil
		case config.ConsensusClient_Prysm:
			return cfg.ExternalPrysm.DoppelgangerDetection.Value.(bool), nil
		case config.ConsensusClient_Teku:
			return false, nil
		default:
			return false, fmt.Errorf("unknown external consensus client [%v] selected", client)
		}

	default:
		return false, fmt.Errorf("unknown consensus client mode [%v]", mode)
	}
}

// Serializes the configuration into a map of maps, compatible with a settings file
func (cfg *RocketPoolConfig) Serialize() map[string]map[string]string {

	masterMap := map[string]map[string]string{}

	// Serialize root params
	rootParams := map[string]string{}
	for _, param := range cfg.GetParameters() {
		param.Serialize(rootParams)
	}
	masterMap[rootConfigName] = rootParams
	masterMap[rootConfigName]["rpDir"] = cfg.RocketPoolDirectory
	masterMap[rootConfigName]["isNative"] = fmt.Sprint(cfg.IsNativeMode)
	masterMap[rootConfigName]["version"] = fmt.Sprintf("v%s", shared.RocketPoolVersion) // Update the version with the current Smartnode version

	// Serialize the subconfigs
	for name, subconfig := range cfg.GetSubconfigs() {
		subconfigParams := map[string]string{}
		for _, param := range subconfig.GetParameters() {
			param.Serialize(subconfigParams)
		}
		masterMap[name] = subconfigParams
	}

	return masterMap
}

// Deserializes a settings file into this config
func (cfg *RocketPoolConfig) Deserialize(masterMap map[string]map[string]string) error {

	// Upgrade the config to the latest version
	err := migration.UpdateConfig(masterMap)
	if err != nil {
		return fmt.Errorf("error upgrading configuration to v%s: %w", shared.RocketPoolVersion, err)
	}

	// Get the network
	network := config.Network_Mainnet
	smartnodeConfig, exists := masterMap["smartnode"]
	if exists {
		networkString, exists := smartnodeConfig[cfg.Smartnode.Network.ID]
		if exists {
			valueType := reflect.TypeOf(networkString)
			paramType := reflect.TypeOf(network)
			if !valueType.ConvertibleTo(paramType) {
				return fmt.Errorf("can't get default network: value type %s cannot be converted to parameter type %s", valueType.Name(), paramType.Name())
			}
			network = reflect.ValueOf(networkString).Convert(paramType).Interface().(config.Network)
		}
	}

	// Deserialize root params
	rootParams := masterMap[rootConfigName]
	for _, param := range cfg.GetParameters() {
		// Note: if the root config doesn't exist, this will end up using the default values for all of its settings
		err := param.Deserialize(rootParams, network)
		if err != nil {
			return fmt.Errorf("error deserializing root config: %w", err)
		}
	}

	cfg.RocketPoolDirectory = masterMap[rootConfigName]["rpDir"]
	cfg.IsNativeMode, err = strconv.ParseBool(masterMap[rootConfigName]["isNative"])
	if err != nil {
		return fmt.Errorf("error parsing isNative: %w", err)
	}
	cfg.Version = masterMap[rootConfigName]["version"]

	// Deserialize the subconfigs
	for name, subconfig := range cfg.GetSubconfigs() {
		subconfigParams := masterMap[name]
		for _, param := range subconfig.GetParameters() {
			// Note: if the subconfig doesn't exist, this will end up using the default values for all of its settings
			err := param.Deserialize(subconfigParams, network)
			if err != nil {
				return fmt.Errorf("error deserializing [%s]: %w", name, err)
			}
		}
	}

	return nil
}

// Generates a collection of environment variables based on this config's settings
func (cfg *RocketPoolConfig) GenerateEnvironmentVariables() map[string]string {

	envVars := map[string]string{}

	// Basic variables and root parameters
	envVars["SMARTNODE_IMAGE"] = cfg.Smartnode.GetSmartnodeContainerTag()
	envVars["ROCKETPOOL_FOLDER"] = cfg.RocketPoolDirectory
	envVars["RETH_ADDRESS"] = cfg.Smartnode.GetRethAddress().Hex()
	envVars[FeeRecipientFileEnvVar] = FeeRecipientFilename // If this is running, we're in Docker mode by definition so use the Docker fee recipient filename
	config.AddParametersToEnvVars(cfg.Smartnode.GetParameters(), envVars)
	config.AddParametersToEnvVars(cfg.GetParameters(), envVars)

	// EC parameters
	if cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_Local {
		envVars["EC_CLIENT"] = fmt.Sprint(cfg.ExecutionClient.Value)
		envVars["EC_HTTP_ENDPOINT"] = fmt.Sprintf("http://%s:%d", Eth1ContainerName, cfg.ExecutionCommon.HttpPort.Value)
		envVars["EC_WS_ENDPOINT"] = fmt.Sprintf("ws://%s:%d", Eth1ContainerName, cfg.ExecutionCommon.WsPort.Value)
		envVars["EC_ENGINE_ENDPOINT"] = fmt.Sprintf("http://%s:%d", Eth1ContainerName, cfg.ExecutionCommon.EnginePort.Value)
		envVars["EC_ENGINE_WS_ENDPOINT"] = fmt.Sprintf("ws://%s:%d", Eth1ContainerName, cfg.ExecutionCommon.EnginePort.Value)

		// Handle open API ports
		rpcMode := cfg.ExecutionCommon.OpenRpcPorts.Value.(config.RPCMode)
		if rpcMode.Open() {
			httpMapping := rpcMode.DockerPortMapping(cfg.ExecutionCommon.HttpPort.Value.(uint16))
			wsMapping := rpcMode.DockerPortMapping(cfg.ExecutionCommon.WsPort.Value.(uint16))
			envVars["EC_OPEN_API_PORTS"] = fmt.Sprintf(", \"%s\", \"%s\"", httpMapping, wsMapping)
		}

		// Common params
		config.AddParametersToEnvVars(cfg.ExecutionCommon.GetParameters(), envVars)

		// Client-specific params
		switch cfg.ExecutionClient.Value.(config.ExecutionClient) {
		case config.ExecutionClient_Geth:
			config.AddParametersToEnvVars(cfg.Geth.GetParameters(), envVars)
			envVars["EC_STOP_SIGNAL"] = gethStopSignal
		case config.ExecutionClient_Nethermind:
			config.AddParametersToEnvVars(cfg.Nethermind.GetParameters(), envVars)
			envVars["EC_STOP_SIGNAL"] = nethermindStopSignal
		case config.ExecutionClient_Besu:
			config.AddParametersToEnvVars(cfg.Besu.GetParameters(), envVars)
			envVars["EC_STOP_SIGNAL"] = besuStopSignal
		}
	} else {
		envVars["EC_CLIENT"] = "X" // X is for external / unknown
		config.AddParametersToEnvVars(cfg.ExternalExecution.GetParameters(), envVars)
	}
	// Get the hostname of the Execution client, necessary for Prometheus to work in hybrid mode
	ecUrl, err := url.Parse(envVars["EC_HTTP_ENDPOINT"])
	if err == nil && ecUrl != nil {
		envVars["EC_HOSTNAME"] = ecUrl.Hostname()
	}

	// CC parameters
	var consensusClient config.ConsensusClient
	if cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_Local {
		consensusClient = cfg.ConsensusClient.Value.(config.ConsensusClient)
		envVars["CC_API_ENDPOINT"] = fmt.Sprintf("http://%s:%d", Eth2ContainerName, cfg.ConsensusCommon.ApiPort.Value)

		// Handle open API ports
		bnOpenPorts := ""
		apiPortMode := cfg.ConsensusCommon.OpenApiPort.Value.(config.RPCMode)
		if apiPortMode.Open() {
			apiPort := cfg.ConsensusCommon.ApiPort.Value.(uint16)
			bnOpenPorts += fmt.Sprintf(", \"%s\"", apiPortMode.DockerPortMapping(apiPort))
		}
		if consensusClient == config.ConsensusClient_Prysm {
			prysmRpcPortMode := cfg.Prysm.OpenRpcPort.Value.(config.RPCMode)
			if prysmRpcPortMode.Open() {
				prysmRpcPort := cfg.Prysm.RpcPort.Value.(uint16)
				bnOpenPorts += fmt.Sprintf(", \"%s\"", prysmRpcPortMode.DockerPortMapping(prysmRpcPort))
			}
		}

		envVars["BN_OPEN_PORTS"] = bnOpenPorts

		// Common params
		config.AddParametersToEnvVars(cfg.ConsensusCommon.GetParameters(), envVars)

		// Client-specific params
		switch consensusClient {
		case config.ConsensusClient_Lighthouse:
			config.AddParametersToEnvVars(cfg.Lighthouse.GetParameters(), envVars)
		case config.ConsensusClient_Lodestar:
			config.AddParametersToEnvVars(cfg.Lodestar.GetParameters(), envVars)
		case config.ConsensusClient_Nimbus:
			config.AddParametersToEnvVars(cfg.Nimbus.GetParameters(), envVars)
		case config.ConsensusClient_Prysm:
			config.AddParametersToEnvVars(cfg.Prysm.GetParameters(), envVars)
			envVars["CC_RPC_ENDPOINT"] = fmt.Sprintf("http://%s:%d", Eth2ContainerName, cfg.Prysm.RpcPort.Value)
		case config.ConsensusClient_Teku:
			config.AddParametersToEnvVars(cfg.Teku.GetParameters(), envVars)
		}
	} else {
		consensusClient = cfg.ExternalConsensusClient.Value.(config.ConsensusClient)

		switch consensusClient {
		case config.ConsensusClient_Lighthouse:
			config.AddParametersToEnvVars(cfg.ExternalLighthouse.GetParameters(), envVars)
		case config.ConsensusClient_Lodestar:
			config.AddParametersToEnvVars(cfg.ExternalLodestar.GetParameters(), envVars)
		case config.ConsensusClient_Nimbus:
			config.AddParametersToEnvVars(cfg.ExternalNimbus.GetParameters(), envVars)
		case config.ConsensusClient_Prysm:
			config.AddParametersToEnvVars(cfg.ExternalPrysm.GetParameters(), envVars)
		case config.ConsensusClient_Teku:
			config.AddParametersToEnvVars(cfg.ExternalTeku.GetParameters(), envVars)
		}
	}
	envVars["CC_CLIENT"] = fmt.Sprint(consensusClient)

	// Graffiti
	identifier := ""
	versionString := fmt.Sprintf("v%s", shared.RocketPoolVersion)
	envVars["ROCKET_POOL_VERSION"] = versionString
	if len(versionString) < 8 {
		ecInitial := strings.ToUpper(string(envVars["EC_CLIENT"][0]))

		var ccInitial string
		switch consensusClient {
		case config.ConsensusClient_Lodestar:
			ccInitial = "S" // Lodestar is special because it conflicts with Lighthouse
		default:
			ccInitial = strings.ToUpper(string(envVars["CC_CLIENT"][0]))
		}
		identifier = fmt.Sprintf("-%s%s", ecInitial, ccInitial)
	}

	graffitiPrefix := fmt.Sprintf("RP%s %s", identifier, versionString)
	envVars["GRAFFITI_PREFIX"] = graffitiPrefix

	customGraffiti := envVars[CustomGraffitiEnvVar]
	if customGraffiti == "" {
		envVars["GRAFFITI"] = graffitiPrefix
	} else {
		envVars["GRAFFITI"] = fmt.Sprintf("%s (%s)", graffitiPrefix, customGraffiti)
	}

	// Get the hostname of the Consensus client, necessary for Prometheus to work in hybrid mode
	ccUrl, err := url.Parse(envVars["CC_API_ENDPOINT"])
	if err == nil && ccUrl != nil {
		envVars["CC_HOSTNAME"] = ccUrl.Hostname()
	}

	// Fallback parameters
	if cfg.UseFallbackClients.Value == true {
		switch consensusClient {
		case config.ConsensusClient_Prysm:
			config.AddParametersToEnvVars(cfg.FallbackPrysm.GetParameters(), envVars)
		default:
			config.AddParametersToEnvVars(cfg.FallbackNormal.GetParameters(), envVars)
		}
	}

	// Metrics
	if cfg.EnableMetrics.Value == true {
		config.AddParametersToEnvVars(cfg.Exporter.GetParameters(), envVars)
		config.AddParametersToEnvVars(cfg.Prometheus.GetParameters(), envVars)
		config.AddParametersToEnvVars(cfg.Grafana.GetParameters(), envVars)

		if cfg.Exporter.RootFs.Value == true {
			envVars["EXPORTER_ROOTFS_COMMAND"] = ", \"--path.rootfs=/rootfs\""
			envVars["EXPORTER_ROOTFS_VOLUME"] = ", \"/:/rootfs:ro\""
		}

		portMode := cfg.Prometheus.OpenPort.Value.(config.RPCMode)
		if portMode.Open() {
			envVars["PROMETHEUS_OPEN_PORTS"] = fmt.Sprintf("\"%s\"", portMode.DockerPortMapping(cfg.Prometheus.Port.Value.(uint16)))
		}

		// Additional metrics flags
		if cfg.Exporter.AdditionalFlags.Value.(string) != "" {
			envVars["EXPORTER_ADDITIONAL_FLAGS"] = fmt.Sprintf(", \"%s\"", cfg.Exporter.AdditionalFlags.Value.(string))
		}
		if cfg.Prometheus.AdditionalFlags.Value.(string) != "" {
			envVars["PROMETHEUS_ADDITIONAL_FLAGS"] = fmt.Sprintf(", \"%s\"", cfg.Prometheus.AdditionalFlags.Value.(string))
		}
	}

	// Bitfly Node Metrics
	if cfg.EnableBitflyNodeMetrics.Value == true {
		config.AddParametersToEnvVars(cfg.BitflyNodeMetrics.GetParameters(), envVars)
	}

	// MEV-Boost
	if cfg.EnableMevBoost.Value == true {
		// Disable for Holesky
		if cfg.Smartnode.Network.Value == config.Network_Holesky {
			cfg.EnableMevBoost.Value = false
		} else {
			config.AddParametersToEnvVars(cfg.MevBoost.GetParameters(), envVars)
			if cfg.MevBoost.Mode.Value == config.Mode_Local {
				envVars[mevBoostRelaysEnvVar] = cfg.MevBoost.GetRelayString()
				envVars[mevBoostUrlEnvVar] = fmt.Sprintf("http://%s:%d", MevBoostContainerName, cfg.MevBoost.Port.Value)

				// Handle open API port
				portMode := cfg.MevBoost.OpenRpcPort.Value.(config.RPCMode)
				if portMode.Open() {
					port := cfg.MevBoost.Port.Value.(uint16)
					envVars["MEV_BOOST_OPEN_API_PORT"] = fmt.Sprintf("\"%s\"", portMode.DockerPortMapping(port))
				}
			}
		}
	}

	// Addons
	cfg.GraffitiWallWriter.UpdateEnvVars(envVars)

	return envVars

}

// The the title for the config
func (cfg *RocketPoolConfig) GetConfigTitle() string {
	return cfg.Title
}

// Update the default settings for all overwrite-on-upgrade parameters
func (cfg *RocketPoolConfig) UpdateDefaults() error {
	// Update the root params
	currentNetwork := cfg.Smartnode.Network.Value.(config.Network)
	for _, param := range cfg.GetParameters() {
		defaultValue, err := param.GetDefault(currentNetwork)
		if err != nil {
			return fmt.Errorf("error getting defaults for root param [%s] on network [%v]: %w", param.ID, currentNetwork, err)
		}
		if param.OverwriteOnUpgrade {
			param.Value = defaultValue
		}
	}

	// Update the subconfigs
	for subconfigName, subconfig := range cfg.GetSubconfigs() {
		for _, param := range subconfig.GetParameters() {
			defaultValue, err := param.GetDefault(currentNetwork)
			if err != nil {
				return fmt.Errorf("error getting defaults for %s param [%s] on network [%v]: %w", subconfigName, param.ID, currentNetwork, err)
			}
			if param.OverwriteOnUpgrade {
				param.Value = defaultValue
			}
		}
	}

	return nil
}

// Get all of the settings that have changed between an old config and this config, and get all of the containers that are affected by those changes - also returns whether or not the selected network was changed
func (cfg *RocketPoolConfig) GetChanges(oldConfig *RocketPoolConfig) (map[string][]config.ChangedSetting, map[config.ContainerID]bool, bool) {
	// Get the map of changed settings by category
	changedSettings := getChangedSettingsMap(oldConfig, cfg)

	// Create a list of all of the container IDs that need to be restarted
	totalAffectedContainers := map[config.ContainerID]bool{}
	for _, settingList := range changedSettings {
		for _, setting := range settingList {
			for container := range setting.AffectedContainers {
				totalAffectedContainers[container] = true
			}
		}
	}

	// Check if the network has changed
	changeNetworks := false
	if oldConfig.Smartnode.Network.Value != cfg.Smartnode.Network.Value {
		changeNetworks = true
	}

	// Return everything
	return changedSettings, totalAffectedContainers, changeNetworks
}

// Checks to see if the current configuration is valid; if not, returns a list of errors
func (cfg *RocketPoolConfig) Validate() []string {
	errors := []string{}

	// Check for illegal blank strings
	/* TODO - this needs to be smarter and ignore irrelevant settings
	for _, param := range config.GetParameters() {
		if param.Type == ParameterType_String && !param.CanBeBlank && param.Value == "" {
			errors = append(errors, fmt.Sprintf("[%s] cannot be blank.", param.Name))
		}
	}

	for name, subconfig := range config.GetSubconfigs() {
		for _, param := range subconfig.GetParameters() {
			if param.Type == ParameterType_String && !param.CanBeBlank && param.Value == "" {
				errors = append(errors, fmt.Sprintf("[%s - %s] cannot be blank.", name, param.Name))
			}
		}
	}
	*/

	// Force switching of Pocket and Infura
	if cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_Local {
		selectedEc := cfg.ExecutionClient.Value.(config.ExecutionClient)
		switch selectedEc {
		case config.ExecutionClient_Obs_Infura:
			errors = append(errors, "You currently have Infura configured as your primary Execution client, but it is no longer supported because it is not compatible with the upcoming Ethereum Merge. Please go back and choose a full Execution client.")
		case config.ExecutionClient_Obs_Pocket:
			errors = append(errors, "You currently have Pocket configured as your primary Execution client, but it is no longer supported because it is not compatible with the upcoming Ethereum Merge. Please go back and choose a full Execution client.")
		}
	}

	// Force all Docker or all Hybrid
	if cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_Local && cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_External {
		errors = append(errors, "You are using a locally-managed Execution client and an externally-managed Consensus client.\nThis configuration is not compatible with The Merge; please select either locally-managed or externally-managed for both the EC and CC.")
	} else if cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_External && cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_Local {
		errors = append(errors, "You are using an externally-managed Execution client and a locally-managed Consensus client.\nThis configuration is not compatible with The Merge; please select either locally-managed or externally-managed for both the EC and CC.")
	}

	// Ensure there's a MEV-boost URL
	if !cfg.IsNativeMode && cfg.EnableMevBoost.Value == true && cfg.Smartnode.Network.Value != config.Network_Holesky {
		switch cfg.MevBoost.Mode.Value.(config.Mode) {
		case config.Mode_Local:
			// In local MEV-boost mode, the user has to have at least one relay
			relays := cfg.MevBoost.GetEnabledMevRelays()
			if len(relays) == 0 {
				errors = append(errors, "You have MEV-boost enabled in local mode but don't have any profiles or relays enabled. Please select at least one profile or relay to use MEV-boost.")
			}
		case config.Mode_External:
			// In external MEV-boost mode, the user has to have an external URL if they're running Docker mode
			if cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_Local && cfg.MevBoost.ExternalUrl.Value.(string) == "" {
				errors = append(errors, "You have MEV-boost enabled in external mode but don't have a URL set. Please enter the external MEV-boost server URL to use it.")
			}
		default:
			errors = append(errors, "You do not have a MEV-Boost mode configured. You must either select a mode in the `rocketpool service config` UI, or disable MEV-Boost.\nNote that MEV-Boost will be required in a future update, at which point you can no longer disable it.")
		}
	}

	return errors
}

// Applies all of the defaults to all of the settings that have them defined
func (cfg *RocketPoolConfig) applyAllDefaults() error {
	for _, param := range cfg.GetParameters() {
		err := param.SetToDefault(cfg.Smartnode.Network.Value.(config.Network))
		if err != nil {
			return fmt.Errorf("error setting root parameter default: %w", err)
		}
	}

	for name, subconfig := range cfg.GetSubconfigs() {
		for _, param := range subconfig.GetParameters() {
			err := param.SetToDefault(cfg.Smartnode.Network.Value.(config.Network))
			if err != nil {
				return fmt.Errorf("error setting parameter default for %s: %w", name, err)
			}
		}
	}

	return nil
}

// Get all of the changed settings between an old and new config
func getChangedSettingsMap(oldConfig *RocketPoolConfig, newConfig *RocketPoolConfig) map[string][]config.ChangedSetting {
	changedSettings := map[string][]config.ChangedSetting{}

	// Root settings
	oldRootParams := oldConfig.GetParameters()
	newRootParams := newConfig.GetParameters()
	changedSettings[oldConfig.Title] = getChangedSettings(oldRootParams, newRootParams, newConfig)

	// Subconfig settings
	oldSubconfigs := oldConfig.GetSubconfigs()
	for name, subConfig := range newConfig.GetSubconfigs() {
		oldParams := oldSubconfigs[name].GetParameters()
		newParams := subConfig.GetParameters()
		changedSettings[subConfig.GetConfigTitle()] = getChangedSettings(oldParams, newParams, newConfig)
	}

	return changedSettings
}

// Get all of the settings that have changed between the given parameter lists.
// Assumes the parameter lists represent identical parameters (e.g. they have the same number of elements and
// each element has the same ID).
func getChangedSettings(oldParams []*config.Parameter, newParams []*config.Parameter, newConfig *RocketPoolConfig) []config.ChangedSetting {
	changedSettings := []config.ChangedSetting{}

	for i, param := range newParams {
		oldValString := fmt.Sprint(oldParams[i].Value)
		newValString := fmt.Sprint(param.Value)
		if oldValString != newValString {
			changedSettings = append(changedSettings, config.ChangedSetting{
				Name:               param.Name,
				OldValue:           oldValString,
				NewValue:           newValString,
				AffectedContainers: getAffectedContainers(param, newConfig),
			})
		}
	}

	return changedSettings
}

// Handles custom container overrides
func getAffectedContainers(param *config.Parameter, cfg *RocketPoolConfig) map[config.ContainerID]bool {

	affectedContainers := map[config.ContainerID]bool{}

	for _, container := range param.AffectsContainers {
		affectedContainers[container] = true
	}

	// Nimbus doesn't operate in split mode, so all of the VC parameters need to get redirected to the BN instead
	if cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_Local &&
		cfg.ConsensusClient.Value.(config.ConsensusClient) == config.ConsensusClient_Nimbus {
		for _, container := range param.AffectsContainers {
			if container == config.ContainerID_Validator {
				affectedContainers[config.ContainerID_Eth2] = true
			}
		}
	}
	return affectedContainers

}
