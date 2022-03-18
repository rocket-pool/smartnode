package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alessio/shellescape"
	"github.com/rocket-pool/smartnode/shared"
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
	NodeContainerName         string = "node"
	PrometheusContainerName   string = "prometheus"
	ValidatorContainerName    string = "validator"
	WatchtowerContainerName   string = "watchtower"
)

// Defaults
const defaultBnMetricsPort uint16 = 9100
const defaultVcMetricsPort uint16 = 9101
const defaultNodeMetricsPort uint16 = 9102
const defaultExporterMetricsPort uint16 = 9103
const defaultWatchtowerMetricsPort uint16 = 9104

// The master configuration struct
type RocketPoolConfig struct {
	Title string `yaml:"title,omitempty"`

	Version string `yaml:"version,omitempty"`

	RocketPoolDirectory string `yaml:"rocketPoolDirectory,omitempty"`

	IsNativeMode bool `yaml:"isNativeMode,omitempty"`

	// Execution client settings
	ExecutionClientMode Parameter `yaml:"executionClientMode"`
	ExecutionClient     Parameter `yaml:"executionClient"`

	// Fallback execution client settings
	UseFallbackExecutionClient  Parameter `yaml:"useFallbackExecutionClient,omitempty"`
	FallbackExecutionClientMode Parameter `yaml:"fallbackExecutionClientMode,omitempty"`
	FallbackExecutionClient     Parameter `yaml:"fallbackExecutionClient,omitempty"`
	ReconnectDelay              Parameter `yaml:"reconnectDelay,omitempty"`

	// Consensus client settings
	ConsensusClientMode     Parameter `yaml:"consensusClientMode,omitempty"`
	ConsensusClient         Parameter `yaml:"consensusClient,omitempty"`
	ExternalConsensusClient Parameter `yaml:"externalConsensusClient,omitempty"`

	// Metrics settings
	EnableMetrics         Parameter `yaml:"enableMetrics,omitempty"`
	BnMetricsPort         Parameter `yaml:"bnMetricsPort,omitempty"`
	VcMetricsPort         Parameter `yaml:"vcMetricsPort,omitempty"`
	NodeMetricsPort       Parameter `yaml:"nodeMetricsPort,omitempty"`
	ExporterMetricsPort   Parameter `yaml:"exporterMetricsPort,omitempty"`
	WatchtowerMetricsPort Parameter `yaml:"watchtowerMetricsPort,omitempty"`

	// The Smartnode configuration
	Smartnode *SmartnodeConfig `yaml:"smartnode"`

	// Execution client configurations
	ExecutionCommon   *ExecutionCommonConfig   `yaml:"executionCommon,omitempty"`
	Geth              *GethConfig              `yaml:"geth,omitempty"`
	Infura            *InfuraConfig            `yaml:"infura,omitempty"`
	Pocket            *PocketConfig            `yaml:"pocket,omitempty"`
	ExternalExecution *ExternalExecutionConfig `yaml:"externalExecution,omitempty"`

	// Fallback Execution client configurations
	FallbackExecutionCommon   *ExecutionCommonConfig   `yaml:"fallbackExecutionCommon,omitempty"`
	FallbackInfura            *InfuraConfig            `yaml:"fallbackInfura,omitempty"`
	FallbackPocket            *PocketConfig            `yaml:"fallbackPocket,omitempty"`
	FallbackExternalExecution *ExternalExecutionConfig `yaml:"fallbackExternalExecution,omitempty"`

	// Consensus client configurations
	ConsensusCommon    *ConsensusCommonConfig    `yaml:"consensusCommon,omitempty"`
	Lighthouse         *LighthouseConfig         `yaml:"lighthouse,omitempty"`
	Nimbus             *NimbusConfig             `yaml:"nimbus,omitempty"`
	Prysm              *PrysmConfig              `yaml:"prysm,omitempty"`
	Teku               *TekuConfig               `yaml:"teku,omitempty"`
	ExternalLighthouse *ExternalLighthouseConfig `yaml:"externalLighthouse,omitempty"`
	ExternalPrysm      *ExternalPrysmConfig      `yaml:"externalPrysm,omitempty"`
	ExternalTeku       *ExternalTekuConfig       `yaml:"externalTeku,omitempty"`

	// Metrics
	Grafana    *GrafanaConfig    `yaml:"grafana,omitempty"`
	Prometheus *PrometheusConfig `yaml:"prometheus,omitempty"`
	Exporter   *ExporterConfig   `yaml:"exporter,omitempty"`

	// Native mode
	Native *NativeConfig `yaml:"native,omitempty"`
}

