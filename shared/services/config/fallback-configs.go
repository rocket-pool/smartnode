package config

// Configuration for fallback Lighthouse
type FallbackNormalConfig struct {
	Title string `yaml:"-"`

	// The URL of the Execution Client HTTP endpoint
	EcHttpUrl Parameter `yaml:"ecHttpUrl,omitempty"`

	// The URL of the Beacon Node HTTP endpoint
	CcHttpUrl Parameter `yaml:"ccHttpUrl,omitempty"`
}

// Configuration for fallback Prysm
type FallbackPrysmConfig struct {
	Title string `yaml:"-"`

	// The URL of the Execution Client HTTP endpoint
	EcHttpUrl Parameter `yaml:"ecHttpUrl,omitempty"`

	// The URL of the Beacon Node HTTP endpoint
	CcHttpUrl Parameter `yaml:"ccHttpUrl,omitempty"`

	// The URL of the JSON-RPC endpoint for the Validator client
	JsonRpcUrl Parameter `yaml:"jsonRpcUrl,omitempty"`
}

// Generates a new FallbackNormalConfig configuration
func NewFallbackNormalConfig(config *RocketPoolConfig) *FallbackNormalConfig {
	return &FallbackNormalConfig{
		Title: "Fallback Client Settings",

		EcHttpUrl: Parameter{
			ID:                   "ecHttpUrl",
			Name:                 "Execution Client URL",
			Description:          "The URL of the HTTP API endpoint for your fallback Execution client.\n\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower},
			EnvironmentVariables: []string{"FALLBACK_EC_API_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		CcHttpUrl: Parameter{
			ID:                   "ccHttpUrl",
			Name:                 "Beacon Node URL",
			Description:          "The URL of the HTTP Beacon API endpoint for your fallback Consensus client.\n\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Validator, ContainerID_Watchtower},
			EnvironmentVariables: []string{"FALLBACK_CC_API_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Generates a new FallbackPrysmConfig configuration
func NewFallbackPrysmConfig(config *RocketPoolConfig) *FallbackPrysmConfig {
	return &FallbackPrysmConfig{
		Title: "Fallback Prysm Settings",

		EcHttpUrl: Parameter{
			ID:                   "ecHttpUrl",
			Name:                 "Execution Client URL",
			Description:          "The URL of the HTTP API endpoint for your fallback Execution client.\n\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower},
			EnvironmentVariables: []string{"FALLBACK_EC_API_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		CcHttpUrl: Parameter{
			ID:                   "ccHttpUrl",
			Name:                 "Beacon Node HTTP URL",
			Description:          "The URL of the HTTP Beacon API endpoint for your fallback Prysm client.\n\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Validator, ContainerID_Watchtower},
			EnvironmentVariables: []string{"FALLBACK_CC_API_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		JsonRpcUrl: Parameter{
			ID:                   "jsonRpcUrl",
			Name:                 "Beacon Node JSON-RPC URL",
			Description:          "The URL of the JSON-RPC API endpoint for your fallback client. Prysm's validator client will need this in order to connect to it.\n\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"FALLBACK_CC_RPC_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (config *FallbackNormalConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.EcHttpUrl,
		&config.CcHttpUrl,
	}
}

// Get the parameters for this config
func (config *FallbackPrysmConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.EcHttpUrl,
		&config.CcHttpUrl,
		&config.JsonRpcUrl,
	}
}

// The the title for the config
func (config *FallbackNormalConfig) GetConfigTitle() string {
	return config.Title
}

// The the title for the config
func (config *FallbackPrysmConfig) GetConfigTitle() string {
	return config.Title
}
