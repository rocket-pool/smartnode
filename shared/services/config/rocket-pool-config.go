package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/alessio/shellescape"
	externalip "github.com/glendc/go-external-ip"
	"github.com/pbnjay/memory"
	"github.com/rocket-pool/smartnode/addons"
	"github.com/rocket-pool/smartnode/addons/rescue_node"
	"github.com/rocket-pool/smartnode/shared"
	"github.com/rocket-pool/smartnode/shared/services/config/migration"
	addontypes "github.com/rocket-pool/smartnode/shared/types/addons"
	"github.com/rocket-pool/smartnode/shared/types/config"
	cfgtypes "github.com/rocket-pool/smartnode/shared/types/config"
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
	AlertmanagerContainerName string = "alertmanager"
	ValidatorContainerName    string = "validator"
	WatchtowerContainerName   string = "watchtower"
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
	Reth              *RethConfig              `yaml:"reth,omitempty"`
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
	Alertmanager      *AlertmanagerConfig      `yaml:"alertmanager,omitempty"`
	Exporter          *ExporterConfig          `yaml:"exporter,omitempty"`
	BitflyNodeMetrics *BitflyNodeMetricsConfig `yaml:"bitflyNodeMetrics,omitempty"`

	// Native mode
	Native *NativeConfig `yaml:"native,omitempty"`

	// MEV-Boost
	EnableMevBoost config.Parameter `yaml:"enableMevBoost,omitempty"`
	MevBoost       *MevBoostConfig  `yaml:"mevBoost,omitempty"`

	// Addons
	GraffitiWallWriter addontypes.SmartnodeAddon `yaml:"addon-gww,omitempty"`
	RescueNode         addontypes.SmartnodeAddon `yaml:"addon-rescue-node,omitempty"`
}

