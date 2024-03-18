package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

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
const defaultOpenBnApiPort string = string(config.RPC_Closed)
const defaultDoppelgangerDetection bool = true

// Common parameters shared by all of the Beacon Clients
type ConsensusCommonConfig struct {
	Title string `yaml:"-"`

	// Custom proposal graffiti
	Graffiti config.Parameter `yaml:"graffiti,omitempty"`

	// The checkpoint sync URL if used
	CheckpointSyncProvider config.Parameter `yaml:"checkpointSyncProvider,omitempty"`

	// The port to use for gossip traffic
	P2pPort config.Parameter `yaml:"p2pPort,omitempty"`

	// The port to expose the HTTP API on
	ApiPort config.Parameter `yaml:"apiPort,omitempty"`

	// Toggle for forwarding the HTTP API port outside of Docker
	OpenApiPort config.Parameter `yaml:"openApiPort,omitempty"`

	// Toggle for enabling doppelganger detection
	DoppelgangerDetection config.Parameter `yaml:"doppelgangerDetection,omitempty"`
}

// Create a new ConsensusCommonParams struct
func NewConsensusCommonConfig(cfg *RocketPoolConfig) *ConsensusCommonConfig {
	portModes := config.PortModes("Allow connections from external hosts. This is safe if you're running your node on your local network. If you're a VPS user, this would expose your node to the internet and could make it vulnerable to MEV/tips theft")

	return &ConsensusCommonConfig{
		Title: "Common Consensus Client Settings",

		Graffiti: config.Parameter{
			ID:                 GraffitiID,
			Name:               "Custom Graffiti",
			Description:        "Add a short message to any blocks you propose, so the world can see what you have to say!\nIt has a 16 character limit.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: defaultGraffiti},
			MaxLength:          16,
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		CheckpointSyncProvider: config.Parameter{
			ID:   CheckpointSyncUrlID,
			Name: "Checkpoint Sync URL",
			Description: "If you would like to instantly sync using an existing Beacon node, enter its URL.\n" +
				"Example: https://checkpoint-sync.holesky.ethpandaops.io (for the Holesky Testnet).\n" +
				"Leave this blank if you want to sync normally from the start of the chain.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: defaultCheckpointSyncProvider},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		P2pPort: config.Parameter{
			ID:                 P2pPortID,
			Name:               "P2P Port",
			Description:        "The port to use for P2P (blockchain) traffic.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultP2pPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ApiPort: config.Parameter{
			ID:                 ApiPortID,
			Name:               "HTTP API Port",
			Description:        "The port your Consensus client should run its HTTP API on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultBnApiPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Api, config.ContainerID_Node, config.ContainerID_Watchtower, config.ContainerID_Eth2, config.ContainerID_Validator, config.ContainerID_Prometheus},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		OpenApiPort: config.Parameter{
			ID:                 OpenApiPortID,
			Name:               "Expose API Port",
			Description:        "Select an option to expose your Consensus client's API port to your localhost or external hosts on the network, so other machines can access it too.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: defaultOpenBnApiPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            portModes,
		},

		DoppelgangerDetection: config.Parameter{
			ID:                 DoppelgangerDetectionID,
			Name:               "Enable Doppelgänger Detection",
			Description:        "If enabled, your client will *intentionally* miss 1 or 2 attestations on startup to check if validator keys are already running elsewhere. If they are, it will disable validation duties for them to prevent you from being slashed.",
			Type:               config.ParameterType_Bool,
			Default:            map[config.Network]interface{}{config.Network_All: defaultDoppelgangerDetection},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *ConsensusCommonConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
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
