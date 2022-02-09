package config

import "reflect"

// The master configuration struct
type MasterConfig struct {

	// The Smartnode configuration
	Smartnode *SmartnodeConfig

	// Toggle for external Execution clients
	ExecutionClientMode Parameter

	// Local Execution client, if selected
	ExecutionClient Parameter

	// Execution client configurations
	ExecutionCommon   *ExecutionCommonParams
	Geth              *GethConfig
	Infura            *InfuraConfig
	Pocket            *PocketConfig
	ExternalExecution *ExternalExecutionConfig

	// Toggles for a fallback Execution client and client mode
	UseFallbackExecutionClient  Parameter
	FallbackExecutionClientMode Parameter

	// Fallback Execution client, if enabled
	FallbackExecutionClient Parameter

	// Fallback Execution client configurations
	FallbackExecutionCommon   *ExecutionCommonParams
	FallbackInfura            *InfuraConfig
	FallbackPocket            *PocketConfig
	FallbackExternalExecution *ExternalExecutionConfig

	// Toggle for external Consensus clients
	ConsensusClientMode Parameter

	// Selected Consensus client
	ConsensusClient         Parameter
	ExternalConsensusClient Parameter

	// Consensus client configurations
	ConsensusCommon   *ConsensusCommonConfig
	Lighthouse        *LighthouseConfig
	Nimbus            *NimbusConfig
	Prysm             *PrysmConfig
	Teku              *TekuConfig
	ExternalConsensus *ExternalConsensusConfig
	ExternalPrysm     *ExternalPrysmConfig

	// Toggle for metrics
	EnableMetrics Parameter

	// Metrics
	Grafana    *GrafanaConfig
	Prometheus *PrometheusConfig
	Exporter   *ExporterConfig
}

