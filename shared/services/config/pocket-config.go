package config

// Constants
const defaultPocketGatewayMainnet string = "lb/613bb4ae8c124d00353c40a1"
const defaultPocketGatewayPrater string = "lb/6126b4a783e49000343a3a47"
const pocketEventLogInterval int = 25000

// Configuration for Pocket
type PocketConfig struct {
	Title string `yaml:"title,omitempty"`

	// Common parameters that Pocket doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"unsupportedCommonParams,omitempty"`

	// Compatible consensus clients
	CompatibleConsensusClients []ConsensusClient `yaml:"compatibleConsensusClients,omitempty"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"eventLogInterval,omitempty"`

	// The Pocket gateway ID
	GatewayID Parameter `yaml:"gatewayID,omitempty"`

	// The Docker Hub tag for Geth
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Pocket configuration
func NewPocketConfig(config *RocketPoolConfig, isFallback bool) *PocketConfig {

	prefix := ""
	if isFallback {
		prefix = "FALLBACK_"
	}

	title := "Pocket Settings"
	if isFallback {
		title = "Fallback Pocket Settings"
	}

	return &PocketConfig{
		Title: title,

		UnsupportedCommonParams: []string{ecWsPortID},

		CompatibleConsensusClients: []ConsensusClient{
			ConsensusClient_Lighthouse,
			ConsensusClient_Nimbus,
			ConsensusClient_Prysm,
			ConsensusClient_Teku,
		},

		EventLogInterval: pocketEventLogInterval,

		GatewayID: Parameter{
			ID:          "gatewayID",
			Name:        "Gateway ID",
			Description: "If you would like to use a custom gateway for Pocket instead of the default Rocket Pool gateway, enter it here.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: defaultPocketGatewayMainnet,
				Network_Prater:  defaultPocketGatewayPrater,
				Network_Kiln:    "",
			},
			Regex:                "(^$|^(lb\\/)?[0-9a-zA-Z]{24,}$)",
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "POCKET_GATEWAY_ID"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Rocket Pool EC Proxy container you want to use on Docker Hub.\nYou should leave this as the default unless you have a good reason to change it.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: powProxyTag},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to the EC Proxy, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (config *PocketConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.GatewayID,
		&config.ContainerTag,
		&config.AdditionalFlags,
	}
}

// The the title for the config
func (config *PocketConfig) GetConfigTitle() string {
	return config.Title
}
