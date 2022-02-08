package config

// Configuration for Infura
type InfuraConfig struct {
	// The master configuration this belongs to
	MasterConfig *Configuration

	// Common parameters that Infura doesn't support and should be hidden
	UnsupportedCommonParams []string

	// The Infura project ID
	ProjectID *Parameter
}

// Generates a new Infura configuration
func NewInfuraConfig(config *Configuration) *InfuraConfig {
	return &InfuraConfig{
		MasterConfig: config,

		ProjectID: &Parameter{
			ID:                   "projectID",
			Name:                 "Project ID",
			Description:          "The ID of your `Ethereum` project in Infura. Note: This is your Project ID, not your Project Secret!",
			Type:                 ParameterType_String,
			Default:              "",
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"INFURA_PROJECT_ID"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}