// Creates a new master Configuration instance
func NewMasterConfig() *MasterConfig {

	config := &MasterConfig{
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
				ID:          "local",
				Name:        "Locally Managed",
				Description: "(Default)\n\nAllow the Smartnode to manage an Execution client for you (Docker Mode)",
				Value:       Mode_Local,
			}, {
				ID:          "external",
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
			Default:              map[Network]interface{}{Network_All: nil},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				ID:          "geth",
				Name:        "Geth",
				Description: "Geth is one of the three original implementations of the Ethereum protocol. It is written in Go, fully open source and licensed under the GNU LGPL v3.",
				Value:       ExecutionClient_Geth,
			}, {
				ID:          "infura",
				Name:        "Infura",
				Description: "Use infura.io as a light client for Eth 1.0. Not recommended for use in production.",
				Value:       ExecutionClient_Infura,
			}, {
				ID:          "pocket",
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
				ID:          "local",
				Name:        "Locally Managed",
				Description: "Allow the Smartnode to manage a fallback Execution client for you (Docker Mode)",
				Value:       Mode_Local,
			}, {
				ID:          "external",
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
			Default:              map[Network]interface{}{Network_All: nil},
			AffectsContainers:    []ContainerID{ContainerID_Eth1Fallback},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				ID:          "infura",
				Name:        "Infura",
				Description: "Use infura.io as a light client for Eth 1.0. Not recommended for use in production.",
				Value:       ExecutionClient_Infura,
			}, {
				ID:          "pocket",
				Name:        "Pocket",
				Description: "Use Pocket Network as a decentralized light client for Eth 1.0. Suitable for use in production.",
				Value:       ExecutionClient_Pocket,
			}},
		},

		ConsensusClientMode: Parameter{
			ID:                   "consensusClientMode",
			Name:                 "Consensus Client Mode",
			Description:          "Choose which mode to use for your Consensus client - locally managed (Docker Mode), or externally managed (Hybrid Mode).",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: nil},
			AffectsContainers:    []ContainerID{ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				ID:          "local",
				Name:        "Locally Managed",
				Description: "Allow the Smartnode to manage a Consensus client for you (Docker Mode)",
			}, {
				ID:          "external",
				Name:        "Externally Managed",
				Description: "Use an existing Consensus client that you manage on your own (Hybrid Mode)",
			}},
		},

		ConsensusClient: Parameter{
			ID:                   "consensusClient",
			Name:                 "Consensus Client",
			Description:          "Select which Consensus client you would like to use.",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: nil},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				ID:          "lighthouse",
				Name:        "Lighthouse",
				Description: "Lighthouse is a Consensus client with a heavy focus on speed and security. The team behind it, Sigma Prime, is an information security and software engineering firm who have funded Lighthouse along with the Ethereum Foundation, Consensys, and private individuals. Lighthouse is built in Rust and offered under an Apache 2.0 License.",
				Value:       ConsensusClient_Lighthouse,
			}, {
				ID:          "nimbus",
				Name:        "Nimbus",
				Description: "Nimbus is a Consensus client implementation that strives to be as lightweight as possible in terms of resources used. This allows it to perform well on embedded systems, resource-restricted devices -- including Raspberry Pis and mobile devices -- and multi-purpose servers.",
				Value:       ConsensusClient_Nimbus,
			}, {
				ID:          "prysm",
				Name:        "Prysm",
				Description: "Prysm is a Go implementation of Ethereum Consensus protocol with a focus on usability, security, and reliability. Prysm is developed by Prysmatic Labs, a company with the sole focus on the development of their client. Prysm is written in Go and released under a GPL-3.0 license.",
				Value:       ConsensusClient_Prysm,
			}, {
				ID:          "teku",
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
			Default:              map[Network]interface{}{Network_All: nil},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				ID:          "lighthouse",
				Name:        "Lighthouse",
				Description: "Lighthouse is a Consensus client with a heavy focus on speed and security. The team behind it, Sigma Prime, is an information security and software engineering firm who have funded Lighthouse along with the Ethereum Foundation, Consensys, and private individuals. Lighthouse is built in Rust and offered under an Apache 2.0 License.",
				Value:       ConsensusClient_Lighthouse,
			}, {
				ID:          "prysm",
				Name:        "Prysm",
				Description: "Prysm is a Go implementation of Ethereum Consensus protocol with a focus on usability, security, and reliability. Prysm is developed by Prysmatic Labs, a company with the sole focus on the development of their client. Prysm is written in Go and released under a GPL-3.0 license.",
				Value:       ConsensusClient_Prysm,
			}, {
				ID:          "teku",
				Name:        "Teku",
				Description: "PegaSys Teku (formerly known as Artemis) is a Java-based Ethereum 2.0 client designed & built to meet institutional needs and security requirements. PegaSys is an arm of ConsenSys dedicated to building enterprise-ready clients and tools for interacting with the core Ethereum platform. Teku is Apache 2 licensed and written in Java, a language notable for its maturity & ubiquity.",
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
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}

	// Set the defaults for choices
	config.ExecutionClientMode.Default[Network_All] = config.ExecutionClientMode.Options[0]
	config.FallbackExecutionClientMode.Default[Network_All] = config.FallbackExecutionClientMode.Options[0]
	config.ConsensusClientMode.Default[Network_All] = config.ConsensusClientMode.Options[0]

	config.Smartnode = NewSmartnodeConfig(config)
	config.ExecutionCommon = NewExecutionCommonParams(config)
	config.Geth = NewGethConfig(config)
	config.Infura = NewInfuraConfig(config)
	config.Pocket = NewPocketConfig(config)
	config.ExternalExecution = NewExternalExecutionConfig(config)
	config.FallbackExecutionCommon = NewExecutionCommonParams(config)
	config.FallbackInfura = NewInfuraConfig(config)
	config.FallbackPocket = NewPocketConfig(config)
	config.FallbackExternalExecution = NewExternalExecutionConfig(config)
	config.ConsensusCommon = NewConsensusCommonConfig(config)
	config.Lighthouse = NewLighthouseConfig(config)
	config.Nimbus = NewNimbusConfig(config)
	config.Prysm = NewPrysmConfig(config)
	config.Teku = NewTekuConfig(config)
	config.ExternalConsensus = NewExternalConsensusConfig(config)
	config.ExternalPrysm = NewExternalPrysmConfig(config)
	config.Grafana = NewGrafanaConfig(config)
	config.Prometheus = NewPrometheusConfig(config)
	config.Exporter = NewExporterConfig(config)

	return config
}

// Triggers when a network change occurs
func (config *MasterConfig) ChangeNetwork(newNetwork Network) {
	currentNetwork, ok := config.Smartnode.Network.Value.(Network)
	if !ok {
		currentNetwork = Network_Unknown
	}
	if currentNetwork != newNetwork {
		changeNetworkImpl(config, currentNetwork, newNetwork)
	}
	config.Smartnode.Network.Value = newNetwork
}

// Reflectively go through a type and apply a network change to all of its parameters and children
func changeNetworkImpl(object interface{}, oldNetwork Network, newNetwork Network) {

	configValue := reflect.ValueOf(object)
	numberOfFields := configValue.NumField()

	for i := 0; i < numberOfFields; i++ {
		field := configValue.Field(i)
		if field.Kind() == reflect.Struct {
			// This is a struct, so recurse into it
			changeNetworkImpl(field, oldNetwork, newNetwork)
		} else {
			fieldAsParam, ok := field.Interface().(Parameter)
			if ok {
				// This is a Parameter - get the defaults and the current value
				currentValue := fieldAsParam.Value
				oldDefault, exists := fieldAsParam.Default[oldNetwork]
				if !exists {
					oldDefault = fieldAsParam.Default[Network_All]
				}
				newDefault, exists := fieldAsParam.Default[newNetwork]
				if !exists {
					newDefault = fieldAsParam.Default[Network_All]
				}

				// If the old value matches the old default, replace it with the new default
				if currentValue == oldDefault {
					fieldAsParam.Value = newDefault
				}
			}
		}
	}

}
