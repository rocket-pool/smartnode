package config

// Constants
const ipfsTag string = "ipfs/go-ipfs:master-2022-04-06-dd06dd0"

// Defaults
const defaultIpfsP2pPort uint16 = 9400

// Configuration for IPFS
type IpfsConfig struct {
	Title string `yaml:"title,omitempty"`

	// The port for P2P traffic
	P2pPort Parameter `yaml:"p2pPort,omitempty"`

	// The name of the profile to use
	Profile Parameter `yaml:"profile,omitempty"`

	// The Docker Hub tag for IPFS
	ContainerTag Parameter `yaml:"containerTag,omitempty"`
}

// Generates a new IPFS config
func NewIpfsConfig(config *RocketPoolConfig) *IpfsConfig {
	return &IpfsConfig{
		Title: "IPFS Settings",

		P2pPort: Parameter{
			ID:                   "p2pPort",
			Name:                 "P2P Port",
			Description:          "The port IPFS should use to connect to other peers on the network.\n\nNOTE: You will have to forward this port in your router's settings for it to be effective!",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultIpfsP2pPort},
			AffectsContainers:    []ContainerID{ContainerID_Ipfs},
			EnvironmentVariables: []string{"IPFS_P2P_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		Profile: Parameter{
			ID:                   "profile",
			Name:                 "Profile Name",
			Description:          "The name of the profile you want to set up IPFS's daemon with.\n\nPlease see https://docs.ipfs.io/how-to/default-profile/#available-profiles for a list of profile names to choose from.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: "lowpower"},
			AffectsContainers:    []ContainerID{ContainerID_Ipfs},
			EnvironmentVariables: []string{"IPFS_PROFILE"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the IPFS container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ipfsTag},
			AffectsContainers:    []ContainerID{ContainerID_Ipfs},
			EnvironmentVariables: []string{"IPFS_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},
	}
}

// Get the parameters for this config
func (config *IpfsConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.P2pPort,
		&config.Profile,
		&config.ContainerTag,
	}
}

// The the title for the config
func (config *IpfsConfig) GetConfigTitle() string {
	return config.Title
}
