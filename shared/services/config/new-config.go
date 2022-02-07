package config

type ContainerID int
type Network int
type ParameterType int

// Enum to describe which container(s) a parameter impacts, so the Smartnode knows which
// ones to restart upon a settings change
const (
	ContainerID_Unknown ContainerID = iota
	ContainerID_Api
	ContainerID_Node
	ContainerID_Watchtower
	ContainerID_Eth1
	ContainerID_Eth2
	ContainerID_Validator
	ContainerID_Grafana
	ContainerID_Prometheus
	ContainerID_Exporter
)

// Enum to describe which network the system is on
const (
	Network_Unknown Network = iota
	Network_Mainnet
	Network_Prater
)

// Enum to describe which data type a parameter's value will have, which
// informs the corresponding UI element and value validation
const (
	ParameterType_Unknown ParameterType = iota
	ParameterType_Int
	ParameterType_Uint16
	ParameterType_Uint
	ParameterType_String
	ParameterType_Bool
	ParameterType_Choice
)

// A parameter that can be configured by the user
type Parameter struct {
	ID                   string
	Name                 string
	Description          string
	Type                 ParameterType
	Default              interface{}
	AffectsContainers    []ContainerID
	EnvironmentVariables []string
	CanBeBlank           bool
	OverwriteOnUpgrade   bool
	Options              []ParameterOption
}

// A single option in a choice parameter
type ParameterOption struct {
	ID          string
	Name        string
	Description string
}

// The value for a parameter
type Setting struct {
	Parameter    *Parameter
	Value        interface{}
	UsingDefault bool
}

// Configuration for the Execution client
type ExecutionConfig struct {
	// External clients (Hybrid mode)
	UseExternalClient Parameter

	// Local clients (Docker mode)
	Client       Parameter
	ClientConfig interface{}
}

// Creates a new Execution client configuration
func NewExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{
		UseExternalClient: Parameter{
			ID:                "useExternalClient",
			Name:              "Use External Client",
			Description:       "Enable this if you already have an Execution client running, and you want the Smartnode to use it instead of managing its own (\"Hybrid mode\").",
			Type:              ParameterType_Bool,
			Default:           false,
			AffectsContainers: []ContainerID{ContainerID_Eth1},
		},

		Client: Parameter{
			ID:                   "client",
			Name:                 "Client",
			Description:          "Select which Execution client you would like to use.",
			Type:                 ParameterType_Choice,
			Default:              nil,
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				ID:          "geth",
				Name:        "Geth",
				Description: "Geth is one of the three original implementations of the Ethereum protocol. It is written in Go, fully open source and licensed under the GNU LGPL v3.",
			}, {
				ID:          "infura",
				Name:        "Infura",
				Description: "Use infura.io as a light client for Eth 1.0. Not recommended for use in production.",
			}, {
				ID:          "pocket",
				Name:        "Pocket",
				Description: "Use Pocket Network as a decentralized light client for Eth 1.0. Suitable for use in production.",
			}},
		},
	}
}

// The master configuration struct
type Configuration struct {

	// The Smartnode configuration
	Smartnode *SmartnodeConfig

	// Toggle for external Execution clients
	UseExternalExecutionClient Parameter

	// Local Execution client, if selected
	ExecutionClient Parameter

	// Execution client configurations
	ExecutionCommon *ExecutionCommonParams
	Geth            *GethConfig
	Infura          *InfuraConfig
	Pocket          *PocketConfig

	// Toggle for a fallback Execution client
	UseFallbackExecutionClient Parameter

	// Fallback Execution client, if enabled
	FallbackExecutionClient Parameter

	// Fallback Execution client configurations
	FallbackExecutionCommon *ExecutionCommonParams
	FallbackInfura          *InfuraConfig
	FallbackPocket          *PocketConfig

	// Toggle for external Consensus clients
	UseExternalConsensusClient Parameter

	// Selected Consensus client
	ConsensusClient Parameter

	// Consensus client configurations
	ConsensusCommon *ConsensusCommonParams
	Lighthouse      *LighthouseConfig
	Nimbus          *NimbusConfig
	Prysm           *PrysmConfig
	Teku            *TekuConfig

	// Toggle for metrics
	EnableMetrics Parameter

	// Metrics
	Grafana    *GrafanaConfig
	Prometheus *PrometheusConfig
	Exporter   *ExporterConfig
}
