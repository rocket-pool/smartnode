package config

import (
	"github.com/rocket-pool/node-manager-core/utils/sys"
)

const (
	lighthouseTagPortableTest string = "rocketpool/lighthouse:b6a78e2"
	lighthouseTagPortableProd string = "sigp/lighthouse:v4.5.0"
	lighthouseTagModernTest   string = "rocketpool/lighthouse:b6a78e2-modern"
	lighthouseTagModernProd   string = "sigp/lighthouse:v4.5.0-modern"
	defaultLhMaxPeers         uint16 = 80
)

// Configuration for Lighthouse
type LighthouseConfig struct {
	Title string `yaml:"-"`

	// The max number of P2P peers to connect to
	MaxPeers Parameter `yaml:"maxPeers,omitempty"`

	// Common parameters that Lighthouse doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// The Docker Hub tag for Lighthouse
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags for the BN
	AdditionalBnFlags Parameter `yaml:"additionalBnFlags,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`

	// The port to use for gossip traffic using the QUIC protocol
	P2pQuicPort Parameter `yaml:"p2pQuicPort,omitempty"`
}

// Generates a new Lighthouse configuration
func NewLighthouseConfig(cfg *RocketPoolConfig) *LighthouseConfig {
	return &LighthouseConfig{
		Title: "Lighthouse Settings",

		MaxPeers: Parameter{
			ID:                 "maxPeers",
			Name:               "Max Peers",
			Description:        "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultLhMaxPeers},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
		P2pQuicPort: Parameter{
			ID:                 P2pQuicPortID,
			Name:               "P2P QUIC Port",
			Description:        "The port to use for P2P (blockchain) traffic using the QUIC protocol.",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultP2pQuicPort},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Lighthouse container you want to use from Docker Hub.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: getLighthouseTagProd(),
				Network_Prater:  getLighthouseTagTest(),
				Network_Devnet:  getLighthouseTagTest(),
				Network_Holesky: getLighthouseTagTest(),
			},
			AffectsContainers:  []ContainerID{ContainerID_Eth2, ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalBnFlags: Parameter{
			ID:                 "additionalBnFlags",
			Name:               "Additional Beacon Client Flags",
			Description:        "Additional custom command line flags you want to pass Lighthouse's Beacon Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		AdditionalVcFlags: Parameter{
			ID:                 "additionalVcFlags",
			Name:               "Additional Validator Client Flags",
			Description:        "Additional custom command line flags you want to pass Lighthouse's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *LighthouseConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.MaxPeers,
		&cfg.P2pQuicPort,
		&cfg.ContainerTag,
		&cfg.AdditionalBnFlags,
		&cfg.AdditionalVcFlags,
	}
}

// Get the common params that this client doesn't support
func (cfg *LighthouseConfig) GetUnsupportedCommonParams() []string {
	return cfg.UnsupportedCommonParams
}

// Get the Docker container name of the validator client
func (cfg *LighthouseConfig) GetValidatorImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the Docker container name of the beacon client
func (cfg *LighthouseConfig) GetBeaconNodeImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the name of the client
func (cfg *LighthouseConfig) GetName() string {
	return "Lighthouse"
}

// The the title for the config
func (cfg *LighthouseConfig) GetConfigTitle() string {
	return cfg.Title
}

// Get the appropriate LH default tag for production
func getLighthouseTagProd() string {
	missingFeatures := sys.GetMissingModernCpuFeatures()
	if len(missingFeatures) > 0 {
		return lighthouseTagPortableProd
	}
	return lighthouseTagModernProd
}

// Get the appropriate LH default tag for testnets
func getLighthouseTagTest() string {
	missingFeatures := sys.GetMissingModernCpuFeatures()
	if len(missingFeatures) > 0 {
		return lighthouseTagPortableTest
	}
	return lighthouseTagModernTest
}