// Get the external IP address. Try finding an IPv4 address first to:
// * Improve peer discovery and node performance
// * Avoid unnecessary container restarts caused by switching between IPv4 and IPv6
func getExternalIP() (net.IP, error) {
	// Try IPv4 first
	ip4Consensus := externalip.DefaultConsensus(nil, nil)
	ip4Consensus.UseIPProtocol(4)
	if ip, err := ip4Consensus.ExternalIP(); err == nil {
		return ip, nil
	}

	// Try IPv6 as fallback
	ip6Consensus := externalip.DefaultConsensus(nil, nil)
	ip6Consensus.UseIPProtocol(6)
	return ip6Consensus.ExternalIP()
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
			ID:                 "executionClientMode",
			Name:               "Execution Client Mode",
			Description:        "Choose which mode to use for your Execution client - locally managed (Docker Mode), or externally managed (Hybrid Mode).",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Eth1, config.ContainerID_Eth2, config.ContainerID_Node, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            clientModes,
		},

		ExecutionClient: config.Parameter{
			ID:                 "executionClient",
			Name:               "Execution Client",
			Description:        "Select which Execution client you would like to run.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.ExecutionClient_Geth},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
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
			}, {
				Name:        "Reth (Beta)",
				Description: getAugmentedEcDescription(config.ExecutionClient_Reth, "Reth is a new Ethereum full node implementation that is focused on being user-friendly, highly modular, as well as being fast and efficient. Reth is fully open source and written in Rust.\n\n[orange]NOTE: Reth is currently in beta status. Please consider this when choosing it."),
				Value:       config.ExecutionClient_Reth,
			}},
		},

		UseFallbackClients: config.Parameter{
			ID:                 "useFallbackClients",
			Name:               "Use Fallback Clients",
			Description:        "Enable this if you would like to specify a fallback Execution and Consensus Client, which will temporarily be used by the Smartnode and your Validator Client if your primary Execution / Consensus client pair ever go offline (e.g. if you switch, prune, or resync your clients).",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Validator, config.ContainerID_Node, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ReconnectDelay: config.Parameter{
			ID:                 "reconnectDelay",
			Name:               "Reconnect Delay",
			Description:        "The delay to wait after your primary Execution or Consensus clients fail before trying to reconnect to them. An example format is \"10h20m30s\" - this would make it 10 hours, 20 minutes, and 30 seconds.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: "60s"},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ConsensusClientMode: config.Parameter{
			ID:                 "consensusClientMode",
			Name:               "Consensus Client Mode",
			Description:        "Choose which mode to use for your Consensus client - locally managed (Docker Mode), or externally managed (Hybrid Mode).",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.Mode_Local},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Eth2, config.ContainerID_Node, config.ContainerID_Prometheus, config.ContainerID_Validator, config.ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            clientModes,
		},

		ConsensusClient: config.Parameter{
			ID:                 "consensusClient",
			Name:               "Consensus Client",
			Description:        "Select which Consensus client you would like to use.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.ConsensusClient_Nimbus},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth2, config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
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
			ID:                 "externalConsensusClient",
			Name:               "Consensus Client",
			Description:        "Select which Consensus client your externally managed client is.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: config.ConsensusClient_Lighthouse},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth2, config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
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
			ID:                 "enableMetrics",
			Name:               "Enable Metrics",
			Description:        "Enable the Smartnode's performance and status metrics system. This will provide you with the node operator's Grafana dashboard.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: true},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth2, config.ContainerID_Grafana, config.ContainerID_Prometheus, config.ContainerID_Exporter, config.ContainerID_Alertmanager},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		EnableODaoMetrics: config.Parameter{
			ID:                 "enableODaoMetrics",
			Name:               "Enable Oracle DAO Metrics",
			Description:        "Enable the tracking of Oracle DAO performance metrics, such as prices and balances submission participation.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		EnableBitflyNodeMetrics: config.Parameter{
			ID:                 "enableBitflyNodeMetrics",
			Name:               "Enable Beaconcha.in Node Metrics",
			Description:        "Enable the Beaconcha.in node metrics integration. This will allow you to track your node's metrics from your phone using the Beaconcha.in App.\n\nFor more information on setting up an account and the app, please visit https://beaconcha.in/mobile.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator, config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		EcMetricsPort: config.Parameter{
			ID:                 "ecMetricsPort",
			Name:               "Execution Client Metrics Port",
			Description:        "The port your Execution client should expose its metrics on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultEcMetricsPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1, config.ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		BnMetricsPort: config.Parameter{
			ID:                 "bnMetricsPort",
			Name:               "Beacon Node Metrics Port",
			Description:        "The port your Consensus client's Beacon Node should expose its metrics on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultBnMetricsPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		VcMetricsPort: config.Parameter{
			ID:                 "vcMetricsPort",
			Name:               "Validator Client Metrics Port",
			Description:        "The port your validator client should expose its metrics on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultVcMetricsPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator, config.ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		NodeMetricsPort: config.Parameter{
			ID:                 "nodeMetricsPort",
			Name:               "Node Metrics Port",
			Description:        "The port your Node container should expose its metrics on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultNodeMetricsPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Node, config.ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ExporterMetricsPort: config.Parameter{
			ID:                 "exporterMetricsPort",
			Name:               "Exporter Metrics Port",
			Description:        "The port that Prometheus's Node Exporter should expose its metrics on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultExporterMetricsPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Exporter, config.ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		WatchtowerMetricsPort: config.Parameter{
			ID:                 "watchtowerMetricsPort",
			Name:               "Watchtower Metrics Port",
			Description:        "The port your Watchtower container should expose its metrics on.\nThis is only relevant for Oracle Nodes.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultWatchtowerMetricsPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Watchtower, config.ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		EnableMevBoost: config.Parameter{
			ID:                 "enableMevBoost",
			Name:               "Enable MEV-Boost",
			Description:        "Enable MEV-Boost, which connects your validator to one or more relays of your choice. The relays act as intermediaries between you and professional block builders that find and extract MEV opportunities. The builders will give you a healthy tip in return, which tends to be worth more than blocks you built on your own.\n\n[orange]NOTE: This toggle is temporary during the early Merge days while relays are still being created. It will be removed in the future.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: true},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_MevBoost},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
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
	cfg.Reth = NewRethConfig(cfg)
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
	cfg.Alertmanager = NewAlertmanagerConfig(cfg)
	cfg.Exporter = NewExporterConfig(cfg)
	cfg.BitflyNodeMetrics = NewBitflyNodeMetricsConfig(cfg)
	cfg.Native = NewNativeConfig(cfg)
	cfg.MevBoost = NewMevBoostConfig(cfg)

	// Addons
	cfg.GraffitiWallWriter = addons.NewGraffitiWallWriter()
	cfg.RescueNode = addons.NewRescueNode()

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
		"reth":               cfg.Reth,
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
		"alertmanager":       cfg.Alertmanager,
		"exporter":           cfg.Exporter,
		"bitflyNodeMetrics":  cfg.BitflyNodeMetrics,
		"native":             cfg.Native,
		"mevBoost":           cfg.MevBoost,
		"addons-gww":         cfg.GraffitiWallWriter.GetConfig(),
		"addons-rescue-node": cfg.RescueNode.GetConfig(),
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
		case config.ExecutionClient_Reth:
			return cfg.Reth.EventLogInterval, nil
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
		case config.ConsensusClient_Lighthouse, config.ConsensusClient_Lodestar, config.ConsensusClient_Nimbus, config.ConsensusClient_Prysm, config.ConsensusClient_Teku:
			return cfg.ConsensusCommon.DoppelgangerDetection.Value.(bool), nil
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
			return cfg.ExternalTeku.DoppelgangerDetection.Value.(bool), nil
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

// Gets the hostname portion of the Execution Client's URI.
// Used by text/template to format prometheus.yml.
func (cfg *RocketPoolConfig) GetExecutionHostname() (string, error) {
	if cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_Local {
		return Eth1ContainerName, nil
	}
	ecUrl, err := url.Parse(cfg.ExternalExecution.HttpUrl.Value.(string))
	if err != nil {
		return "", fmt.Errorf("Invalid External Execution URL %s: %w", cfg.ExternalExecution.HttpUrl.Value.(string), err)
	}

	return ecUrl.Hostname(), nil
}

// Gets the hostname portion of the Consensus Client's URI.
// Used by text/template to format prometheus.yml.
func (cfg *RocketPoolConfig) GetConsensusHostname() (string, error) {
	if cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_Local {
		return Eth2ContainerName, nil
	}

	var rawUrl string

	consensusClient := cfg.ExternalConsensusClient.Value.(config.ConsensusClient)

	switch consensusClient {
	case config.ConsensusClient_Lighthouse:
		rawUrl = cfg.ExternalLighthouse.HttpUrl.Value.(string)
	case config.ConsensusClient_Lodestar:
		rawUrl = cfg.ExternalLodestar.HttpUrl.Value.(string)
	case config.ConsensusClient_Nimbus:
		rawUrl = cfg.ExternalNimbus.HttpUrl.Value.(string)
	case config.ConsensusClient_Prysm:
		rawUrl = cfg.ExternalPrysm.HttpUrl.Value.(string)
	case config.ConsensusClient_Teku:
		rawUrl = cfg.ExternalTeku.HttpUrl.Value.(string)
	}
	ccUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", fmt.Errorf("Invalid External Consensus URL %s: %w", rawUrl, err)
	}

	return ccUrl.Hostname(), nil
}

