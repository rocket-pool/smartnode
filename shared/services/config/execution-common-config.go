package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

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
	defaultOpenEcApiPort string = string(config.RPC_Closed)
)

// Configuration for the Execution client
type ExecutionCommonConfig struct {
	Title string `yaml:"-"`

	// The HTTP API port
	HttpPort config.Parameter `yaml:"httpPort,omitempty"`

	// The Websocket API port
	WsPort config.Parameter `yaml:"wsPort,omitempty"`

	// The Engine API port
	EnginePort config.Parameter `yaml:"enginePort,omitempty"`

	// Toggle for forwarding the HTTP and Websocket API ports outside of Docker
	OpenRpcPorts config.Parameter `yaml:"openRpcPorts,omitempty"`

	// P2P traffic port
	P2pPort config.Parameter `yaml:"p2pPort,omitempty"`

	// The suggested block gas limit
	SuggestedBlockGasLimit config.Parameter `yaml:"suggestedBlockGasLimit,omitempty"`

	// Label for Ethstats
	EthstatsLabel config.Parameter `yaml:"ethstatsLabel,omitempty"`

	// Login info for Ethstats
	EthstatsLogin config.Parameter `yaml:"ethstatsLogin,omitempty"`
}

// Create a new ExecutionCommonConfig struct
func NewExecutionCommonConfig(cfg *RocketPoolConfig) *ExecutionCommonConfig {
	rpcPortModes := config.PortModes("")

	return &ExecutionCommonConfig{
		Title: "Common Execution Client Settings",

		HttpPort: config.Parameter{
			ID:                 ecHttpPortID,
			Name:               "HTTP API Port",
			Description:        "The port your Execution client should use for its HTTP API endpoint (also known as HTTP RPC API endpoint).",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultEcHttpPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth1, config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		WsPort: config.Parameter{
			ID:                 ecWsPortID,
			Name:               "Websocket API Port",
			Description:        "The port your Execution client should use for its Websocket API endpoint (also known as Websocket RPC API endpoint).",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultEcWsPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1, config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		EnginePort: config.Parameter{
			ID:                 ecEnginePortID,
			Name:               "Engine API Port",
			Description:        "The port your Execution client should use for its Engine API endpoint (the endpoint the Consensus client will connect to post-merge).",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultEcEnginePort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1, config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		OpenRpcPorts: config.Parameter{
			ID:                 ecOpenRpcPortsID,
			Name:               "Expose RPC Ports",
			Description:        "Expose the HTTP and Websocket RPC ports to other processes on your machine, or to your local network so other machines can access your Execution Client's RPC endpoint.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: defaultOpenEcApiPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            rpcPortModes,
		},

		SuggestedBlockGasLimit: config.Parameter{
			ID:                 "suggestedBlockGasLimit",
			Name:               "Suggested Block Gas Limit",
			Description:        "The block gas limit that should be used for locally built blocks. Leave blank to follow the Execution Client's default.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		P2pPort: config.Parameter{
			ID:                 "p2pPort",
			Name:               "P2P Port",
			Description:        "The port the Execution Client should use for P2P (blockchain) traffic to communicate with other nodes.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultEcP2pPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		EthstatsLabel: config.Parameter{
			ID:                 "ethstatsLabel",
			Name:               "ETHStats Label",
			Description:        "If you would like to report your Execution client statistics to https://ethstats.net/, enter the label you want to use here.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		EthstatsLogin: config.Parameter{
			ID:                 "ethstatsLogin",
			Name:               "ETHStats Login",
			Description:        "If you would like to report your Execution client statistics to https://ethstats.net/, enter the login you want to use here.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *ExecutionCommonConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.HttpPort,
		&cfg.WsPort,
		&cfg.EnginePort,
		&cfg.OpenRpcPorts,
		&cfg.SuggestedBlockGasLimit,
		&cfg.P2pPort,
		&cfg.EthstatsLabel,
		&cfg.EthstatsLogin,
	}
}

// The title for the config
func (cfg *ExecutionCommonConfig) GetConfigTitle() string {
	return cfg.Title
}
