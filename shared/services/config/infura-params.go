package config

// Constants
const infuraEventLogInterval int = 25000

// Configuration for Infura
type InfuraConfig struct {
	// Common parameters that Infura doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"unsupportedCommonParams,omitempty"`

	// Compatible consensus clients
	CompatibleConsensusClients []ConsensusClient `yaml:"compatibleConsensusClients,omitempty"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"eventLogInterval,omitempty"`

	// The Infura project ID
	ProjectID Parameter `yaml:"projectID,omitempty"`

	// The Docker Hub tag for Geth
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Infura configuration
func NewInfuraConfig(config *RocketPoolConfig, isFallback bool) *InfuraConfig {

	prefix := ""
	if isFallback {
		prefix = "FALLBACK_"
	}

	return &InfuraConfig{
		CompatibleConsensusClients: []ConsensusClient{
			ConsensusClient_Lighthouse,
			ConsensusClient_Nimbus,
			ConsensusClient_Prysm,
			ConsensusClient_Teku,
		},

		EventLogInterval: infuraEventLogInterval,

		ProjectID: Parameter{
			ID:                   "projectID",
			Name:                 "Project ID",
			Description:          "The ID of your `Ethereum` project in Infura. Note: This is your Project ID, not your Project Secret!",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "INFURA_PROJECT_ID"},
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
func (config *InfuraConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.ProjectID,
		&config.ContainerTag,
		&config.AdditionalFlags,
	}
}