// Gets the tag of the vc container
// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) GetVCContainerTag() (string, error) {
	cCfg, err := cfg.GetSelectedConsensusClientConfig()
	if err != nil {
		return "", err
	}

	return cCfg.GetValidatorImage(), nil
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) ExecutionClientLocal() bool {
	return cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_Local
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) ConsensusClientLocal() bool {
	return cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_Local
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) ConsensusClientApiUrl() (string, error) {
	// Check if Rescue Node is in-use
	cc, _ := cfg.GetSelectedConsensusClient()
	overrides, err := cfg.RescueNode.(*rescue_node.RescueNode).GetOverrides(cc)
	if err != nil {
		return "", fmt.Errorf("error using Rescue Node: %w", err)
	}
	if overrides != nil {
		// Use the rescue node
		return overrides.CcApiEndpoint, nil
	}

	if cfg.ConsensusClientLocal() {
		// Use the eth2 container
		return fmt.Sprintf("http://%s:%d", Eth2ContainerName, cfg.ConsensusCommon.ApiPort.Value), nil
	}

	cCfg, err := cfg.GetSelectedConsensusClientConfig()
	if err != nil {
		return "", err
	}

	// Use the external eth2 client
	return cCfg.(config.ExternalConsensusConfig).GetApiUrl(), nil
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) ConsensusClientRpcUrl() (string, error) {
	// Check if Rescue Node is in-use
	cc, _ := cfg.GetSelectedConsensusClient()
	if cc != config.ConsensusClient_Prysm {
		return "", nil
	}

	overrides, err := cfg.RescueNode.(*rescue_node.RescueNode).GetOverrides(cc)
	if err != nil {
		return "", fmt.Errorf("error using Rescue Node: %w", err)
	}
	if overrides != nil {
		// Use the rescue node
		return overrides.CcRpcEndpoint, nil
	}

	if cfg.ConsensusClientLocal() {
		// Use the eth2 container
		return fmt.Sprintf("%s:%d", Eth2ContainerName, cfg.Prysm.RpcPort.Value), nil
	}

	// Use the external RPC endpoint
	return cfg.ExternalPrysm.JsonRpcUrl.Value.(string), nil
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) FallbackCcApiUrl() string {
	if !cfg.UseFallbackClients.Value.(bool) {
		return ""
	}

	cc, _ := cfg.GetSelectedConsensusClient()
	if cc == config.ConsensusClient_Prysm {
		return cfg.FallbackPrysm.CcHttpUrl.Value.(string)
	}

	return cfg.FallbackNormal.CcHttpUrl.Value.(string)
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) FallbackCcRpcUrl() string {
	if !cfg.UseFallbackClients.Value.(bool) {
		return ""
	}

	cc, _ := cfg.GetSelectedConsensusClient()
	if cc != config.ConsensusClient_Prysm {
		return ""
	}

	return cfg.FallbackPrysm.JsonRpcUrl.Value.(string)
}

