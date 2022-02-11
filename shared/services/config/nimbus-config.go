package config

const nimbusTag string = "statusim/nimbus-eth2:multiarch-v1.6.0"

// Configuration for Nimbus
type NimbusConfig struct {
	// Common parameters that Nimbus doesn't support and should be hidden
	UnsupportedCommonParams []string

	// The Docker Hub tag for Nimbus
	ContainerName Parameter

	// Custom command line flags for Nimbus
	AdditionalFlags Parameter
}

// Generates a new Nimbus configuration
func NewNimbusConfig(config *MasterConfig) *NimbusConfig {
	return &NimbusConfig{
		ContainerName: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Nimbus container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: nimbusTag},
			AffectsContainers:    []ContainerID{ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{"BN_CONTAINER_TAG", "VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to Nimbus, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Handle a network change on all of the parameters
func (config *NimbusConfig) changeNetwork(oldNetwork Network, newNetwork Network) {
	changeNetworkForParameter(&config.ContainerName, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.AdditionalFlags, oldNetwork, newNetwork)
}

// Get the common params that this client doesn't support
func (config *NimbusConfig) GetUnsupportedCommonParams() []string {
	return config.UnsupportedCommonParams
}
