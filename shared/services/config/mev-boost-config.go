package config

// Constants
const (
	mevBoostTag       string = "flashbots/mev-boost:v0.7.9"
	mevBoostUrlEnvVar string = "MEV_BOOST_URL"
)

// Configuration for MEV Boost
type MevBoostConfig struct {
	Title string `yaml:"-"`

	// Ownership mode
	Mode Parameter `yaml:"mode,omitempty"`

	// MEV Boost relays
	Relays Parameter `yaml:"relays,omitempty"`

	// The RPC port
	Port Parameter `yaml:"port,omitempty"`

	// The Docker Hub tag for MEV Boost
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags Parameter `yaml:"additionalFlags,omitempty"`

	// The URL of an external MEV Boost client
	ExternalUrl Parameter `yaml:"externalUrl"`
}

// Generates a new MEV Boost configuration
func NewMevBoostConfig(config *RocketPoolConfig) *MevBoostConfig {
	return &MevBoostConfig{
		Title: "MEV Boost Settings",

		Mode: Parameter{
			ID:                   "mode",
			Name:                 "MEV Boost Mode",
			Description:          "Choose whether to let the Smartnode manage your MEV boost instance (Locally Managed), or if you manage your own outside of the Smartnode stack (Externally Managed).",
			Type:                 ParameterType_Choice,
			Default:              map[Network]interface{}{Network_All: Mode_Local},
			AffectsContainers:    []ContainerID{ContainerID_Eth2, ContainerID_MevBoost},
			EnvironmentVariables: []string{},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
			Options: []ParameterOption{{
				Name:        "Locally Managed",
				Description: "Allow the Smartnode to manage an MEV boost client for you",
				Value:       Mode_Local,
			}, {
				Name:        "Externally Managed",
				Description: "Use an existing MEV boost client that you manage on your own",
				Value:       Mode_External,
			}},
		},

		Relays: Parameter{
			ID:          "relays",
			Name:        "Relays",
			Description: "A comma-separated list of MEV Boost relay URLs you want to connect to",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: "",
				Network_Prater:  "",
				Network_Kiln:    "https://0xb5246e299aeb782fbc7c91b41b3284245b1ed5206134b0028b81dfb974e5900616c67847c2354479934fc4bb75519ee1@builder-relay-kiln.flashbots.net?id=rocketpool",
				Network_Ropsten: "https://0xb124d80a00b80815397b4e7f1f05377ccc83aeeceb6be87963ba3649f1e6efa32ca870a88845917ec3f26a8e2aa25c77@builder-relay-ropsten.flashbots.net?id=rocketpool",
			},
			AffectsContainers:    []ContainerID{ContainerID_MevBoost},
			EnvironmentVariables: []string{"MEV_BOOST_RELAYS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   true,
		},

		Port: Parameter{
			ID:                   "port",
			Name:                 "Port",
			Description:          "The port that MEV Boost should serve its API on.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: uint16(18550)},
			AffectsContainers:    []ContainerID{ContainerID_Eth2, ContainerID_MevBoost},
			EnvironmentVariables: []string{"MEV_BOOST_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the MEV Boost container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: mevBoostTag},
			AffectsContainers:    []ContainerID{ContainerID_MevBoost},
			EnvironmentVariables: []string{"MEV_BOOST_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to MEV Boost, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_MevBoost},
			EnvironmentVariables: []string{"MEV_BOOST_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		ExternalUrl: Parameter{
			ID:                   "externalUrl",
			Name:                 "External URL",
			Description:          "The URL of the external MEV Boost client or provider",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{mevBoostUrlEnvVar},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (config *MevBoostConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.Mode,
		&config.Relays,
		&config.Port,
		&config.ContainerTag,
		&config.AdditionalFlags,
		&config.ExternalUrl,
	}
}

// The the title for the config
func (config *MevBoostConfig) GetConfigTitle() string {
	return config.Title
}