// Used by text/template to format validator.yml
// Only returns the user-entered value, not the prefixed value
func (cfg *RocketPoolConfig) CustomGraffiti() (string, error) {
	if cfg.ConsensusClientLocal() {
		return cfg.ConsensusCommon.Graffiti.Value.(string), nil
	}

	cc, _ := cfg.GetSelectedConsensusClient()
	switch cc {
	case config.ConsensusClient_Lighthouse:
		return cfg.ExternalLighthouse.Graffiti.Value.(string), nil
	case config.ConsensusClient_Lodestar:
		return cfg.ExternalLodestar.Graffiti.Value.(string), nil
	case config.ConsensusClient_Nimbus:
		return cfg.ExternalNimbus.Graffiti.Value.(string), nil
	case config.ConsensusClient_Prysm:
		return cfg.ExternalPrysm.Graffiti.Value.(string), nil
	case config.ConsensusClient_Teku:
		return cfg.ExternalTeku.Graffiti.Value.(string), nil
	default:
	}
	return "", fmt.Errorf("unknown external consensus client [%v] selected", cc)
}

// Used by text/template to format validator.yml
// Only returns the the prefix
func (cfg *RocketPoolConfig) GraffitiPrefix() string {
	// Graffiti
	identifier := ""
	versionString := fmt.Sprintf("v%s", shared.RocketPoolVersion)
	if len(versionString) < 8 {
		var ecInitial string
		if !cfg.ExecutionClientLocal() {
			ecInitial = "X"
		} else {
			ecInitial = strings.ToUpper(string(cfg.ExecutionClient.Value.(config.ExecutionClient))[:1])
		}

		var ccInitial string
		consensusClient, _ := cfg.GetSelectedConsensusClient()
		switch consensusClient {
		case config.ConsensusClient_Lodestar:
			ccInitial = "S" // Lodestar is special because it conflicts with Lighthouse
		default:
			ccInitial = strings.ToUpper(string(consensusClient)[:1])
		}
		identifier = fmt.Sprintf("-%s%s", ecInitial, ccInitial)
	}

	return fmt.Sprintf("RP%s %s", identifier, versionString)
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) Graffiti() (string, error) {
	prefix := cfg.GraffitiPrefix()
	customGraffiti, err := cfg.CustomGraffiti()
	if err != nil {
		return "", err
	}
	if customGraffiti == "" {
		return prefix, nil
	}
	return fmt.Sprintf("%s (%s)", prefix, customGraffiti), nil
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) RocketPoolVersion() string {
	return shared.RocketPoolVersion
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) VcAdditionalFlags() (string, error) {
	// Check if Rescue Node is in-use
	cc, mode := cfg.GetSelectedConsensusClient()

	overrides, err := cfg.RescueNode.(*rescue_node.RescueNode).GetOverrides(cc)
	if err != nil {
		return "", fmt.Errorf("error using Rescue Node: %w", err)
	}

	var addtlFlags string
	switch mode {
	case config.Mode_Local:
		client := cfg.ConsensusClient.Value.(config.ConsensusClient)
		switch client {
		case config.ConsensusClient_Lighthouse:
			addtlFlags = cfg.Lighthouse.AdditionalVcFlags.Value.(string)
		case config.ConsensusClient_Lodestar:
			addtlFlags = cfg.Lodestar.AdditionalVcFlags.Value.(string)
		case config.ConsensusClient_Nimbus:
			addtlFlags = cfg.Nimbus.AdditionalVcFlags.Value.(string)
		case config.ConsensusClient_Prysm:
			addtlFlags = cfg.Prysm.AdditionalVcFlags.Value.(string)
		case config.ConsensusClient_Teku:
			addtlFlags = cfg.Teku.AdditionalVcFlags.Value.(string)
		default:
			return "", fmt.Errorf("unknown consensus client [%v] selected", client)
		}

	case config.Mode_External:
		client := cfg.ExternalConsensusClient.Value.(config.ConsensusClient)
		switch client {
		case config.ConsensusClient_Lighthouse:
			addtlFlags = cfg.ExternalLighthouse.AdditionalVcFlags.Value.(string)
		case config.ConsensusClient_Lodestar:
			addtlFlags = cfg.ExternalLodestar.AdditionalVcFlags.Value.(string)
		case config.ConsensusClient_Nimbus:
			addtlFlags = cfg.ExternalNimbus.AdditionalVcFlags.Value.(string)
		case config.ConsensusClient_Prysm:
			addtlFlags = cfg.ExternalPrysm.AdditionalVcFlags.Value.(string)
		case config.ConsensusClient_Teku:
			addtlFlags = cfg.ExternalTeku.AdditionalVcFlags.Value.(string)
		default:
			return "", fmt.Errorf("unknown external consensus client [%v] selected", client)
		}

	default:
		return "", fmt.Errorf("unknown consensus client mode [%v]", mode)
	}

	first := true
	out := ""
	if addtlFlags != "" {
		first = false
		out = addtlFlags
	}
	if overrides != nil && overrides.VcAdditionalFlags != "" {
		if !first {
			out = out + " "
		}
		out = out + overrides.VcAdditionalFlags
	}
	return out, nil
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) FeeRecipientFile() string {
	return FeeRecipientFilename
}

