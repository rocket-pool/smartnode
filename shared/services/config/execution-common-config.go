package config

// Param IDs
const ecHttpPortID string = "httpPort"
const ecWsPortID string = "wsPort"
const ecOpenRpcPortsID string = "openRpcPorts"

// Defaults
const defaultEcHttpPort uint16 = 8545
const defaultEcWsPort uint16 = 8546
const defaultOpenEcApiPort bool = false

// Configuration for the Execution client
type ExecutionCommonConfig struct {
	// The HTTP API port
	HttpPort Parameter

	// The Websocket API port
	WsPort Parameter

	// Toggle for forwarding the HTTP and Websocket API ports outside of Docker
	OpenRpcPorts Parameter
}

// Create a new ExecutionCommonConfig struct
func NewExecutionCommonConfig(config *MasterConfig) *ExecutionCommonConfig {
	return &ExecutionCommonConfig{
		HttpPort: Parameter{
			ID:                   ecHttpPortID,
			Name:                 "HTTP Port",
			Description:          "The port %s should use for its HTTP RPC endpoint.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultEcHttpPort},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth1, ContainerID_Eth2},
			EnvironmentVariables: []string{"EC_HTTP_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		WsPort: Parameter{
			ID:                   ecWsPortID,
			Name:                 "Websocket Port",
			Description:          "The port %s should use for its Websocket RPC endpoint.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultEcWsPort},
			AffectsContainers:    []ContainerID{ContainerID_Eth1, ContainerID_Eth2},
			EnvironmentVariables: []string{"EC_WS_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		OpenRpcPorts: Parameter{
			ID:                   ecOpenRpcPortsID,
			Name:                 "Open RPC Ports",
			Description:          "Open the HTTP and Websocket RPC ports to your local network, so other local machines can access your Execution Client's RPC endpoint.",
			Type:                 ParameterType_Bool,
			Default:              map[Network]interface{}{Network_All: defaultOpenEcApiPort},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Handle a network change on all of the parameters
func (config *ExecutionCommonConfig) changeNetwork(oldNetwork Network, newNetwork Network) {
	changeNetworkForParameter(&config.HttpPort, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.WsPort, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.OpenRpcPorts, oldNetwork, newNetwork)
}
