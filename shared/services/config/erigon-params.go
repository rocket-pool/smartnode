package config

import (
	"runtime"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	erigonTagProd            string = "erigontech/erigon:v3.6.0"
	erigonTagTest            string = "erigontech/erigon:v3.6.0"
	erigonEventLogInterval   int    = 1000
	erigonStopSignal         string = "SIGINT"
	defaultErigonTorrentPort uint16 = 42069
)

// Configuration for Erigon
type ErigonConfig struct {
	Title string `yaml:"-"`

	// Common config.Parameters that Erigon doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// Compatible consensus clients
	CompatibleConsensusClients []config.ConsensusClient `yaml:"-"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"-"`

	// Max number of P2P peers to connect to
	MaxPeers config.Parameter `yaml:"maxPeers,omitempty"`

	// BitTorrent port used for snapshot sync
	TorrentPort config.Parameter `yaml:"torrentPort,omitempty"`

	// The Docker Hub tag for Erigon
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Erigon configuration
func NewErigonConfig(cfg *RocketPoolConfig) *ErigonConfig {
	return &ErigonConfig{
		Title: "Erigon Settings",

		UnsupportedCommonParams: []string{},

		CompatibleConsensusClients: []config.ConsensusClient{
			config.ConsensusClient_Lighthouse,
			config.ConsensusClient_Lodestar,
			config.ConsensusClient_Nimbus,
			config.ConsensusClient_Prysm,
			config.ConsensusClient_Teku,
		},

		EventLogInterval: erigonEventLogInterval,

		MaxPeers: config.Parameter{
			ID:                 "maxPeers",
			Name:               "Max Peers",
			Description:        "The maximum number of peers Erigon should connect to. This can be lowered to improve performance on low-power systems or constrained networks. We recommend keeping it at 12 or higher.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: calculateErigonPeers()},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		TorrentPort: config.Parameter{
			ID:                 "torrentPort",
			Name:               "Torrent Port",
			Description:        "The port Erigon should use for BitTorrent snapshot sync. This must be reachable from the internet (TCP and UDP), just like the P2P port.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultErigonTorrentPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: config.Parameter{
			ID:                 "containerTag",
			Name:               "Container Tag",
			Description:        "The tag name of the Erigon container you want to use on Docker Hub.",
			Type:               config.ParameterType_String,
			Default:            clientTagDefaults(cfg.networks, erigonTagProd, erigonTagTest),
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalFlags: config.Parameter{
			ID:                 "additionalFlags",
			Name:               "Additional Flags",
			Description:        "Additional custom command line flags you want to pass to Erigon, to take advantage of other settings that the Smart Node's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Calculate the default number of Erigon peers
func calculateErigonPeers() uint16 {
	if runtime.GOARCH == "arm64" {
		return 16
	}
	return 32
}

// Get the config.Parameters for this config
func (cfg *ErigonConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.MaxPeers,
		&cfg.TorrentPort,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// The title for the config
func (cfg *ErigonConfig) GetConfigTitle() string {
	return cfg.Title
}
