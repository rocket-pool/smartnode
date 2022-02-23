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
}

// Generates a new Infura configuration
func NewInfuraConfig(config *RocketPoolConfig) *InfuraConfig {
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
			EnvironmentVariables: []string{"INFURA_PROJECT_ID"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (config *InfuraConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.ProjectID,
	}
}
