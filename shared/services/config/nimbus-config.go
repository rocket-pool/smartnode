package config

import (
	"runtime"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	nimbusTagTest            string = "statusim/nimbus-eth2:multiarch-v23.1.0"
	nimbusTagProd            string = "statusim/nimbus-eth2:multiarch-v23.1.0"
	defaultNimbusMaxPeersArm uint16 = 100
	defaultNimbusMaxPeersAmd uint16 = 160
)

// Configuration for Nimbus
type NimbusConfig struct {
	Title string `yaml:"-"`

	// The max number of P2P peers to connect to
	MaxPeers config.Parameter `yaml:"maxPeers,omitempty"`

	// Common parameters that Nimbus doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// The Docker Hub tag for Nimbus
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for Nimbus
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Nimbus configuration
func NewNimbusConfig(cfg *RocketPoolConfig) *NimbusConfig {
	return &NimbusConfig{
		Title: "Nimbus Settings",

		MaxPeers: config.Parameter{
			ID:                   "maxPeers",
			Name:                 "Max Peers",
			Description:          "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: getNimbusDefaultPeers()},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Nimbus container you want to use on Docker Hub.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: nimbusTagProd,
				config.Network_Prater:  nimbusTagTest,
				config.Network_Devnet:  nimbusTagTest,
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_Validator},
			EnvironmentVariables: []string{"BN_CONTAINER_TAG", "VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: config.Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to Nimbus, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (cfg *NimbusConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.MaxPeers,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// Get the common params that this client doesn't support
func (cfg *NimbusConfig) GetUnsupportedCommonParams() []string {
	return cfg.UnsupportedCommonParams
}

// Get the Docker container name of the validator client
func (cfg *NimbusConfig) GetValidatorImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the name of the client
func (cfg *NimbusConfig) GetName() string {
	return "Nimbus"
}

// The the title for the config
func (cfg *NimbusConfig) GetConfigTitle() string {
	return cfg.Title
}

func getNimbusDefaultPeers() uint16 {
	if runtime.GOARCH == "arm64" {
		return defaultNimbusMaxPeersArm
	}

	return defaultNimbusMaxPeersAmd
}
