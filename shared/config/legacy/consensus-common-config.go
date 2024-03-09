package config

// Param IDs
const GraffitiID string = "graffiti"
const CheckpointSyncUrlID string = "checkpointSyncUrl"
const P2pPortID string = "p2pPort"
const P2pQuicPortID string = "p2pQuicPort"
const ApiPortID string = "apiPort"
const OpenApiPortID string = "openApiPort"
const DoppelgangerDetectionID string = "doppelgangerDetection"

// Defaults
const defaultGraffiti string = ""
const defaultCheckpointSyncProvider string = ""
const defaultP2pPort uint16 = 9001
const defaultP2pQuicPort uint16 = 8001
const defaultBnApiPort uint16 = 5052
const defaultOpenBnApiPort string = string(RPC_Closed)
const defaultDoppelgangerDetection bool = true

// Common parameters shared by all of the Beacon Clients
type ConsensusCommonConfig struct {
	Title string `yaml:"-"`

	// Custom proposal graffiti
	Graffiti Parameter `yaml:"graffiti,omitempty"`

	// The checkpoint sync URL if used
	CheckpointSyncProvider Parameter `yaml:"checkpointSyncProvider,omitempty"`

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
func NewConsensusCommonConfig(cfg *RocketPoolConfig) *ConsensusCommonConfig {
	portModes := PortModes("Allow connections from external hosts. This is safe if you're running your node on your local network. If you're a VPS user, this would expose your node to the internet and could make it vulnerable to MEV/tips theft")

	return &ConsensusCommonConfig{
		Title: "Common Consensus Client Settings",

		Graffiti: Parameter{
			ID:                 GraffitiID,
			Name:               "Custom Graffiti",
			Description:        "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: defaultGraffiti},
			MaxLength:          16,
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		CheckpointSyncProvider: Parameter{
			ID:   CheckpointSyncUrlID,
			Name: "Checkpoint Sync URL",
			Description: "If you would like to instantly sync using an existing Beacon node, enter its URL.\n" +
				"Example: https://<project ID>:<secret>@eth2-beacon-prater.infura.io\n" +
				"Leave this blank if you want to sync normally from the start of the chain.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: defaultCheckpointSyncProvider},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		P2pPort: Parameter{
			ID:                 P2pPortID,
			Name:               "P2P Port",
			Description:        "The port to use for P2P (blockchain) traffic.",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultP2pPort},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ApiPort: Parameter{
			ID:                 ApiPortID,
			Name:               "HTTP API Port",
			Description:        "The port your Consensus client should run its HTTP API on.",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: defaultBnApiPort},
			AffectsContainers:  []ContainerID{ContainerID_Api, ContainerID_Node, ContainerID_Watchtower, ContainerID_Eth2, ContainerID_Validator, ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		OpenApiPort: Parameter{
			ID:                 OpenApiPortID,
			Name:               "Expose API Port",
			Description:        "Select an option to expose your Consensus client's API port to your localhost or external hosts on the network, so other machines can access it too.",
			Type:               ParameterType_Choice,
			Default:            map[Network]interface{}{Network_All: defaultOpenBnApiPort},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            portModes,
		},

		DoppelgangerDetection: Parameter{
			ID:                 DoppelgangerDetectionID,
			Name:               "Enable Doppelgänger Detection",
			Description:        "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:               ParameterType_Bool,
			Default:            map[Network]interface{}{Network_All: defaultDoppelgangerDetection},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *ConsensusCommonConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.Graffiti,
		&cfg.CheckpointSyncProvider,
		&cfg.P2pPort,
		&cfg.ApiPort,
		&cfg.OpenApiPort,
		&cfg.DoppelgangerDetection,
	}
}

// The the title for the config
func (cfg *ConsensusCommonConfig) GetConfigTitle() string {
	return cfg.Title
}