// Used by text/template to format validator.yml
func (cfg *RocketPoolConfig) MevBoostUrl() string {
	if !cfg.EnableMevBoost.Value.(bool) {
		return ""
	}

	if cfg.MevBoost.Mode.Value == config.Mode_Local {
		return fmt.Sprintf("http://%s:%d", MevBoostContainerName, cfg.MevBoost.Port.Value)
	}

	return cfg.MevBoost.ExternalUrl.Value.(string)
}

// Gets the tag of the ec container
// Used by text/template to format eth1.yml
func (cfg *RocketPoolConfig) GetECContainerTag() (string, error) {
	if !cfg.ExecutionClientLocal() {
		return "", fmt.Errorf("Execution client is external, there is no container tag")
	}

	switch cfg.ExecutionClient.Value.(config.ExecutionClient) {
	case config.ExecutionClient_Geth:
		return cfg.Geth.ContainerTag.Value.(string), nil
	case config.ExecutionClient_Nethermind:
		return cfg.Nethermind.ContainerTag.Value.(string), nil
	case config.ExecutionClient_Besu:
		return cfg.Besu.ContainerTag.Value.(string), nil
	case config.ExecutionClient_Reth:
		return cfg.Reth.ContainerTag.Value.(string), nil
	}

	return "", fmt.Errorf("Unknown Execution Client %s", string(cfg.ExecutionClient.Value.(config.ExecutionClient)))
}

// Gets the stop signal of the ec container
// Used by text/template to format eth1.yml
func (cfg *RocketPoolConfig) GetECStopSignal() (string, error) {
	if !cfg.ExecutionClientLocal() {
		return "", fmt.Errorf("Execution client is external, there is no stop signal")
	}

	switch cfg.ExecutionClient.Value.(config.ExecutionClient) {
	case config.ExecutionClient_Geth:
		return gethStopSignal, nil
	case config.ExecutionClient_Nethermind:
		return nethermindStopSignal, nil
	case config.ExecutionClient_Besu:
		return besuStopSignal, nil
	case config.ExecutionClient_Reth:
		return rethStopSignal, nil
	}

	return "", fmt.Errorf("Unknown Execution Client %s", string(cfg.ExecutionClient.Value.(config.ExecutionClient)))
}

