package config

// Configuration for external Execution clients
type ExternalExecutionConfig struct {
	// The URL of the HTTP endpoint
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// The URL of the websocket endpoint
	WsUrl Parameter `yaml:"wsUrl,omitempty"`
}

// Configuration for external Consensus clients
type ExternalLighthouseConfig struct {
	// The URL of the HTTP endpoint
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// The Docker Hub tag for Lighthouse
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Configuration for an external Prysm clients
type ExternalPrysmConfig struct {
	// The URL of the gRPC (REST) endpoint for the Beacon API
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// The URL of the JSON-RPC endpoint for the Validator client
	JsonRpcUrl Parameter `yaml:"jsonRpcUrl,omitempty"`

	// The Docker Hub tag for Prysm's VC
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Configuration for an external Teku client
type ExternalTekuConfig struct {
	// The URL of the HTTP endpoint
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// The Docker Hub tag for Teku
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Generates a new ExternalExecutionConfig configuration
func NewExternalExecutionConfig(config *RocketPoolConfig) *ExternalExecutionConfig {
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

// Generates a new ExternalLighthouseClient configuration
func NewExternalLighthouseConfig(config *RocketPoolConfig) *ExternalLighthouseConfig {
	return &ExternalLighthouseConfig{
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

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Lighthouse container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: lighthouseTag},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalVcFlags: Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Lighthouse's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Generates a new ExternalPrysmConfig configuration
func NewExternalPrysmConfig(config *RocketPoolConfig) *ExternalPrysmConfig {
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

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Prysm validator container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: getPrysmVcTag()},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalVcFlags: Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Prysm's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Generates a new ExternalTekuClient configuration
func NewExternalTekuConfig(config *RocketPoolConfig) *ExternalTekuConfig {
	return &ExternalTekuConfig{
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

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Teku container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: tekuTag},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalVcFlags: Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Teku's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
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
func (config *ExternalLighthouseConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.HttpUrl,
		&config.ContainerTag,
		&config.AdditionalVcFlags,
	}
}

// Get the parameters for this config
func (config *ExternalPrysmConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.HttpUrl,
		&config.JsonRpcUrl,
		&config.ContainerTag,
		&config.AdditionalVcFlags,
	}
}

// Get the parameters for this config
func (config *ExternalTekuConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.HttpUrl,
		&config.ContainerTag,
		&config.AdditionalVcFlags,
	}
}

// Get the Docker container name of the validator client
func (config *ExternalLighthouseConfig) GetValidatorImage() string {
	return config.ContainerTag.Value.(string)
}

// Get the Docker container name of the validator client
func (config *ExternalPrysmConfig) GetValidatorImage() string {
	return config.ContainerTag.Value.(string)
}

// Get the Docker container name of the validator client
func (config *ExternalTekuConfig) GetValidatorImage() string {
	return config.ContainerTag.Value.(string)
}

// Get the API url from the config
func (config *ExternalLighthouseConfig) GetApiUrl() string {
	return config.HttpUrl.Value.(string)
}

// Get the API url from the config
func (config *ExternalPrysmConfig) GetApiUrl() string {
	return config.HttpUrl.Value.(string)
}

// Get the API url from the config
func (config *ExternalTekuConfig) GetApiUrl() string {
	return config.HttpUrl.Value.(string)
}

// Get the name of the client
func (config *ExternalLighthouseConfig) GetName() string {
	return "Lighthouse"
}

// Get the name of the client
func (config *ExternalPrysmConfig) GetName() string {
	return "Prysm"
}

// Get the name of the client
func (config *ExternalTekuConfig) GetName() string {
	return "Teku"
}
