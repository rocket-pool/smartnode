package config

// Configuration for Infura
type InfuraConfig struct {
	// Common parameters that Infura doesn't support and should be hidden
	UnsupportedCommonParams []string

	// Compatible consensus clients
	CompatibleConsensusClients []ConsensusClient

	// The Infura project ID
	ProjectID Parameter
}

// Generates a new Infura configuration
func NewInfuraConfig(config *MasterConfig) *InfuraConfig {
	return &InfuraConfig{
		CompatibleConsensusClients: []ConsensusClient{
			ConsensusClient_Lighthouse,
			ConsensusClient_Nimbus,
			ConsensusClient_Prysm,
			ConsensusClient_Teku,
		},

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

// Handle a network change on all of the parameters
func (config *InfuraConfig) changeNetwork(oldNetwork Network, newNetwork Network) {
	changeNetworkForParameter(&config.ProjectID, oldNetwork, newNetwork)
}