// Gets the stop signal of the ec container
// Used by text/template to format eth1.yml
func (cfg *RocketPoolConfig) GetECOpenAPIPorts() string {
	rpcMode := cfg.ExecutionCommon.OpenRpcPorts.Value.(config.RPCMode)
	if !rpcMode.Open() {
		return ""
	}
	httpMapping := rpcMode.DockerPortMapping(cfg.ExecutionCommon.HttpPort.Value.(uint16))
	wsMapping := rpcMode.DockerPortMapping(cfg.ExecutionCommon.WsPort.Value.(uint16))
	return fmt.Sprintf(", \"%s\", \"%s\"", httpMapping, wsMapping)
}

// Gets the max peers of the ec container
// Used by text/template to format eth1.yml
func (cfg *RocketPoolConfig) GetECMaxPeers() (uint16, error) {
	if !cfg.ExecutionClientLocal() {
		return 0, fmt.Errorf("Execution client is external, there is no max peers")
	}

	switch cfg.ExecutionClient.Value.(config.ExecutionClient) {
	case config.ExecutionClient_Geth:
		return cfg.Geth.MaxPeers.Value.(uint16), nil
	case config.ExecutionClient_Nethermind:
		return cfg.Nethermind.MaxPeers.Value.(uint16), nil
	case config.ExecutionClient_Besu:
		return cfg.Besu.MaxPeers.Value.(uint16), nil
	case config.ExecutionClient_Reth:
		return cfg.Reth.MaxPeers.Value.(uint16), nil
	}

	return 0, fmt.Errorf("Unknown Execution Client %s", string(cfg.ExecutionClient.Value.(config.ExecutionClient)))
}

// Used by text/template to format eth1.yml
func (cfg *RocketPoolConfig) GetECAdditionalFlags() (string, error) {
	if !cfg.ExecutionClientLocal() {
		return "", fmt.Errorf("Execution client is external, there are no additional flags")
	}

	switch cfg.ExecutionClient.Value.(config.ExecutionClient) {
	case config.ExecutionClient_Geth:
		return cfg.Geth.AdditionalFlags.Value.(string), nil
	case config.ExecutionClient_Nethermind:
		return cfg.Nethermind.AdditionalFlags.Value.(string), nil
	case config.ExecutionClient_Besu:
		return cfg.Besu.AdditionalFlags.Value.(string), nil
	case config.ExecutionClient_Reth:
		return cfg.Reth.AdditionalFlags.Value.(string), nil
	}

	return "", fmt.Errorf("Unknown Execution Client %s", string(cfg.ExecutionClient.Value.(config.ExecutionClient)))
}

// Used by text/template to format eth1.yml
func (cfg *RocketPoolConfig) GetExternalIp() string {
	// Get the external IP address
	ip, err := getExternalIP()
	if err != nil {
		fmt.Println("Warning: couldn't get external IP address; if you're using Nimbus or Besu, it may have trouble finding peers:")
		fmt.Println(err.Error())
		return ""
	}

	if ip.To4() == nil {
		fmt.Println("Warning: external IP address is v6; if you're using Nimbus or Besu, it may have trouble finding peers:")
	}

	return ip.String()
}

// Gets the tag of the cc container
// Used by text/template to format eth2.yml
func (cfg *RocketPoolConfig) GetBeaconContainerTag() (string, error) {
	cCfg, err := cfg.GetSelectedConsensusClientConfig()
	if err != nil {
		return "", err
	}

	return cCfg.GetBeaconNodeImage(), nil
}

