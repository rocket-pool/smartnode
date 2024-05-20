package config

import (
	"runtime"

	"github.com/pbnjay/memory"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	rethTagProd          string = "ghcr.io/paradigmxyz/reth:v0.2.0-beta.6"
	rethTagTest          string = "ghcr.io/paradigmxyz/reth:v0.2.0-beta.6"
	rethEventLogInterval int    = 1000
	rethStopSignal       string = "SIGTERM"
)

// Configuration for Reth
type RethConfig struct {
	Title string `yaml:"-"`

	// Common config.Parameters that Reth doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// Compatible consensus clients
	CompatibleConsensusClients []config.ConsensusClient `yaml:"-"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"-"`

	// Size of Reth's Cache
	CacheSize config.Parameter `yaml:"cacheSize,omitempty"`

	// Max number of P2P peers that can connect to this node
	MaxInboundPeers config.Parameter `yaml:"maxInboundPeers,omitempty"`

	// Max number of P2P peers that this node can connect to
	MaxOutboundPeers config.Parameter `yaml:"maxOutboundPeers,omitempty"`

	// The Docker Hub tag for Reth
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Reth configuration
func NewRethConfig(cfg *RocketPoolConfig) *RethConfig {
	return &RethConfig{
		Title: "Reth Settings",

		UnsupportedCommonParams: []string{},

		CompatibleConsensusClients: []config.ConsensusClient{
			config.ConsensusClient_Lighthouse,
			config.ConsensusClient_Lodestar,
			config.ConsensusClient_Nimbus,
			config.ConsensusClient_Prysm,
			config.ConsensusClient_Teku,
		},

		EventLogInterval: rethEventLogInterval,

		CacheSize: config.Parameter{
			ID:                 "cache",
			Name:               "Cache Size",
			Description:        "The amount of RAM (in MB) you want Reth's cache to use. Larger values mean your disk space usage will increase slower, and you will have to prune less frequently. The default is based on how much total RAM your system has but you can adjust it manually.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: calculateRethCache()},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		MaxInboundPeers: config.Parameter{
			ID:                 "maxInboundPeers",
			Name:               "Max Inbound Peers",
			Description:        "The maximum number of inbound peers that should be allowed to connect to Reth (peers that request to connect to your node). This can be lowered to improve performance on low-power systems or constrained networks. Inbound peers requires you to have properly forwarded ports. We recommend keeping the sum of this and max outbound peers at 12 or higher.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: calculateRethPeers()},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		MaxOutboundPeers: config.Parameter{
			ID:                 "maxOutboundPeers",
			Name:               "Max Outbound Peers",
			Description:        "The maximum number of outbound peers that Reth can connect to (peers that your node requests to connect to). This can be lowered to improve performance on low-power systems or constrained networks. Outbound peers do not require proper port forwarding, but are slower to accumulate than inbound peers. We recommend keeping the sum of this and max outbound peers at 12 or higher.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: calculateRethPeers()},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Reth container you want to use.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: rethTagProd,
				config.Network_Holesky: rethTagTest,
				config.Network_Devnet:  rethTagTest,
			},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalFlags: config.Parameter{
			ID:                 "additionalFlags",
			Name:               "Additional Flags",
			Description:        "Additional custom command line flags you want to pass to Reth, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Calculate the recommended size for Reth's cache based on the amount of system RAM
func calculateRethCache() uint64 {
	totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024

	if totalMemoryGB == 0 {
		return 0
	} else if totalMemoryGB < 9 {
		return 256
	} else if totalMemoryGB < 13 {
		return 2048
	} else if totalMemoryGB < 17 {
		return 4096
	} else if totalMemoryGB < 25 {
		return 8192
	} else if totalMemoryGB < 33 {
		return 12288
	} else {
		return 16384
	}
}

// Calculate the default number of Reth peers
func calculateRethPeers() uint16 {
	if runtime.GOARCH == "arm64" {
		return 12
	}
	return 25
}

// Get the config.Parameters for this config
func (cfg *RethConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.CacheSize,
		&cfg.MaxInboundPeers,
		&cfg.MaxOutboundPeers,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// The the title for the config
func (cfg *RethConfig) GetConfigTitle() string {
	return cfg.Title
}
