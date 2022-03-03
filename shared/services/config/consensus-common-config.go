package config

// Param IDs
const GraffitiID string = "graffiti"
const CheckpointSyncUrlID string = "checkpointSyncUrl"
const MaxPeersID string = "maxPeers"
const P2pPortID string = "p2pPort"
const ApiPortID string = "apiPort"
const OpenApiPortID string = "openApiPort"
const DoppelgangerDetectionID string = "doppelgangerDetection"

// Defaults
const defaultGraffiti string = ""
const defaultCheckpointSyncProvider string = ""
const defaultMaxPeers uint16 = 100
const defaultP2pPort uint16 = 9001
const defaultBnApiPort uint16 = 5052
const defaultOpenBnApiPort bool = false
const defaultDoppelgangerDetection bool = true

// Common parameters shared by all of the Beacon Clients
type ConsensusCommonConfig struct {
	Title string `yaml:"title`

	// Custom proposal graffiti
	Graffiti Parameter `yaml:"graffiti,omitempty"`

	// The checkpoint sync URL if used
	CheckpointSyncProvider Parameter `yaml:"checkpointSyncProvider,omitempty"`

	// The max number of P2P peers to connect to
	MaxPeers Parameter `yaml:"maxPeers,omitempty"`

	// The port to use for gossip traffic
	P2pPort Parameter `yaml:"p2pPort,omitempty"`

	// The port to expose the HTTP API on
	ApiPort Parameter `yaml:"apiPort,omitempty"`

	// Toggle for forwarding the HTTP API port outside of Docker
	OpenApiPort Parameter `yaml:"openApiPort,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection Parameter `yaml:"doppelgangerDetection,omitempty"`
}

// Create a new ConsensusCommonParams struct
func NewConsensusCommonConfig(config *RocketPoolConfig) *ConsensusCommonConfig {
	return &ConsensusCommonConfig{
		Title: "Common Consensus Client Settings",

		Graffiti: Parameter{
			ID:                   GraffitiID,
			Name:                 "Custom Graffiti",
			Description:          "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: defaultGraffiti},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"CUSTOM_GRAFFITI"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		CheckpointSyncProvider: Parameter{
			ID:   CheckpointSyncUrlID,
			Name: "Checkpoint Sync URL",
			Description: "If you would like to instantly sync using an existing Beacon node, enter its URL.\n" +
				"Example: https://<project ID>:<secret>@eth2-beacon-prater.infura.io\n" +
				"Leave this blank if you want to sync normally from the start of the chain.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: defaultCheckpointSyncProvider},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"CHECKPOINT_SYNC_URL"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		MaxPeers: Parameter{
			ID:                   MaxPeersID,
			Name:                 "Max Peers",
			Description:          "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultMaxPeers},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		P2pPort: Parameter{
			ID:                   P2pPortID,
			Name:                 "P2P Port",
			Description:          "The port to use for P2P (blockchain) traffic.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultP2pPort},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_P2P_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ApiPort: Parameter{
			ID:                   ApiPortID,
			Name:                 "HTTP API Port",
			Description:          "The port your Consensus client should run its HTTP API on.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultBnApiPort},
			AffectsContainers:    []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator, ContainerID_Prometheus},
			EnvironmentVariables: []string{"BN_API_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		OpenApiPort: Parameter{
			ID:                   OpenApiPortID,
			Name:                 "Open API Port",
			Description:          "Enable this to open your Consensus client's API port to your local network, so other machines can access it too.",
			Type:                 ParameterType_Bool,
			Default:              map[Network]interface{}{Network_All: defaultOpenBnApiPort},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_OPEN_API_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		DoppelgangerDetection: Parameter{
			ID:                   DoppelgangerDetectionID,
			Name:                 "Enable Doppelgänger Detection",
			Description:          "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:                 ParameterType_Bool,
			Default:              map[Network]interface{}{Network_All: defaultDoppelgangerDetection},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"DOPPELGANGER_DETECTION"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the parameters for this config
func (config *ConsensusCommonConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.Graffiti,
		&config.CheckpointSyncProvider,
		&config.MaxPeers,
		&config.P2pPort,
		&config.ApiPort,
		&config.OpenApiPort,
		&config.DoppelgangerDetection,
	}
}

// The the title for the config
func (config *ConsensusCommonConfig) GetConfigTitle() string {
	return config.Title
}
