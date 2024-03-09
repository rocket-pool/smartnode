package config

const (
	lodestarTagTest         string = "chainsafe/lodestar:v1.12.1"
	lodestarTagProd         string = "chainsafe/lodestar:v1.12.1"
	defaultLodestarMaxPeers uint16 = 50
)

// Configuration for Lodestar
type LodestarConfig struct {
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
}

// Generates a new Lodestar configuration
func NewLodestarConfig(cfg *RocketPoolConfig) *LodestarConfig {
	return &LodestarConfig{
		Title: "Lodestar Settings",

		MaxPeers: Parameter{
			ID:                 "maxPeers",
			Name:               "Max Peers",
			Description:        "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultLodestarMaxPeers},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Lodestar container you want to use from Docker Hub.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: lodestarTagProd,
				Network_Prater:  lodestarTagTest,
				Network_Devnet:  lodestarTagTest,
				Network_Holesky: lodestarTagTest,
			},
			AffectsContainers:  []ContainerID{ContainerID_Eth2, ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalBnFlags: Parameter{
			ID:                 "additionalBnFlags",
			Name:               "Additional Beacon Client Flags",
			Description:        "Additional custom command line flags you want to pass Lodestar's Beacon Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		AdditionalVcFlags: Parameter{
			ID:                 "additionalVcFlags",
			Name:               "Additional Validator Client Flags",
			Description:        "Additional custom command line flags you want to pass Lodestar's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *LodestarConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.MaxPeers,
		&cfg.ContainerTag,
		&cfg.AdditionalBnFlags,
		&cfg.AdditionalVcFlags,
	}
}

// Get the common params that this client doesn't support
func (cfg *LodestarConfig) GetUnsupportedCommonParams() []string {
	return cfg.UnsupportedCommonParams
}

// Get the Docker container name of the validator client
func (cfg *LodestarConfig) GetValidatorImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the Docker container name of the beacon client
func (cfg *LodestarConfig) GetBeaconNodeImage() string {
	return cfg.ContainerTag.Value.(string)
}

// Get the name of the client
func (cfg *LodestarConfig) GetName() string {
	return "Lodestar"
}

// The the title for the config
func (cfg *LodestarConfig) GetConfigTitle() string {
	return cfg.Title
}
