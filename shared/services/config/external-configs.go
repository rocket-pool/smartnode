package config

// Configuration for external Execution clients
type ExternalExecutionConfig struct {
	// The URL of the HTTP endpoint
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// The URL of the websocket endpoint
	WsUrl Parameter `yaml:"wsUrl,omitempty"`
}

// Configuration for external Consensus clients
type ExternalConsensusConfig struct {
	// The URL of the HTTP endpoint
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`
}

// Configuration for external Consensus clients
type ExternalPrysmConfig struct {
	// The URL of the gRPC (REST) endpoint for the Beacon API
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// The URL of the JSON-RPC endpoint for the Validator client
	JsonRpcUrl Parameter `yaml:"jsonRpcUrl,omitempty"`
}

// Generates a new ExternalExecutionConfig configuration
func NewExternalExecutionConfig(config *MasterConfig) *ExternalExecutionConfig {
	return &ExternalExecutionConfig{
		HttpUrl: Parameter{
			ID:                   "httpUrl",
			Name:                 "HTTP URL",
			Description:          "The URL of the HTTP RPC endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limiations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_EXTERNAL_HTTP_URL"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		WsUrl: Parameter{
			ID:                   "wsUrl",
			Name:                 "Websocket URL",
			Description:          "The URL of the Websocket RPC endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limiations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_EXTERNAL_WS_URL"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Generates a new ExternalConsensusClient configuration
func NewExternalConsensusConfig(config *MasterConfig) *ExternalConsensusConfig {
	return &ExternalConsensusConfig{
		HttpUrl: Parameter{
			ID:                   "httpUrl",
			Name:                 "HTTP URL",
			Description:          "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limiations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"CC_EXTERNAL_HTTP_URL"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Generates a new ExternalPrysmConfig configuration
func NewExternalPrysmConfig(config *MasterConfig) *ExternalPrysmConfig {
	return &ExternalPrysmConfig{
		HttpUrl: Parameter{
			ID:                   "httpUrl",
			Name:                 "HTTP URL",
			Description:          "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limiations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"CC_EXTERNAL_HTTP_URL"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		JsonRpcUrl: Parameter{
			ID:                   "jsonRpcUrl",
			Name:                 "JSON-RPC URL",
			Description:          "The URL of the JSON-RPC API endpoint for your external client. Prysm's validator client will need this in order to connect to it.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limiations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"CC_EXTERNAL_JSON_RPC_URL"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (config *ExternalExecutionConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.HttpUrl,
		&config.WsUrl,
	}
}

// Get the parameters for this config
func (config *ExternalConsensusConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.HttpUrl,
	}
}

// Get the parameters for this config
func (config *ExternalPrysmConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.HttpUrl,
		&config.JsonRpcUrl,
	}
}
