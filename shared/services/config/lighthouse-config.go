package config

const lighthouseTag string = "sigp/lighthouse:v2.1.2"

// Configuration for Lighthouse
type LighthouseConfig struct {
	// The master configuration this belongs to
	MasterConfig *MasterConfig

	// Common parameters that Lighthouse doesn't support and should be hidden
	UnsupportedCommonParams []string

	// The Docker Hub tag for Lighthouse
	ContainerTag *Parameter

	// Custom command line flags for the BN
	AdditionalBnFlags *Parameter

	// Custom command line flags for the VC
	AdditionalVcFlags *Parameter
}

// Generates a new Lighthouse configuration
func NewLighthouseConfig(config *MasterConfig) *LighthouseConfig {
	return &LighthouseConfig{
		MasterConfig: config,

		ContainerTag: &Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Lighthouse container you want to use from Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: lighthouseTag},
			AffectsContainers:    []ContainerID{ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{"BN_CONTAINER_TAG", "VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalBnFlags: &Parameter{
			ID:                   "additionalBnFlags",
			Name:                 "Additional Beacon Client Flags",
			Description:          "Additional custom command line flags you want to pass Lighthouse's Beacon Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		AdditionalVcFlags: &Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Lighthouse's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}
