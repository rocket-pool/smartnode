package config

// Configuration for external Execution clients
type ExternalExecutionConfig struct {
	// The URL of the HTTP endpoint
	HttpUrl Parameter

	// The URL of the websocket endpoint
	WsUrl Parameter
}

// Generates a new ExternalExecutionConfig configuration
func NewExternalExecutionConfig(commonParams *ExecutionCommonParams) *ExternalExecutionConfig {
	return &ExternalExecutionConfig{
		HttpUrl: Parameter{
			ID:                   "httpUrl",
			Name:                 "HTTP URL",
			Description:          "The URL of the HTTP RPC endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limiations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              "",
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
			Default:              "",
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_EXTERNAL_WS_URL"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}
