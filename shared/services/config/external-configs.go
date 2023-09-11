package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Configuration for external Execution clients
type ExternalExecutionConfig struct {
	Title string `yaml:"-"`

	// The URL of the HTTP endpoint
	HttpUrl config.Parameter `yaml:"httpUrl,omitempty"`

	// The URL of the websocket endpoint
	WsUrl config.Parameter `yaml:"wsUrl,omitempty"`
}

// Configuration for external Consensus clients
type ExternalLighthouseConfig struct {
	Title string `yaml:"-"`

	// The URL of the HTTP endpoint
	HttpUrl config.Parameter `yaml:"httpUrl,omitempty"`

	// Custom proposal graffiti
	Graffiti config.Parameter `yaml:"graffiti,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection config.Parameter `yaml:"doppelgangerDetection,omitempty"`

	// The Docker Hub tag for Lighthouse
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags config.Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Configuration for external Consensus clients
type ExternalLodestarConfig struct {
	Title string `yaml:"-"`

	// The URL of the HTTP endpoint
	HttpUrl config.Parameter `yaml:"httpUrl,omitempty"`

	// Custom proposal graffiti
	Graffiti config.Parameter `yaml:"graffiti,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection config.Parameter `yaml:"doppelgangerDetection,omitempty"`

	// The Docker Hub tag for Lighthouse
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags config.Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Configuration for external Consensus clients
type ExternalNimbusConfig struct {
	Title string `yaml:"-"`

	// The URL of the HTTP endpoint
	HttpUrl config.Parameter `yaml:"httpUrl,omitempty"`

	// Custom proposal graffiti
	Graffiti config.Parameter `yaml:"graffiti,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection config.Parameter `yaml:"doppelgangerDetection,omitempty"`

	// The Docker Hub tag for Lighthouse
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags config.Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Configuration for an external Prysm clients
type ExternalPrysmConfig struct {
	Title string `yaml:"-"`

	// The URL of the gRPC (REST) endpoint for the Beacon API
	HttpUrl config.Parameter `yaml:"httpUrl,omitempty"`

	// Custom proposal graffiti
	Graffiti config.Parameter `yaml:"graffiti,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection config.Parameter `yaml:"doppelgangerDetection,omitempty"`

	// The URL of the JSON-RPC endpoint for the Validator client
	JsonRpcUrl config.Parameter `yaml:"jsonRpcUrl,omitempty"`

	// The Docker Hub tag for Prysm's VC
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags config.Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Configuration for an external Teku client
type ExternalTekuConfig struct {
	Title string `yaml:"-"`

	// The URL of the HTTP endpoint
	HttpUrl config.Parameter `yaml:"httpUrl,omitempty"`

	// Custom proposal graffiti
	Graffiti config.Parameter `yaml:"graffiti,omitempty"`

	// The Docker Hub tag for Teku
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags config.Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Generates a new ExternalExecutionConfig configuration
func NewExternalExecutionConfig(cfg *RocketPoolConfig) *ExternalExecutionConfig {
	return &ExternalExecutionConfig{
		Title: "External Execution Client Settings",

		HttpUrl: config.Parameter{
			ID:                   "httpUrl",
			Name:                 "HTTP URL",
			Description:          "The URL of the HTTP RPC endpoint for your external Execution client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead, for example 'http://192.168.1.100:8545'.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Eth2, config.ContainerID_Node, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{"EC_HTTP_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		WsUrl: config.Parameter{
			ID:                   "wsUrl",
			Name:                 "Websocket URL",
			Description:          "The URL of the Websocket RPC endpoint for your external Execution client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead, for example 'http://192.168.1.100:8546'.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Api, config.ContainerID_Eth2, config.ContainerID_Node, config.ContainerID_Watchtower},
			EnvironmentVariables: []string{"EC_WS_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Generates a new ExternalLighthouseClient configuration
func NewExternalLighthouseConfig(cfg *RocketPoolConfig) *ExternalLighthouseConfig {
	return &ExternalLighthouseConfig{
		Title: "External Lighthouse Settings",

		HttpUrl: config.Parameter{
			ID:                   "httpUrl",
			Name:                 "HTTP URL",
			Description:          "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1, config.ContainerID_Api, config.ContainerID_Validator, config.ContainerID_Watchtower, config.ContainerID_Node},
			EnvironmentVariables: []string{"CC_API_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Graffiti: config.Parameter{
			ID:                   GraffitiID,
			Name:                 "Custom Graffiti",
			Description:          "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: defaultGraffiti},
			MaxLength:            16,
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{CustomGraffitiEnvVar},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		DoppelgangerDetection: config.Parameter{
			ID:                   DoppelgangerDetectionID,
			Name:                 "Enable Doppelgänger Detection",
			Description:          "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: defaultDoppelgangerDetection},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"DOPPELGANGER_DETECTION"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Lighthouse container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: getLighthouseTagProd(),
				config.Network_Prater:  getLighthouseTagTest(),
				config.Network_Devnet:  getLighthouseTagTest(),
				config.Network_Holesky: getLighthouseTagTest(),
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalVcFlags: config.Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Lighthouse's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Generates a new ExternalLodestarClient configuration
func NewExternalLodestarConfig(cfg *RocketPoolConfig) *ExternalLodestarConfig {
	return &ExternalLodestarConfig{
		Title: "External Lodestar Settings",

		HttpUrl: config.Parameter{
			ID:                   "httpUrl",
			Name:                 "HTTP URL",
			Description:          "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1, config.ContainerID_Api, config.ContainerID_Validator, config.ContainerID_Watchtower, config.ContainerID_Node},
			EnvironmentVariables: []string{"CC_API_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Graffiti: config.Parameter{
			ID:                   GraffitiID,
			Name:                 "Custom Graffiti",
			Description:          "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: defaultGraffiti},
			MaxLength:            16,
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{CustomGraffitiEnvVar},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		DoppelgangerDetection: config.Parameter{
			ID:                   DoppelgangerDetectionID,
			Name:                 "Enable Doppelgänger Detection",
			Description:          "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: defaultDoppelgangerDetection},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"DOPPELGANGER_DETECTION"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Lodestar container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: lodestarTagProd,
				config.Network_Prater:  lodestarTagTest,
				config.Network_Devnet:  lodestarTagTest,
				config.Network_Holesky: lodestarTagTest,
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalVcFlags: config.Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Lodestar's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Generates a new ExternalNimbusConfig configuration
func NewExternalNimbusConfig(cfg *RocketPoolConfig) *ExternalNimbusConfig {

	return &ExternalNimbusConfig{
		Title: "External Nimbus Settings",

		HttpUrl: config.Parameter{
			ID:                   "httpUrl",
			Name:                 "HTTP URL",
			Description:          "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1, config.ContainerID_Api, config.ContainerID_Validator, config.ContainerID_Watchtower, config.ContainerID_Node},
			EnvironmentVariables: []string{"CC_API_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Graffiti: config.Parameter{
			ID:                   GraffitiID,
			Name:                 "Custom Graffiti",
			Description:          "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: defaultGraffiti},
			MaxLength:            16,
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"CUSTOM_GRAFFITI"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		DoppelgangerDetection: config.Parameter{
			ID:                   DoppelgangerDetectionID,
			Name:                 "Enable Doppelgänger Detection",
			Description:          "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: defaultDoppelgangerDetection},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"DOPPELGANGER_DETECTION"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Nimbus validator container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: nimbusVcTagProd,
				config.Network_Prater:  nimbusVcTagTest,
				config.Network_Devnet:  nimbusVcTagTest,
				config.Network_Holesky: nimbusVcTagTest,
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalVcFlags: config.Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Nimbus's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Generates a new ExternalPrysmConfig configuration
func NewExternalPrysmConfig(cfg *RocketPoolConfig) *ExternalPrysmConfig {
	return &ExternalPrysmConfig{
		Title: "External Prysm Settings",

		HttpUrl: config.Parameter{
			ID:                   "httpUrl",
			Name:                 "HTTP URL",
			Description:          "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1, config.ContainerID_Api, config.ContainerID_Validator, config.ContainerID_Watchtower, config.ContainerID_Node},
			EnvironmentVariables: []string{"CC_API_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		JsonRpcUrl: config.Parameter{
			ID:                   "jsonRpcUrl",
			Name:                 "gRPC URL",
			Description:          "The URL of the gRPC API endpoint for your external client. Prysm's validator client will need this in order to connect to it.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"CC_RPC_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Graffiti: config.Parameter{
			ID:                   GraffitiID,
			Name:                 "Custom Graffiti",
			Description:          "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: defaultGraffiti},
			MaxLength:            16,
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{CustomGraffitiEnvVar},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		DoppelgangerDetection: config.Parameter{
			ID:                   DoppelgangerDetectionID,
			Name:                 "Enable Doppelgänger Detection",
			Description:          "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: defaultDoppelgangerDetection},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"DOPPELGANGER_DETECTION"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Prysm validator container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: getPrysmVcProdTag(),
				config.Network_Prater:  getPrysmVcTestTag(),
				config.Network_Devnet:  getPrysmVcTestTag(),
				config.Network_Holesky: getPrysmVcTestTag(),
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalVcFlags: config.Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Prysm's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Generates a new ExternalTekuClient configuration
func NewExternalTekuConfig(cfg *RocketPoolConfig) *ExternalTekuConfig {
	return &ExternalTekuConfig{
		Title: "External Teku Settings",

		HttpUrl: config.Parameter{
			ID:                   "httpUrl",
			Name:                 "HTTP URL",
			Description:          "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1, config.ContainerID_Api, config.ContainerID_Validator, config.ContainerID_Watchtower, config.ContainerID_Node},
			EnvironmentVariables: []string{"CC_API_ENDPOINT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Graffiti: config.Parameter{
			ID:                   GraffitiID,
			Name:                 "Custom Graffiti",
			Description:          "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: defaultGraffiti},
			MaxLength:            16,
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{CustomGraffitiEnvVar},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Teku container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: tekuTagProd,
				config.Network_Prater:  tekuTagTest,
				config.Network_Devnet:  tekuTagTest,
				config.Network_Holesky: tekuTagTest,
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalVcFlags: config.Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Teku's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (cfg *ExternalExecutionConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.HttpUrl,
		&cfg.WsUrl,
	}
}

// Get the parameters for this config
func (cfg *ExternalLighthouseConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.HttpUrl,
		&cfg.Graffiti,
		&cfg.DoppelgangerDetection,
		&cfg.ContainerTag,
		&cfg.AdditionalVcFlags,
	}
}

// Get the parameters for this config
func (cfg *ExternalNimbusConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.HttpUrl,
		&cfg.Graffiti,
		&cfg.DoppelgangerDetection,
		&cfg.ContainerTag,
		&cfg.AdditionalVcFlags,
	}
}

// Get the parameters for this config
func (cfg *ExternalLodestarConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.HttpUrl,
		&cfg.Graffiti,
		&cfg.DoppelgangerDetection,
		&cfg.ContainerTag,
		&cfg.AdditionalVcFlags,
	}
}

// Get the parameters for this config
func (cfg *ExternalPrysmConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.HttpUrl,
		&cfg.JsonRpcUrl,
		&cfg.Graffiti,
		&cfg.DoppelgangerDetection,
		&cfg.ContainerTag,
		&cfg.AdditionalVcFlags,
	}
}

// Get the parameters for this config
func (cfg *ExternalTekuConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.HttpUrl,
		&cfg.Graffiti,
		&cfg.ContainerTag,
		&cfg.AdditionalVcFlags,
	}
}

// Get the Docker container name of the validator client
func (cfg *ExternalLighthouseConfig) GetValidatorImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the Docker container name of the validator client
func (cfg *ExternalLodestarConfig) GetValidatorImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the Docker container name of the validator client
func (cfg *ExternalNimbusConfig) GetValidatorImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the Docker container name of the validator client
func (cfg *ExternalPrysmConfig) GetValidatorImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the Docker container name of the validator client
func (cfg *ExternalTekuConfig) GetValidatorImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the API url from the config
func (cfg *ExternalLighthouseConfig) GetApiUrl() string {
	return cfg.HttpUrl.Value.(string)
}

// Get the API url from the config
func (cfg *ExternalNimbusConfig) GetApiUrl() string {
	return cfg.HttpUrl.Value.(string)
}

// Get the API url from the config
func (cfg *ExternalLodestarConfig) GetApiUrl() string {
	return cfg.HttpUrl.Value.(string)
}

// Get the API url from the config
func (cfg *ExternalPrysmConfig) GetApiUrl() string {
	return cfg.HttpUrl.Value.(string)
}

// Get the API url from the config
func (cfg *ExternalTekuConfig) GetApiUrl() string {
	return cfg.HttpUrl.Value.(string)
}

// Get the name of the client
func (cfg *ExternalLighthouseConfig) GetName() string {
	return "Lighthouse"
}

// Get the name of the client
func (cfg *ExternalNimbusConfig) GetName() string {
	return "Nimbus"
}

// Get the name of the client
func (cfg *ExternalLodestarConfig) GetName() string {
	return "Lodestar"
}

// Get the name of the client
func (cfg *ExternalPrysmConfig) GetName() string {
	return "Prysm"
}

// Get the name of the client
func (cfg *ExternalTekuConfig) GetName() string {
	return "Teku"
}

// The the title for the config
func (cfg *ExternalExecutionConfig) GetConfigTitle() string {
	return cfg.Title
}

// The the title for the config
func (cfg *ExternalLighthouseConfig) GetConfigTitle() string {
	return cfg.Title
}

// The the title for the config
func (cfg *ExternalLodestarConfig) GetConfigTitle() string {
	return cfg.Title
}

// The the title for the config
func (cfg *ExternalNimbusConfig) GetConfigTitle() string {
	return cfg.Title
}

// The the title for the config
func (cfg *ExternalPrysmConfig) GetConfigTitle() string {
	return cfg.Title
}

// The the title for the config
func (cfg *ExternalTekuConfig) GetConfigTitle() string {
	return cfg.Title
}
