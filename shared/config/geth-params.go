package config

import (
	"runtime"

	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	gethTagProd          string = "ethereum/client-go:v1.13.2"
	gethTagTest          string = "ethereum/client-go:v1.13.2"
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

	// The flag for enabling PBSS
	EnablePbss config.Parameter `yaml:"enablePbss,omitempty"`

	// Max number of P2P peers to connect to
	MaxPeers config.Parameter `yaml:"maxPeers,omitempty"`

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

		EnablePbss: config.Parameter{
			ID:                   "enablePbss",
			Name:                 "Enable PBSS",
			Description:          "Enable Geth's new path-based state scheme. With this enabled, you will no longer need to manually prune Geth; it will automatically prune its database in real-time.\n\n[orange]NOTE:\nEnabling this will require you to remove and resync your Geth DB using `rocketpool service resync-eth1`.\nYou will need a synced fallback node configured before doing this, or you will no longer be able to attest until it has finished resyncing!",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"GETH_ENABLE_PBSS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		MaxPeers: config.Parameter{
			ID:                   "maxPeers",
			Name:                 "Max Peers",
			Description:          "The maximum number of peers Geth should connect to. This can be lowered to improve performance on low-power systems or constrained config.Networks. We recommend keeping it at 12 or higher.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: calculateGethPeers()},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Geth container you want to use on Docker Hub.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: gethTagProd,
				config.Network_Prater:  gethTagTest,
				config.Network_Devnet:  gethTagTest,
				config.Network_Holesky: gethTagTest,
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: config.Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to Geth, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
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
		&cfg.EnablePbss,
		&cfg.MaxPeers,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// The the title for the config
func (cfg *GethConfig) GetConfigTitle() string {
	return cfg.Title
}
