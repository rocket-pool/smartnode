package config

import (
	"runtime"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	gethTagProd          string = "ethereum/client-go:v1.16.8"
	gethTagTest          string = "ethereum/client-go:v1.16.8"
	gethEventLogInterval int    = 1000
	gethStopSignal       string = "SIGTERM"
)

// Configuration for Geth
type GethConfig struct {
	Title string `yaml:"-"`

	// Common config.Parameters that Geth doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// Compatible consensus clients
	CompatibleConsensusClients []config.ConsensusClient `yaml:"-"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"-"`

	// Max number of P2P peers to connect to
	MaxPeers config.Parameter `yaml:"maxPeers,omitempty"`

	// Number of seconds EVM calls can run before timing out
	EvmTimeout config.Parameter `yaml:"evmTimeout,omitempty"`

	// The Docker Hub tag for Geth
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Geth configuration
func NewGethConfig(cfg *RocketPoolConfig) *GethConfig {
	return &GethConfig{
		Title: "Geth Settings",

		UnsupportedCommonParams: []string{},

		CompatibleConsensusClients: []config.ConsensusClient{
			config.ConsensusClient_Lighthouse,
			config.ConsensusClient_Lodestar,
			config.ConsensusClient_Nimbus,
			config.ConsensusClient_Prysm,
			config.ConsensusClient_Teku,
		},

		EventLogInterval: gethEventLogInterval,

		MaxPeers: config.Parameter{
			ID:                 "maxPeers",
			Name:               "Max Peers",
			Description:        "The maximum number of peers Geth should connect to. This can be lowered to improve performance on low-power systems or constrained config.Networks. We recommend keeping it at 12 or higher.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: calculateGethPeers()},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		EvmTimeout: config.Parameter{
			ID:                 "evmTimeout",
			Name:               "EVM Timeout",
			Description:        "The number of seconds an Execution Client API call is allowed to run before Geth times out and aborts it. Increase this if you see a lot of timeout errors in your logs.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: uint16(5)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Geth container you want to use on Docker Hub.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: gethTagProd,
				config.Network_Devnet:  gethTagTest,
				config.Network_Testnet: gethTagTest,
			},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalFlags: config.Parameter{
			ID:                 "additionalFlags",
			Name:               "Additional Flags",
			Description:        "Additional custom command line flags you want to pass to Geth, to take advantage of other settings that the Smart Node's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Calculate the default number of Geth peers
func calculateGethPeers() uint16 {
	if runtime.GOARCH == "arm64" {
		return 25
	}
	return 50
}

// Get the config.Parameters for this config
func (cfg *GethConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.MaxPeers,
		&cfg.EvmTimeout,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// The title for the config
func (cfg *GethConfig) GetConfigTitle() string {
	return cfg.Title
}
