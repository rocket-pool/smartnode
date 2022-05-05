package config

const (
	// Param IDs
	ecHttpPortID     string = "httpPort"
	ecWsPortID       string = "wsPort"
	ecOpenRpcPortsID string = "openRpcPorts"

	// Defaults
	defaultEcP2pPort     uint16 = 30303
	defaultEcHttpPort    uint16 = 8545
	defaultEcWsPort      uint16 = 8546
	defaultOpenEcApiPort bool   = false
)

// Configuration for the Execution client
type ExecutionCommonConfig struct {
	Title string `yaml:"-"`

	// The HTTP API port
	HttpPort Parameter `yaml:"httpPort,omitempty"`

	// The Websocket API port
	WsPort Parameter `yaml:"wsPort,omitempty"`

	// Toggle for forwarding the HTTP and Websocket API ports outside of Docker
	OpenRpcPorts Parameter `yaml:"openRpcPorts,omitempty"`

	// P2P traffic port
	P2pPort Parameter `yaml:"p2pPort,omitempty"`

	// Label for Ethstats
	EthstatsLabel Parameter `yaml:"ethstatsLabel,omitempty"`

	// Login info for Ethstats
	EthstatsLogin Parameter `yaml:"ethstatsLogin,omitempty"`
}

// Create a new ExecutionCommonConfig struct
func NewExecutionCommonConfig(config *RocketPoolConfig, isFallback bool) *ExecutionCommonConfig {

	prefix := ""
	if isFallback {
		prefix = "FALLBACK_"
	}

	title := "Common Execution Client Settings"
	if isFallback {
		title = "Common Fallback Execution Client Settings"
	}

	return &ExecutionCommonConfig{
		Title: title,

		HttpPort: Parameter{
			ID:                   ecHttpPortID,
			Name:                 "HTTP Port",
			Description:          "The port your Execution client should use for its HTTP RPC endpoint.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultEcHttpPort},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2},
			EnvironmentVariables: []string{prefix + "EC_HTTP_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		WsPort: Parameter{
			ID:                   ecWsPortID,
			Name:                 "Websocket Port",
			Description:          "The port your Execution client should use for its Websocket RPC endpoint.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultEcWsPort},
			AffectsContainers:    []ContainerID{ContainerID_Eth1, ContainerID_Eth2},
			EnvironmentVariables: []string{prefix + "EC_WS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		OpenRpcPorts: Parameter{
			ID:                   ecOpenRpcPortsID,
			Name:                 "Expose RPC Ports",
			Description:          "Expose the HTTP and Websocket RPC ports to your local network, so other local machines can access your Execution Client's RPC endpoint.",
			Type:                 ParameterType_Bool,
			Default:              map[Network]interface{}{Network_All: defaultOpenEcApiPort},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		P2pPort: Parameter{
			ID:                   "p2pPort",
			Name:                 "P2P Port",
			Description:          "The port Geth should use for P2P (blockchain) traffic to communicate with other nodes.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultEcP2pPort},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_P2P_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		EthstatsLabel: Parameter{
			ID:                   "ethstatsLabel",
			Name:                 "ETHStats Label",
			Description:          "If you would like to report your Execution client statistics to https://ethstats.net/, enter the label you want to use here.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "ETHSTATS_LABEL"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		EthstatsLogin: Parameter{
			ID:                   "ethstatsLogin",
			Name:                 "ETHStats Login",
			Description:          "If you would like to report your Execution client statistics to https://ethstats.net/, enter the login you want to use here.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "ETHSTATS_LOGIN"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (config *ExecutionCommonConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.HttpPort,
		&config.WsPort,
		&config.OpenRpcPorts,
		&config.P2pPort,
		&config.EthstatsLabel,
		&config.EthstatsLogin,
	}
}

// The the title for the config
func (config *ExecutionCommonConfig) GetConfigTitle() string {
	return config.Title
}
