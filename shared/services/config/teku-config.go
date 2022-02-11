package config

const tekuTag string = "consensys/teku:22.1.1"

// Configuration for Teku
type TekuConfig struct {
	// Common parameters that Teku doesn't support and should be hidden
	UnsupportedCommonParams []string

	// The Docker Hub tag for Lighthouse
	ContainerTag Parameter

	// Custom command line flags for the BN
	AdditionalBnFlags Parameter

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter
}

// Generates a new Teku configuration
func NewTekuConfig(config *MasterConfig) *TekuConfig {
	return &TekuConfig{
		UnsupportedCommonParams: []string{
			DoppelgangerDetectionID,
		},

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Teku container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: tekuTag},
			AffectsContainers:    []ContainerID{ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{"BN_CONTAINER_TAG", "VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalBnFlags: Parameter{
			ID:                   "additionalBnFlags",
			Name:                 "Additional Beacon Node Flags",
			Description:          "Additional custom command line flags you want to pass Teku's Beacon Node, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
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

// Handle a network change on all of the parameters
func (config *TekuConfig) changeNetwork(oldNetwork Network, newNetwork Network) {
	changeNetworkForParameter(&config.ContainerTag, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.AdditionalBnFlags, oldNetwork, newNetwork)
	changeNetworkForParameter(&config.AdditionalVcFlags, oldNetwork, newNetwork)
}

// Get the common params that this client doesn't support
func (config *TekuConfig) GetUnsupportedCommonParams() []string {
	return config.UnsupportedCommonParams
}