// Load configuration settings from a file
func LoadFromFile(path string) (*RocketPoolConfig, error) {

	// Return nil if the file doesn't exist
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	// Read the file
	configBytes, err := ioutil.ReadFile(path)
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

	config := &RocketPoolConfig{
		Title:               "Top-level Settings",
		RocketPoolDirectory: rpDir,
		IsNativeMode:        isNativeMode,

		ExecutionClientMode: Parameter{
			ID:                   "executionClientMode",
			Name:                 "Execution Client Mode",
			Description:          "Choose which mode to use for your Execution client - locally managed (Docker Mode), or externally managed (Hybrid Mode).",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				Name:        "Locally Managed",
				Description: "Allow the Smartnode to manage an Execution client for you (Docker Mode)",
				Value:       Mode_Local,
			}, {
				Name:        "Externally Managed",
				Description: "Use an existing Execution client that you manage on your own (Hybrid Mode)",
				Value:       Mode_External,
			}},
		},

		ExecutionClient: Parameter{
			ID:                   "executionClient",
			Name:                 "Execution Client",
			Description:          "Select which Execution client you would like to run.",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: ExecutionClient_Geth},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				Name:        "Geth",
				Description: "Geth is one of the three original implementations of the Ethereum protocol. It is written in Go, fully open source and licensed under the GNU LGPL v3.",
				Value:       ExecutionClient_Geth,
			}, {
				Name:        "Infura",
				Description: "Use infura.io as a light client for Eth 1.0. Not recommended for use in production.",
				Value:       ExecutionClient_Infura,
			}, {
				Name:        "Pocket",
				Description: "Use Pocket Network as a decentralized light client for Eth 1.0. Suitable for use in production.",
				Value:       ExecutionClient_Pocket,
			}},
		},

		UseFallbackExecutionClient: Parameter{
			ID:                   "useFallbackExecutionClient",
			Name:                 "Use Fallback Execution Client",
			Description:          "Enable this if you would like to specify a fallback Execution client, which will temporarily be used by the Smartnode and your Consensus client if your primary Execution client ever goes offline.",
			Type:                 ParameterType_Bool,
			Default:              map[Network]interface{}{Network_All: false},
			AffectsContainers:    []ContainerID{ContainerID_Eth1Fallback},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		FallbackExecutionClientMode: Parameter{
			ID:                   "fallbackExecutionClientMode",
			Name:                 "Fallback Execution Client Mode",
			Description:          "Choose which mode to use for your fallback Execution client - locally managed (Docker Mode), or externally managed (Hybrid Mode).",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: nil},
			AffectsContainers:    []ContainerID{ContainerID_Eth1Fallback},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				Name:        "Locally Managed",
				Description: "Allow the Smartnode to manage a fallback Execution client for you (Docker Mode)",
				Value:       Mode_Local,
			}, {
				Name:        "Externally Managed",
				Description: "Use an existing fallback Execution client that you manage on your own (Hybrid Mode)",
				Value:       Mode_External,
			}},
		},

		FallbackExecutionClient: Parameter{
			ID:                   "fallbackExecutionClient",
			Name:                 "Fallback Execution Client",
			Description:          "Select which fallback Execution client you would like to run.",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: ExecutionClient_Pocket},
			AffectsContainers:    []ContainerID{ContainerID_Eth1Fallback},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				Name:        "Infura",
				Description: "Use infura.io as a light client for Eth 1.0. Not recommended for use in production.",
				Value:       ExecutionClient_Infura,
			}, {
				Name:        "Pocket",
				Description: "Use Pocket Network as a decentralized light client for Eth 1.0. Suitable for use in production.",
				Value:       ExecutionClient_Pocket,
			}},
		},

		ReconnectDelay: Parameter{
			ID:                   "reconnectDelay",
			Name:                 "Reconnect Delay",
			Description:          "The delay to wait after the primary Execution client fails before trying to reconnect to it. The format is \"10h20m30s\".",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: "60s"},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ConsensusClientMode: Parameter{
			ID:                   "consensusClientMode",
			Name:                 "Consensus Client Mode",
			Description:          "Choose which mode to use for your Consensus client - locally managed (Docker Mode), or externally managed (Hybrid Mode).",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: Mode_Local},
			AffectsContainers:    []ContainerID{ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				Name:        "Locally Managed",
				Description: "Allow the Smartnode to manage a Consensus client for you (Docker Mode)",
				Value:       Mode_Local,
			}, {
				Name:        "Externally Managed",
				Description: "Use an existing Consensus client that you manage on your own (Hybrid Mode)",
				Value:       Mode_External,
			}},
		},

		ConsensusClient: Parameter{
			ID:                   "consensusClient",
			Name:                 "Consensus Client",
			Description:          "Select which Consensus client you would like to use.",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: ConsensusClient_Nimbus},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				Name:        "Lighthouse",
				Description: "Lighthouse is a Consensus client with a heavy focus on speed and security. The team behind it, Sigma Prime, is an information security and software engineering firm who have funded Lighthouse along with the Ethereum Foundation, Consensys, and private individuals. Lighthouse is built in Rust and offered under an Apache 2.0 License.",
				Value:       ConsensusClient_Lighthouse,
			}, {
				Name:        "Nimbus",
				Description: "Nimbus is a Consensus client implementation that strives to be as lightweight as possible in terms of resources used. This allows it to perform well on embedded systems, resource-restricted devices -- including Raspberry Pis and mobile devices -- and multi-purpose servers.",
				Value:       ConsensusClient_Nimbus,
			}, {
				Name:        "Prysm",
				Description: "Prysm is a Go implementation of Ethereum Consensus protocol with a focus on usability, security, and reliability. Prysm is developed by Prysmatic Labs, a company with the sole focus on the development of their client. Prysm is written in Go and released under a GPL-3.0 license.",
				Value:       ConsensusClient_Prysm,
			}, {
				Name:        "Teku",
				Description: "PegaSys Teku (formerly known as Artemis) is a Java-based Ethereum 2.0 client designed & built to meet institutional needs and security requirements. PegaSys is an arm of ConsenSys dedicated to building enterprise-ready clients and tools for interacting with the core Ethereum platform. Teku is Apache 2 licensed and written in Java, a language notable for its maturity & ubiquity.",
				Value:       ConsensusClient_Teku,
			}},
		},

		ExternalConsensusClient: Parameter{
			ID:                   "externalConsensusClient",
			Name:                 "Consensus Client",
			Description:          "Select which Consensus client your externally managed client is.",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: ConsensusClient_Lighthouse},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				Name:        "Lighthouse",
				Description: "Select this if you will use Lighthouse as your Consensus client.",
				Value:       ConsensusClient_Lighthouse,
			}, {
				Name:        "Prysm",
				Description: "Select this if you will use Prysm as your Consensus client.",
				Value:       ConsensusClient_Prysm,
			}, {
				Name:        "Teku",
				Description: "Select this if you will use Teku as your Consensus client.",
				Value:       ConsensusClient_Teku,
			}},
		},

		EnableMetrics: Parameter{
			ID:                   "enableMetrics",
			Name:                 "Enable Metrics",
			Description:          "Enable the Smartnode's performance and status metrics system. This will provide you with the node operator's Grafana dashboard.",
			Type:                 ParameterType_Bool,
			Default:              map[Network]interface{}{Network_All: false},
			AffectsContainers:    []ContainerID{ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Grafana, ContainerID_Prometheus, ContainerID_Exporter},
			EnvironmentVariables: []string{"ENABLE_METRICS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		BnMetricsPort: Parameter{
			ID:                   "bnMetricsPort",
			Name:                 "Beacon Node Metrics Port",
			Description:          "The port your Consensus client's Beacon Node should expose its metrics on.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultBnMetricsPort},
			AffectsContainers:    []ContainerID{ContainerID_Eth2, ContainerID_Prometheus},
			EnvironmentVariables: []string{"BN_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		VcMetricsPort: Parameter{
			ID:                   "vcMetricsPort",
			Name:                 "Validator Client Metrics Port",
			Description:          "The port your validator client should expose its metrics on.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultVcMetricsPort},
			AffectsContainers:    []ContainerID{ContainerID_Validator, ContainerID_Prometheus},
			EnvironmentVariables: []string{"VC_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		NodeMetricsPort: Parameter{
			ID:                   "nodeMetricsPort",
			Name:                 "Node Metrics Port",
			Description:          "The port your Node container should expose its metrics on.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultNodeMetricsPort},
			AffectsContainers:    []ContainerID{ContainerID_Node, ContainerID_Prometheus},
			EnvironmentVariables: []string{"NODE_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ExporterMetricsPort: Parameter{
			ID:                   "exporterMetricsPort",
			Name:                 "Exporter Metrics Port",
			Description:          "The port that Prometheus's Node Exporter should expose its metrics on.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultExporterMetricsPort},
			AffectsContainers:    []ContainerID{ContainerID_Exporter, ContainerID_Prometheus},
			EnvironmentVariables: []string{"EXPORTER_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		WatchtowerMetricsPort: Parameter{
			ID:                   "watchtowerMetricsPort",
			Name:                 "Watchtower Metrics Port",
			Description:          "The port your Watchtower container should expose its metrics on.\nThis is only relevant for Oracle Nodes.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultWatchtowerMetricsPort},
			AffectsContainers:    []ContainerID{ContainerID_Watchtower, ContainerID_Prometheus},
			EnvironmentVariables: []string{"WATCHTOWER_METRICS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}

	// Set the defaults for choices
	config.ExecutionClientMode.Default[Network_All] = config.ExecutionClientMode.Options[0].Value
	config.FallbackExecutionClientMode.Default[Network_All] = config.FallbackExecutionClientMode.Options[0].Value
	config.ConsensusClientMode.Default[Network_All] = config.ConsensusClientMode.Options[0].Value

	config.Smartnode = NewSmartnodeConfig(config)
	config.ExecutionCommon = NewExecutionCommonConfig(config, false)
	config.Geth = NewGethConfig(config, false)
	config.Infura = NewInfuraConfig(config, false)
	config.Pocket = NewPocketConfig(config, false)
	config.ExternalExecution = NewExternalExecutionConfig(config)
	config.FallbackExecutionCommon = NewExecutionCommonConfig(config, true)
	config.FallbackInfura = NewInfuraConfig(config, true)
	config.FallbackPocket = NewPocketConfig(config, true)
	config.FallbackExternalExecution = NewExternalExecutionConfig(config)
	config.ConsensusCommon = NewConsensusCommonConfig(config)
	config.Lighthouse = NewLighthouseConfig(config)
	config.Nimbus = NewNimbusConfig(config)
	config.Prysm = NewPrysmConfig(config)
	config.Teku = NewTekuConfig(config)
	config.ExternalLighthouse = NewExternalLighthouseConfig(config)
	config.ExternalPrysm = NewExternalPrysmConfig(config)
	config.ExternalTeku = NewExternalTekuConfig(config)
	config.Grafana = NewGrafanaConfig(config)
	config.Prometheus = NewPrometheusConfig(config)
	config.Exporter = NewExporterConfig(config)
	config.Native = NewNativeConfig(config)

	// Apply the default values for mainnet
	config.Smartnode.Network.Value = config.Smartnode.Network.Options[0].Value
	config.applyAllDefaults()

	return config
}

// Create a copy of this configuration.
func (config *RocketPoolConfig) CreateCopy() *RocketPoolConfig {
	newConfig := NewRocketPoolConfig(config.RocketPoolDirectory, config.IsNativeMode)

	newParams := newConfig.GetParameters()
	for i, param := range config.GetParameters() {
		newParams[i].Value = param.Value
	}

	newSubconfigs := newConfig.GetSubconfigs()
	for name, subConfig := range config.GetSubconfigs() {
		newParams := newSubconfigs[name].GetParameters()
		for i, param := range subConfig.GetParameters() {
			newParams[i].Value = param.Value
		}
	}

	return newConfig
}

// Get the parameters for this config
func (config *RocketPoolConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.ExecutionClientMode,
		&config.ExecutionClient,
		&config.UseFallbackExecutionClient,
		&config.FallbackExecutionClientMode,
		&config.FallbackExecutionClient,
		&config.ReconnectDelay,
		&config.ConsensusClientMode,
		&config.ConsensusClient,
		&config.ExternalConsensusClient,
		&config.EnableMetrics,
		&config.BnMetricsPort,
		&config.VcMetricsPort,
		&config.NodeMetricsPort,
		&config.ExporterMetricsPort,
		&config.WatchtowerMetricsPort,
	}
}

// Get the subconfigurations for this config
func (config *RocketPoolConfig) GetSubconfigs() map[string]Config {
	return map[string]Config{
		"smartnode":                 config.Smartnode,
		"executionCommon":           config.ExecutionCommon,
		"geth":                      config.Geth,
		"infura":                    config.Infura,
		"pocket":                    config.Pocket,
		"externalExecution":         config.ExternalExecution,
		"fallbackExecutionCommon":   config.FallbackExecutionCommon,
		"fallbackInfura":            config.FallbackInfura,
		"fallbackPocket":            config.FallbackPocket,
		"fallbackExternalExecution": config.FallbackExternalExecution,
		"consensusCommon":           config.ConsensusCommon,
		"lighthouse":                config.Lighthouse,
		"nimbus":                    config.Nimbus,
		"prysm":                     config.Prysm,
		"teku":                      config.Teku,
		"externalLighthouse":        config.ExternalLighthouse,
		"externalPrysm":             config.ExternalPrysm,
		"externalTeku":              config.ExternalTeku,
		"grafana":                   config.Grafana,
		"prometheus":                config.Prometheus,
		"exporter":                  config.Exporter,
		"native":                    config.Native,
	}
}

// Handle a network change on all of the parameters
func (config *RocketPoolConfig) ChangeNetwork(newNetwork Network) {

	// Get the current network
	oldNetwork, ok := config.Smartnode.Network.Value.(Network)
	if !ok {
		oldNetwork = Network_Unknown
	}
	if oldNetwork == newNetwork {
		return
	}
	config.Smartnode.Network.Value = newNetwork

	// Update the master parameters
	rootParams := config.GetParameters()
	for _, param := range rootParams {
		param.changeNetwork(oldNetwork, newNetwork)
	}

	// Update all of the child config objects
	subconfigs := config.GetSubconfigs()
	for _, subconfig := range subconfigs {
		for _, param := range subconfig.GetParameters() {
			param.changeNetwork(oldNetwork, newNetwork)
		}
	}

}

// Get the Consensus clients compatible with the config's EC and fallback EC selection
func (config *RocketPoolConfig) GetCompatibleConsensusClients() ([]ParameterOption, []string) {

	// Get the compatible clients based on the EC choice
	var compatibleConsensusClients []ConsensusClient
	executionClient := config.ExecutionClient.Value.(ExecutionClient)
	switch executionClient {
	case ExecutionClient_Geth:
		compatibleConsensusClients = config.Geth.CompatibleConsensusClients
	case ExecutionClient_Infura:
		compatibleConsensusClients = config.Infura.CompatibleConsensusClients
	case ExecutionClient_Pocket:
		compatibleConsensusClients = config.Pocket.CompatibleConsensusClients
	}

	// Get the compatible clients based on the fallback EC choice
	var fallbackCompatibleConsensusClients []ConsensusClient
	if config.UseFallbackExecutionClient.Value == true {
		fallbackExecutionClient := config.FallbackExecutionClient.Value.(ExecutionClient)
		switch fallbackExecutionClient {
		case ExecutionClient_Infura:
			compatibleConsensusClients = config.FallbackInfura.CompatibleConsensusClients
		case ExecutionClient_Pocket:
			compatibleConsensusClients = config.FallbackPocket.CompatibleConsensusClients
		}
	}

	// Sort every consensus client into good and bad lists
	var goodClients []ParameterOption
	var badClients []string
	for _, consensusClient := range config.ConsensusClient.Options {
		// Get the value for one of the consensus client options
		clientValue := consensusClient.Value.(ConsensusClient)

		// Check if it's in the list of clients compatible with the EC
		isGood := false
		for _, compatibleWithEC := range compatibleConsensusClients {
			if compatibleWithEC == clientValue {
				isGood = true
				break
			}
		}

		// If it isn't, append it to the list of bad clients and move on
		if !isGood {
			badClients = append(badClients, consensusClient.Name)
			continue
		}

		// Check the fallback EC too
		if len(fallbackCompatibleConsensusClients) > 0 {
			isGood = false
			for _, compatibleWithFallbackEC := range fallbackCompatibleConsensusClients {
				if compatibleWithFallbackEC == clientValue {
					isGood = true
					break
				}
			}

			if !isGood {
				badClients = append(badClients, consensusClient.Name)
				continue
			}
		}

		// If we get here, it's compatible.
		goodClients = append(goodClients, consensusClient)
	}

	return goodClients, badClients

}

// Get the configuration for the selected client
func (config *RocketPoolConfig) GetSelectedConsensusClientConfig() (ConsensusConfig, error) {
	if config.IsNativeMode {
		return nil, fmt.Errorf("consensus config is not available in native mode")
	}

	mode := config.ConsensusClientMode.Value.(Mode)
	switch mode {
	case Mode_Local:
		client := config.ConsensusClient.Value.(ConsensusClient)
		switch client {
		case ConsensusClient_Lighthouse:
			return config.Lighthouse, nil
		case ConsensusClient_Nimbus:
			return config.Nimbus, nil
		case ConsensusClient_Prysm:
			return config.Prysm, nil
		case ConsensusClient_Teku:
			return config.Teku, nil
		default:
			return nil, fmt.Errorf("unknown consensus client [%v] selected", client)
		}

	case Mode_External:
		client := config.ExternalConsensusClient.Value.(ConsensusClient)
		switch client {
		case ConsensusClient_Lighthouse:
			return config.ExternalLighthouse, nil
		case ConsensusClient_Prysm:
			return config.ExternalPrysm, nil
		case ConsensusClient_Teku:
			return config.ExternalTeku, nil
		default:
			return nil, fmt.Errorf("unknown external consensus client [%v] selected", client)
		}

	default:
		return nil, fmt.Errorf("unknown consensus client mode [%v]", mode)
	}
}

// Serializes the configuration into a map of maps, compatible with a settings file
func (config *RocketPoolConfig) Serialize() map[string]map[string]string {

	masterMap := map[string]map[string]string{}

	// Serialize root params
	rootParams := map[string]string{}
	for _, param := range config.GetParameters() {
		param.serialize(rootParams)
	}
	masterMap[rootConfigName] = rootParams
	masterMap[rootConfigName]["rpDir"] = config.RocketPoolDirectory
	masterMap[rootConfigName]["isNative"] = fmt.Sprint(config.IsNativeMode)
	masterMap[rootConfigName]["version"] = fmt.Sprintf("v%s", shared.RocketPoolVersion) // Update the version with the current Smartnode version

	// Serialize the subconfigs
	for name, subconfig := range config.GetSubconfigs() {
		subconfigParams := map[string]string{}
		for _, param := range subconfig.GetParameters() {
			param.serialize(subconfigParams)
		}
		masterMap[name] = subconfigParams
	}

	return masterMap
}

// Deserializes a settings file into this config
func (config *RocketPoolConfig) Deserialize(masterMap map[string]map[string]string) error {

	// Deserialize root params
	rootParams, exists := masterMap[rootConfigName]
	if !exists {
		return fmt.Errorf("missing config section [%s]", rootConfigName)
	}
	for _, param := range config.GetParameters() {
		err := param.deserialize(rootParams)
		if err != nil {
			return fmt.Errorf("error deserializing root config: %w", err)
		}
	}

	var err error
	config.RocketPoolDirectory = masterMap[rootConfigName]["rpDir"]
	config.IsNativeMode, err = strconv.ParseBool(masterMap[rootConfigName]["isNative"])
	if err != nil {
		return fmt.Errorf("error parsing isNative: %w", err)
	}
	config.Version = masterMap[rootConfigName]["version"]

	// Deserialize the subconfigs
	for name, subconfig := range config.GetSubconfigs() {
		subconfigParams, exists := masterMap[name]
		if !exists {
			return fmt.Errorf("missing config section [%s]", name)
		}
		for _, param := range subconfig.GetParameters() {
			err := param.deserialize(subconfigParams)
			if err != nil {
				return fmt.Errorf("error deserializing [%s]: %w", name, err)
			}
		}
	}

	return nil
}

// Generates a collection of environment variables based on this config's settings
func (config *RocketPoolConfig) GenerateEnvironmentVariables() map[string]string {

	envVars := map[string]string{}

	// Basic variables and root parameters
	envVars["SMARTNODE_IMAGE"] = config.Smartnode.GetSmartnodeContainerTag()
	envVars["ROCKETPOOL_FOLDER"] = config.RocketPoolDirectory
	addParametersToEnvVars(config.Smartnode.GetParameters(), envVars)
	addParametersToEnvVars(config.GetParameters(), envVars)

	// EC parameters
	envVars["EC_CLIENT"] = fmt.Sprint(config.ExecutionClient.Value)
	if config.ExecutionClientMode.Value.(Mode) == Mode_Local {
		envVars["EC_HTTP_ENDPOINT"] = fmt.Sprintf("http://%s:%d", Eth1ContainerName, config.ExecutionCommon.HttpPort.Value)
		envVars["EC_WS_ENDPOINT"] = fmt.Sprintf("ws://%s:%d", Eth1ContainerName, config.ExecutionCommon.WsPort.Value)

		// Handle open API ports
		if config.ExecutionCommon.OpenRpcPorts.Value == true {
			ecHttpPort := config.ExecutionCommon.HttpPort.Value.(uint16)
			ecWsPort := config.ExecutionCommon.WsPort.Value.(uint16)
			envVars["EC_OPEN_API_PORTS"] = fmt.Sprintf(", \"%d:%d/tcp\", \"%d:%d/tcp\"", ecHttpPort, ecHttpPort, ecWsPort, ecWsPort)
		}

		// Common params
		addParametersToEnvVars(config.ExecutionCommon.GetParameters(), envVars)

		// Client-specific params
		switch config.ExecutionClient.Value.(ExecutionClient) {
		case ExecutionClient_Geth:
			addParametersToEnvVars(config.Geth.GetParameters(), envVars)
		case ExecutionClient_Infura:
			addParametersToEnvVars(config.Infura.GetParameters(), envVars)
		case ExecutionClient_Pocket:
			addParametersToEnvVars(config.Pocket.GetParameters(), envVars)
		}
	} else {
		addParametersToEnvVars(config.ExternalExecution.GetParameters(), envVars)
	}

	// Fallback EC parameters
	envVars["FALLBACK_EC_CLIENT"] = fmt.Sprint(config.FallbackExecutionClient.Value)
	if config.UseFallbackExecutionClient.Value == true {
		if config.FallbackExecutionClientMode.Value.(Mode) == Mode_Local {
			envVars["FALLBACK_EC_HTTP_ENDPOINT"] = fmt.Sprintf("http://%s:%d", Eth1FallbackContainerName, config.FallbackExecutionCommon.HttpPort.Value)
			envVars["FALLBACK_EC_WS_ENDPOINT"] = fmt.Sprintf("ws://%s:%d", Eth1FallbackContainerName, config.FallbackExecutionCommon.WsPort.Value)

			// Handle open API ports
			if config.FallbackExecutionCommon.OpenRpcPorts.Value == true {
				ecHttpPort := config.FallbackExecutionCommon.HttpPort.Value.(uint16)
				ecWsPort := config.FallbackExecutionCommon.WsPort.Value.(uint16)
				envVars["FALLBACK_EC_OPEN_API_PORTS"] = fmt.Sprintf("\"%d:%d/tcp\", \"%d:%d/tcp\"", ecHttpPort, ecHttpPort, ecWsPort, ecWsPort)
			}

			// Common params
			addParametersToEnvVars(config.FallbackExecutionCommon.GetParameters(), envVars)

			// Client-specific params
			switch config.FallbackExecutionClient.Value.(ExecutionClient) {
			case ExecutionClient_Infura:
				addParametersToEnvVars(config.FallbackInfura.GetParameters(), envVars)
			case ExecutionClient_Pocket:
				addParametersToEnvVars(config.FallbackPocket.GetParameters(), envVars)
			}
		} else {
			envVars["FALLBACK_EC_HTTP_ENDPOINT"] = fmt.Sprint(config.FallbackExternalExecution.HttpUrl.Value)
			envVars["FALLBACK_EC_WS_ENDPOINT"] = fmt.Sprint(config.FallbackExternalExecution.WsUrl.Value)
			addParametersToEnvVars(config.FallbackExternalExecution.GetParameters(), envVars)
		}
	}

	// CC parameters
	if config.ConsensusClientMode.Value.(Mode) == Mode_Local {
		envVars["CC_CLIENT"] = fmt.Sprint(config.ConsensusClient.Value)
		envVars["CC_API_ENDPOINT"] = fmt.Sprintf("http://%s:%d", Eth2ContainerName, config.ConsensusCommon.ApiPort.Value)

		// Handle open API ports
		bnOpenPorts := ""
		if config.ConsensusCommon.OpenApiPort.Value == true {
			ccApiPort := config.ConsensusCommon.ApiPort.Value.(uint16)
			bnOpenPorts += fmt.Sprintf(", \"%d:%d/tcp\"", ccApiPort, ccApiPort)
		}
		if config.ConsensusClient.Value.(ConsensusClient) == ConsensusClient_Prysm && config.Prysm.OpenRpcPort.Value == true {
			prysmRpcPort := config.Prysm.RpcPort.Value.(uint16)
			bnOpenPorts += fmt.Sprintf(", \"%d:%d/tcp\"", prysmRpcPort, prysmRpcPort)
		}
		envVars["BN_OPEN_PORTS"] = bnOpenPorts

		// Common params
		addParametersToEnvVars(config.ConsensusCommon.GetParameters(), envVars)

		// Client-specific params
		switch config.ConsensusClient.Value.(ConsensusClient) {
		case ConsensusClient_Lighthouse:
			addParametersToEnvVars(config.Lighthouse.GetParameters(), envVars)
		case ConsensusClient_Nimbus:
			addParametersToEnvVars(config.Nimbus.GetParameters(), envVars)
		case ConsensusClient_Prysm:
			addParametersToEnvVars(config.Prysm.GetParameters(), envVars)
			envVars["CC_RPC_ENDPOINT"] = fmt.Sprintf("http://%s:%d", Eth2ContainerName, config.Prysm.RpcPort.Value)
		case ConsensusClient_Teku:
			addParametersToEnvVars(config.Teku.GetParameters(), envVars)
		}
	} else {
		envVars["CC_CLIENT"] = fmt.Sprint(config.ExternalConsensusClient.Value)

		switch config.ExternalConsensusClient.Value.(ConsensusClient) {
		case ConsensusClient_Lighthouse:
			addParametersToEnvVars(config.ExternalLighthouse.GetParameters(), envVars)
		case ConsensusClient_Prysm:
			addParametersToEnvVars(config.ExternalPrysm.GetParameters(), envVars)
		case ConsensusClient_Teku:
			addParametersToEnvVars(config.ExternalTeku.GetParameters(), envVars)
		}
	}

	// Metrics
	if config.EnableMetrics.Value == true {
		addParametersToEnvVars(config.Exporter.GetParameters(), envVars)
		addParametersToEnvVars(config.Prometheus.GetParameters(), envVars)
		addParametersToEnvVars(config.Grafana.GetParameters(), envVars)

		if config.Exporter.RootFs.Value == true {
			envVars["EXPORTER_ROOTFS_COMMAND"] = ", \"--path.rootfs=/rootfs\""
			envVars["EXPORTER_ROOTFS_VOLUME"] = ", \"/:/rootfs:ro\""
		}

		if config.Prometheus.OpenPort.Value == true {
			envVars["PROMETHEUS_OPEN_PORTS"] = fmt.Sprintf("%d:%d/tcp", config.Prometheus.Port.Value, config.Prometheus.Port.Value)
		}

	}

	return envVars

}

// Applies all of the defaults to all of the settings that have them defined
func (config *RocketPoolConfig) applyAllDefaults() error {
	for _, param := range config.GetParameters() {
		err := param.setToDefault(config.Smartnode.Network.Value.(Network))
		if err != nil {
			return fmt.Errorf("error setting root parameter default: %w", err)
		}
	}

	for name, subconfig := range config.GetSubconfigs() {
		for _, param := range subconfig.GetParameters() {
			err := param.setToDefault(config.Smartnode.Network.Value.(Network))
			if err != nil {
				return fmt.Errorf("error setting parameter default for %s: %w", name, err)
			}
		}
	}

	return nil
}

// Add the parameters to the collection of environment variabes
func addParametersToEnvVars(params []*Parameter, envVars map[string]string) {
	for _, param := range params {
		for _, envVar := range param.EnvironmentVariables {
			if envVar != "" {
				envVars[envVar] = fmt.Sprint(param.Value)
			}
		}
	}
}

// The the title for the config
func (config *RocketPoolConfig) GetConfigTitle() string {
	return config.Title
}

// Update the default settings for all overwrite-on-upgrade parameters
func (config *RocketPoolConfig) UpdateDefaults() error {
	// Update the root params
	currentNetwork := config.Smartnode.Network.Value.(Network)
	for _, param := range config.GetParameters() {
		defaultValue, err := param.GetDefault(currentNetwork)
		if err != nil {
			return fmt.Errorf("error getting defaults for root param [%s] on network [%v]: %w", param.ID, currentNetwork, err)
		}
		if param.OverwriteOnUpgrade {
			param.Value = defaultValue
		}
	}

	// Update the subconfigs
	for subconfigName, subconfig := range config.GetSubconfigs() {
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
