package config

import "runtime"

const nimbusTag string = "statusim/nimbus-eth2:multiarch-v22.5.2"
const defaultNimbusMaxPeersArm uint16 = 100
const defaultNimbusMaxPeersAmd uint16 = 160

// Configuration for Nimbus
type NimbusConfig struct {
	Title string `yaml:"-"`

	// The max number of P2P peers to connect to
	MaxPeers Parameter `yaml:"maxPeers,omitempty"`

	// Common parameters that Nimbus doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// The Docker Hub tag for Nimbus
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for Nimbus
	AdditionalFlags Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Nimbus configuration
func NewNimbusConfig(config *RocketPoolConfig) *NimbusConfig {
	return &NimbusConfig{
		Title: "Nimbus Settings",

		MaxPeers: Parameter{
			ID:                   "maxPeers",
			Name:                 "Max Peers",
			Description:          "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: getNimbusDefaultPeers()},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: Parameter{
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

// Get the parameters for this config
func (config *NimbusConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.MaxPeers,
		&config.ContainerTag,
		&config.AdditionalFlags,
	}
}

// Get the common params that this client doesn't support
func (config *NimbusConfig) GetUnsupportedCommonParams() []string {
	return config.UnsupportedCommonParams
}

// Get the Docker container name of the validator client
func (config *NimbusConfig) GetValidatorImage() string {
	return config.ContainerTag.Value.(string)
}

// Get the name of the client
func (config *NimbusConfig) GetName() string {
	return "Nimbus"
}

// The the title for the config
func (config *NimbusConfig) GetConfigTitle() string {
	return config.Title
}

func getNimbusDefaultPeers() uint16 {
	if runtime.GOARCH == "arm64" {
		return defaultNimbusMaxPeersArm
	} else {
		return defaultNimbusMaxPeersAmd
	}
}