// Used by text/template to format eth2.yml
func (cfg *RocketPoolConfig) GetBnOpenPorts() []string {
	// Handle open API ports
	bnOpenPorts := make([]string, 0)
	consensusClient := cfg.ConsensusClient.Value.(config.ConsensusClient)
	apiPortMode := cfg.ConsensusCommon.OpenApiPort.Value.(config.RPCMode)
	if apiPortMode.Open() {
		apiPort := cfg.ConsensusCommon.ApiPort.Value.(uint16)
		bnOpenPorts = append(bnOpenPorts, apiPortMode.DockerPortMapping(apiPort))
	}
	if consensusClient == config.ConsensusClient_Prysm {
		prysmRpcPortMode := cfg.Prysm.OpenRpcPort.Value.(config.RPCMode)
		if prysmRpcPortMode.Open() {
			prysmRpcPort := cfg.Prysm.RpcPort.Value.(uint16)
			bnOpenPorts = append(bnOpenPorts, prysmRpcPortMode.DockerPortMapping(prysmRpcPort))
		}
	}
	return bnOpenPorts
}

// Used by text/template to format eth2.yml
func (cfg *RocketPoolConfig) GetEcHttpEndpoint() string {
	if cfg.ExecutionClientLocal() {
		return fmt.Sprintf("http://%s:%d", Eth1ContainerName, cfg.ExecutionCommon.HttpPort.Value)
	}

	return cfg.ExternalExecution.HttpUrl.Value.(string)
}

// Used by text/template to format eth2.yml
func (cfg *RocketPoolConfig) GetEcWsEndpoint() string {
	if cfg.ExecutionClientLocal() {
		return fmt.Sprintf("ws://%s:%d", Eth1ContainerName, cfg.ExecutionCommon.WsPort.Value)
	}

	return cfg.ExternalExecution.WsUrl.Value.(string)
}

// Gets the max peers of the bn container
// Used by text/template to format eth2.yml
func (cfg *RocketPoolConfig) GetBNMaxPeers() (uint16, error) {
	if !cfg.ConsensusClientLocal() {
		return 0, fmt.Errorf("Consensus client is external, there is no max peers")
	}

	switch cfg.ConsensusClient.Value.(config.ConsensusClient) {
	case config.ConsensusClient_Lighthouse:
		return cfg.Lighthouse.MaxPeers.Value.(uint16), nil
	case config.ConsensusClient_Lodestar:
		return cfg.Lodestar.MaxPeers.Value.(uint16), nil
	case config.ConsensusClient_Nimbus:
		return cfg.Nimbus.MaxPeers.Value.(uint16), nil
	case config.ConsensusClient_Teku:
		return cfg.Teku.MaxPeers.Value.(uint16), nil
	case config.ConsensusClient_Prysm:
		return cfg.Prysm.MaxPeers.Value.(uint16), nil
	}

	return 0, fmt.Errorf("Unknown Consensus Client %s", string(cfg.ConsensusClient.Value.(config.ConsensusClient)))
}

// Used by text/template to format eth2.yml
func (cfg *RocketPoolConfig) GetBNAdditionalFlags() (string, error) {
	if !cfg.ConsensusClientLocal() {
		return "", fmt.Errorf("Consensus client is external, there are no additional flags")
	}

	switch cfg.ConsensusClient.Value.(config.ConsensusClient) {
	case config.ConsensusClient_Lighthouse:
		return cfg.Lighthouse.AdditionalBnFlags.Value.(string), nil
	case config.ConsensusClient_Lodestar:
		return cfg.Lodestar.AdditionalBnFlags.Value.(string), nil
	case config.ConsensusClient_Nimbus:
		return cfg.Nimbus.AdditionalBnFlags.Value.(string), nil
	case config.ConsensusClient_Teku:
		return cfg.Teku.AdditionalBnFlags.Value.(string), nil
	case config.ConsensusClient_Prysm:
		return cfg.Prysm.AdditionalBnFlags.Value.(string), nil
	}

	return "", fmt.Errorf("Unknown Consensus Client %s", string(cfg.ConsensusClient.Value.(config.ConsensusClient)))
}

// Used by text/template to format exporter.yml
func (cfg *RocketPoolConfig) GetExporterAdditionalFlags() []string {
	flags := strings.Trim(cfg.Exporter.AdditionalFlags.Value.(string), " ")
	if flags == "" {
		return nil
	}
	return strings.Split(flags, " ")
}

// Used by text/template to format prometheus.yml
func (cfg *RocketPoolConfig) GetPrometheusAdditionalFlags() []string {
	flags := strings.Trim(cfg.Prometheus.AdditionalFlags.Value.(string), " ")
	if flags == "" {
		return nil
	}
	return strings.Split(flags, " ")
}

