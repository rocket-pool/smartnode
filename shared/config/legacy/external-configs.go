package config

// Configuration for external Execution clients
type ExternalExecutionConfig struct {
	Title string `yaml:"-"`

	// The URL of the HTTP endpoint
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// The URL of the websocket endpoint
	WsUrl Parameter `yaml:"wsUrl,omitempty"`
}

// Configuration for external Consensus clients
type ExternalLighthouseConfig struct {
	Title string `yaml:"-"`

	// The URL of the HTTP endpoint
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// Custom proposal graffiti
	Graffiti Parameter `yaml:"graffiti,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection Parameter `yaml:"doppelgangerDetection,omitempty"`

	// The Docker Hub tag for Lighthouse
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Configuration for external Consensus clients
type ExternalLodestarConfig struct {
	Title string `yaml:"-"`

	// The URL of the HTTP endpoint
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// Custom proposal graffiti
	Graffiti Parameter `yaml:"graffiti,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection Parameter `yaml:"doppelgangerDetection,omitempty"`

	// The Docker Hub tag for Lighthouse
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Configuration for external Consensus clients
type ExternalNimbusConfig struct {
	Title string `yaml:"-"`

	// The URL of the HTTP endpoint
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// Custom proposal graffiti
	Graffiti Parameter `yaml:"graffiti,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection Parameter `yaml:"doppelgangerDetection,omitempty"`

	// The Docker Hub tag for Lighthouse
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Configuration for an external Prysm clients
type ExternalPrysmConfig struct {
	Title string `yaml:"-"`

	// The URL of the gRPC (REST) endpoint for the Beacon API
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// Custom proposal graffiti
	Graffiti Parameter `yaml:"graffiti,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection Parameter `yaml:"doppelgangerDetection,omitempty"`

	// The URL of the JSON-RPC endpoint for the Validator client
	JsonRpcUrl Parameter `yaml:"jsonRpcUrl,omitempty"`

	// The Docker Hub tag for Prysm's VC
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Configuration for an external Teku client
type ExternalTekuConfig struct {
	Title string `yaml:"-"`

	// The URL of the HTTP endpoint
	HttpUrl Parameter `yaml:"httpUrl,omitempty"`

	// Custom proposal graffiti
	Graffiti Parameter `yaml:"graffiti,omitempty"`

	// The Docker Hub tag for Teku
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection Parameter `yaml:"doppelgangerDetection,omitempty"`
}

// Generates a new ExternalExecutionConfig configuration
func NewExternalExecutionConfig(cfg *RocketPoolConfig) *ExternalExecutionConfig {
	return &ExternalExecutionConfig{
		Title: "External Execution Client Settings",

		HttpUrl: Parameter{
			ID:                 "httpUrl",
			Name:               "HTTP URL",
			Description:        "The URL of the HTTP RPC endpoint for your external Execution client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead, for example 'http://192.168.1.100:8545'.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Api, ContainerID_Eth2, ContainerID_Node, ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		WsUrl: Parameter{
			ID:                 "wsUrl",
			Name:               "Websocket URL",
			Description:        "The URL of the Websocket RPC endpoint for your external Execution client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead, for example 'http://192.168.1.100:8546'.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Api, ContainerID_Eth2, ContainerID_Node, ContainerID_Watchtower},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}
}

// Generates a new ExternalLighthouseClient configuration
func NewExternalLighthouseConfig(cfg *RocketPoolConfig) *ExternalLighthouseConfig {
	return &ExternalLighthouseConfig{
		Title: "External Lighthouse Settings",

		HttpUrl: Parameter{
			ID:                 "httpUrl",
			Name:               "HTTP URL",
			Description:        "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth1, ContainerID_Api, ContainerID_Validator, ContainerID_Watchtower, ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		Graffiti: Parameter{
			ID:                 GraffitiID,
			Name:               "Custom Graffiti",
			Description:        "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: defaultGraffiti},
			MaxLength:          16,
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		DoppelgangerDetection: Parameter{
			ID:                 DoppelgangerDetectionID,
			Name:               "Enable Doppelgänger Detection",
			Description:        "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:               ParameterType_Bool,
			Default:            map[Network]interface{}{Network_All: defaultDoppelgangerDetection},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Lighthouse container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: getLighthouseTagProd(),
				Network_Prater:  getLighthouseTagTest(),
				Network_Devnet:  getLighthouseTagTest(),
				Network_Holesky: getLighthouseTagTest(),
			},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalVcFlags: Parameter{
			ID:                 "additionalVcFlags",
			Name:               "Additional Validator Client Flags",
			Description:        "Additional custom command line flags you want to pass Lighthouse's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Generates a new ExternalLodestarClient configuration
func NewExternalLodestarConfig(cfg *RocketPoolConfig) *ExternalLodestarConfig {
	return &ExternalLodestarConfig{
		Title: "External Lodestar Settings",

		HttpUrl: Parameter{
			ID:                 "httpUrl",
			Name:               "HTTP URL",
			Description:        "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth1, ContainerID_Api, ContainerID_Validator, ContainerID_Watchtower, ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		Graffiti: Parameter{
			ID:                 GraffitiID,
			Name:               "Custom Graffiti",
			Description:        "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: defaultGraffiti},
			MaxLength:          16,
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		DoppelgangerDetection: Parameter{
			ID:                 DoppelgangerDetectionID,
			Name:               "Enable Doppelgänger Detection",
			Description:        "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:               ParameterType_Bool,
			Default:            map[Network]interface{}{Network_All: defaultDoppelgangerDetection},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Lodestar container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: lodestarTagProd,
				Network_Prater:  lodestarTagTest,
				Network_Devnet:  lodestarTagTest,
				Network_Holesky: lodestarTagTest,
			},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalVcFlags: Parameter{
			ID:                 "additionalVcFlags",
			Name:               "Additional Validator Client Flags",
			Description:        "Additional custom command line flags you want to pass Lodestar's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Generates a new ExternalNimbusConfig configuration
func NewExternalNimbusConfig(cfg *RocketPoolConfig) *ExternalNimbusConfig {

	return &ExternalNimbusConfig{
		Title: "External Nimbus Settings",

		HttpUrl: Parameter{
			ID:                 "httpUrl",
			Name:               "HTTP URL",
			Description:        "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth1, ContainerID_Api, ContainerID_Validator, ContainerID_Watchtower, ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		Graffiti: Parameter{
			ID:                 GraffitiID,
			Name:               "Custom Graffiti",
			Description:        "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: defaultGraffiti},
			MaxLength:          16,
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		DoppelgangerDetection: Parameter{
			ID:                 DoppelgangerDetectionID,
			Name:               "Enable Doppelgänger Detection",
			Description:        "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:               ParameterType_Bool,
			Default:            map[Network]interface{}{Network_All: defaultDoppelgangerDetection},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Nimbus validator container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: nimbusVcTagProd,
				Network_Prater:  nimbusVcTagTest,
				Network_Devnet:  nimbusVcTagTest,
				Network_Holesky: nimbusVcTagTest,
			},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalVcFlags: Parameter{
			ID:                 "additionalVcFlags",
			Name:               "Additional Validator Client Flags",
			Description:        "Additional custom command line flags you want to pass Nimbus's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Generates a new ExternalPrysmConfig configuration
func NewExternalPrysmConfig(cfg *RocketPoolConfig) *ExternalPrysmConfig {
	return &ExternalPrysmConfig{
		Title: "External Prysm Settings",

		HttpUrl: Parameter{
			ID:                 "httpUrl",
			Name:               "HTTP URL",
			Description:        "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth1, ContainerID_Api, ContainerID_Validator, ContainerID_Watchtower, ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		JsonRpcUrl: Parameter{
			ID:                 "jsonRpcUrl",
			Name:               "gRPC URL",
			Description:        "The URL of the gRPC API endpoint for your external client. Prysm's validator client will need this in order to connect to it.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		Graffiti: Parameter{
			ID:                 GraffitiID,
			Name:               "Custom Graffiti",
			Description:        "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: defaultGraffiti},
			MaxLength:          16,
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		DoppelgangerDetection: Parameter{
			ID:                 DoppelgangerDetectionID,
			Name:               "Enable Doppelgänger Detection",
			Description:        "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:               ParameterType_Bool,
			Default:            map[Network]interface{}{Network_All: defaultDoppelgangerDetection},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Prysm validator container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: prysmVcProd,
				Network_Prater:  prysmVcTest,
				Network_Devnet:  prysmVcTest,
				Network_Holesky: prysmVcTest,
			},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalVcFlags: Parameter{
			ID:                 "additionalVcFlags",
			Name:               "Additional Validator Client Flags",
			Description:        "Additional custom command line flags you want to pass Prysm's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Generates a new ExternalTekuClient configuration
func NewExternalTekuConfig(cfg *RocketPoolConfig) *ExternalTekuConfig {
	return &ExternalTekuConfig{
		Title: "External Teku Settings",

		HttpUrl: Parameter{
			ID:                 "httpUrl",
			Name:               "HTTP URL",
			Description:        "The URL of the HTTP Beacon API endpoint for your external client.\nNOTE: If you are running it on the same machine as the Smartnode, addresses like `localhost` and `127.0.0.1` will not work due to Docker limitations. Enter your machine's LAN IP address instead.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth1, ContainerID_Api, ContainerID_Validator, ContainerID_Watchtower, ContainerID_Node},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		Graffiti: Parameter{
			ID:                 GraffitiID,
			Name:               "Custom Graffiti",
			Description:        "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: defaultGraffiti},
			MaxLength:          16,
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		DoppelgangerDetection: Parameter{
			ID:                 DoppelgangerDetectionID,
			Name:               "Enable Doppelgänger Detection",
			Description:        "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:               ParameterType_Bool,
			Default:            map[Network]interface{}{Network_All: defaultDoppelgangerDetection},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Teku container you want to use from Docker Hub. This will be used for the Validator Client that Rocket Pool manages with your minipool keys.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: tekuTagProd,
				Network_Prater:  tekuTagTest,
				Network_Devnet:  tekuTagTest,
				Network_Holesky: tekuTagTest,
			},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalVcFlags: Parameter{
			ID:                 "additionalVcFlags",
			Name:               "Additional Validator Client Flags",
			Description:        "Additional custom command line flags you want to pass Teku's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *ExternalExecutionConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.HttpUrl,
		&cfg.WsUrl,
	}
}

// Get the parameters for this config
func (cfg *ExternalLighthouseConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.HttpUrl,
		&cfg.Graffiti,
		&cfg.DoppelgangerDetection,
		&cfg.ContainerTag,
		&cfg.AdditionalVcFlags,
	}
}

// Get the parameters for this config
func (cfg *ExternalNimbusConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.HttpUrl,
		&cfg.Graffiti,
		&cfg.DoppelgangerDetection,
		&cfg.ContainerTag,
		&cfg.AdditionalVcFlags,
	}
}

// Get the parameters for this config
func (cfg *ExternalLodestarConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.HttpUrl,
		&cfg.Graffiti,
		&cfg.DoppelgangerDetection,
		&cfg.ContainerTag,
		&cfg.AdditionalVcFlags,
	}
}

// Get the parameters for this config
func (cfg *ExternalPrysmConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.HttpUrl,
		&cfg.JsonRpcUrl,
		&cfg.Graffiti,
		&cfg.DoppelgangerDetection,
		&cfg.ContainerTag,
		&cfg.AdditionalVcFlags,
	}
}

// Get the parameters for this config
func (cfg *ExternalTekuConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.HttpUrl,
		&cfg.Graffiti,
		&cfg.DoppelgangerDetection,
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

// Get the Docker container name of the beacon client
func (cfg *ExternalLighthouseConfig) GetBeaconNodeImage() string {
	return ""
}

// Get the Docker container name of the beacon client
func (cfg *ExternalLodestarConfig) GetBeaconNodeImage() string {
	return ""
}

// Get the Docker container name of the beacon client
func (cfg *ExternalNimbusConfig) GetBeaconNodeImage() string {
	return ""
}

// Get the Docker container name of the beacon client
func (cfg *ExternalPrysmConfig) GetBeaconNodeImage() string {
	return ""
}

// Get the Docker container name of the beacon client
func (cfg *ExternalTekuConfig) GetBeaconNodeImage() string {
	return ""
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

// Get the doppelganger detection from the config
func (cfg *ExternalLighthouseConfig) GetDoppelgangerDetection() bool {
	return cfg.DoppelgangerDetection.Value.(bool)
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
