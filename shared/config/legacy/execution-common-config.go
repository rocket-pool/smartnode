package config

const (
	// Param IDs
	ecHttpPortID     string = "httpPort"
	ecWsPortID       string = "wsPort"
	ecEnginePortID   string = "enginePort"
	ecOpenRpcPortsID string = "openRpcPorts"

	// Defaults
	defaultEcP2pPort     uint16 = 30303
	defaultEcHttpPort    uint16 = 8545
	defaultEcWsPort      uint16 = 8546
	defaultEcEnginePort  uint16 = 8551
	defaultOpenEcApiPort string = string(RPC_Closed)
)

// Configuration for the Execution client
type ExecutionCommonConfig struct {
	Title string `yaml:"-"`

	// The HTTP API port
	HttpPort Parameter `yaml:"httpPort,omitempty"`

	// The Websocket API port
	WsPort Parameter `yaml:"wsPort,omitempty"`

	// The Engine API port
	EnginePort Parameter `yaml:"enginePort,omitempty"`

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
func NewExecutionCommonConfig(cfg *RocketPoolConfig) *ExecutionCommonConfig {
	rpcPortModes := PortModes("")

	return &ExecutionCommonConfig{
		Title: "Common Execution Client Settings",

		HttpPort: Parameter{
			ID:                 ecHttpPortID,
			Name:               "HTTP API Port",
			Description:        "The port your Execution client should use for its HTTP API endpoint (also known as HTTP RPC API endpoint).",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultEcHttpPort},
			AffectsContainers:  []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		WsPort: Parameter{
			ID:                 ecWsPortID,
			Name:               "Websocket API Port",
			Description:        "The port your Execution client should use for its Websocket API endpoint (also known as Websocket RPC API endpoint).",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultEcWsPort},
			AffectsContainers:  []ContainerID{ContainerID_Eth1, ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		EnginePort: Parameter{
			ID:                 ecEnginePortID,
			Name:               "Engine API Port",
			Description:        "The port your Execution client should use for its Engine API endpoint (the endpoint the Consensus client will connect to post-merge).",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultEcEnginePort},
			AffectsContainers:  []ContainerID{ContainerID_Eth1, ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		OpenRpcPorts: Parameter{
			ID:                 ecOpenRpcPortsID,
			Name:               "Expose RPC Ports",
			Description:        "Expose the HTTP and Websocket RPC ports to other processes on your machine, or to your local network so other machines can access your Execution Client's RPC endpoint.",
			Type:               ParameterType_Choice,
			Default:            map[Network]interface{}{Network_All: defaultOpenEcApiPort},
			AffectsContainers:  []ContainerID{ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            rpcPortModes,
		},

		P2pPort: Parameter{
			ID:                 "p2pPort",
			Name:               "P2P Port",
			Description:        "The port Geth should use for P2P (blockchain) traffic to communicate with other nodes.",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultEcP2pPort},
			AffectsContainers:  []ContainerID{ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		EthstatsLabel: Parameter{
			ID:                 "ethstatsLabel",
			Name:               "ETHStats Label",
			Description:        "If you would like to report your Execution client statistics to https://ethstats.net/, enter the label you want to use here.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		EthstatsLogin: Parameter{
			ID:                 "ethstatsLogin",
			Name:               "ETHStats Login",
			Description:        "If you would like to report your Execution client statistics to https://ethstats.net/, enter the login you want to use here.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *ExecutionCommonConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.HttpPort,
		&cfg.WsPort,
		&cfg.EnginePort,
		&cfg.OpenRpcPorts,
		&cfg.P2pPort,
		&cfg.EthstatsLabel,
		&cfg.EthstatsLogin,
	}
}

// The the title for the config
func (cfg *ExecutionCommonConfig) GetConfigTitle() string {
	return cfg.Title
}