// Used by text/template to format prometheus.yml
func (cfg *RocketPoolConfig) GetPrometheusOpenPorts() string {
	portMode := cfg.Prometheus.OpenPort.Value.(config.RPCMode)
	if !portMode.Open() {
		return ""
	}
	return fmt.Sprintf("\"%s\"", portMode.DockerPortMapping(cfg.Prometheus.Port.Value.(uint16)))
}

// Used by text/template to format mev-boost.yml
func (cfg *RocketPoolConfig) GetMevBoostOpenPorts() string {
	portMode := cfg.MevBoost.OpenRpcPort.Value.(config.RPCMode)
	if !portMode.Open() {
		return ""
	}
	port := cfg.MevBoost.Port.Value.(uint16)
	return fmt.Sprintf("\"%s\"", portMode.DockerPortMapping(port))
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

	// Force all Docker or all Hybrid
	if cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_Local && cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_External {
		errors = append(errors, "You are using a locally-managed Execution client and an externally-managed Consensus client.\nThis configuration is not compatible with The Merge; please select either locally-managed or externally-managed for both the EC and CC.")
	} else if cfg.ExecutionClientMode.Value.(config.Mode) == config.Mode_External && cfg.ConsensusClientMode.Value.(config.Mode) == config.Mode_Local {
		errors = append(errors, "You are using an externally-managed Execution client and a locally-managed Consensus client.\nThis configuration is not compatible with The Merge; please select either locally-managed or externally-managed for both the EC and CC.")
	}

	if cfg.ExecutionClient.Value.(cfgtypes.ExecutionClient) == cfgtypes.ExecutionClient_Reth && cfg.Smartnode.Network.Value.(cfgtypes.Network) == cfgtypes.Network_Mainnet {
		errors = append(errors, "The Reth client is currently an alpha release and not to be used on Mainnet")
	}

	// Ensure there's a MEV-boost URL
	if cfg.Smartnode.Network.Value == config.Network_Holesky || cfg.Smartnode.Network.Value == config.Network_Devnet {
		// Disabled on Holesky
		cfg.EnableMevBoost.Value = false
	}
	if !cfg.IsNativeMode && cfg.EnableMevBoost.Value == true {
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

	// Technically not required since native mode doesn't support addons, but defensively check to make sure a native mode
	// user hasn't tried to configure the rescue node via the TUI
	if cfg.RescueNode.GetEnabledParameter().Value.(bool) {
		if cfg.IsNativeMode {
			errors = append(errors, "Rescue Node add-on is incompatible with native mode.\nYou can still connect manually, visit the rescue node website for more information.")
		}

		params := cfg.RescueNode.GetConfig().GetParameters()
		for _, param := range params {
			if param.Type != config.ParameterType_String {
				continue
			}

			if param.Value.(string) == "" {
				errors = append(errors, "Rescue Node requires both a username and a password.")
				break
			}
		}
	}

	// Ensure the selected port numbers are unique. Keeps track of all the errors
	portMap := make(map[interface{}]bool)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.ConsensusCommon.ApiPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.ConsensusCommon.P2pPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.ExecutionCommon.EnginePort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.ExecutionCommon.WsPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.ExecutionCommon.P2pPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.ExecutionCommon.HttpPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.BnMetricsPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.EcMetricsPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.ExporterMetricsPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.NodeMetricsPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.VcMetricsPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.WatchtowerMetricsPort, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.Grafana.Port, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.MevBoost.Port, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.Prometheus.Port, errors)
	portMap, errors = addAndCheckForDuplicate(portMap, cfg.Alertmanager.Port, errors)
	_, errors = addAndCheckForDuplicate(portMap, cfg.Lighthouse.P2pQuicPort, errors)

	return errors
}

func addAndCheckForDuplicate(portMap map[interface{}]bool, param config.Parameter, errors []string) (map[interface{}]bool, []string) {
	port := fmt.Sprintf("%v", param.Value)
	if port == "" {
		return portMap, errors
	}
	if portMap[port] {
		return portMap, append(errors, fmt.Sprintf("Port %s for %s is already in use", port, param.Name))
	} else {
		portMap[port] = true
	}
	return portMap, errors

}

func (cfg *RocketPoolConfig) GetNetwork() config.Network {
	return cfg.Smartnode.Network.Value.(config.Network)
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
