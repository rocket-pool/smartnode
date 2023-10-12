package config

import (
	"runtime"

	"github.com/pbnjay/memory"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	nethermindTagProd          string = "nethermind/nethermind:1.21.0"
	nethermindTagTest          string = "nethermind/nethermind:1.20.4"
	nethermindEventLogInterval int    = 1000
	nethermindStopSignal       string = "SIGTERM"
)

// Configuration for Nethermind
type NethermindConfig struct {
	Title string `yaml:"-"`

	// Common parameters that Nethermind doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// Compatible consensus clients
	CompatibleConsensusClients []config.ConsensusClient `yaml:"-"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"-"`

	// Nethermind's cache memory hint
	CacheSize config.Parameter `yaml:"cacheSize,omitempty"`

	// Max number of P2P peers to connect to
	MaxPeers config.Parameter `yaml:"maxPeers,omitempty"`

	// Nethermind's memory for pruning
	PruneMemSize config.Parameter `yaml:"pruneMemSize,omitempty"`

	// Flag for downloading complete chain history instead of starting from Beacon deployment
	DownloadCompleteHistory config.Parameter `yaml:"downloadCompleteHistory,omitempty"`

	// Additional modules to enable on the primary JSON RPC endpoint
	AdditionalModules config.Parameter `yaml:"additionalModules,omitempty"`

	// Additional JSON RPC URLs
	AdditionalUrls config.Parameter `yaml:"additionalUrls,omitempty"`

	// The Docker Hub tag for Nethermind
	ContainerTag config.Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags config.Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Nethermind configuration
func NewNethermindConfig(cfg *RocketPoolConfig) *NethermindConfig {
	return &NethermindConfig{
		Title: "Nethermind Settings",

		UnsupportedCommonParams: []string{},

		CompatibleConsensusClients: []config.ConsensusClient{
			config.ConsensusClient_Lighthouse,
			config.ConsensusClient_Lodestar,
			config.ConsensusClient_Nimbus,
			config.ConsensusClient_Prysm,
			config.ConsensusClient_Teku,
		},

		EventLogInterval: nethermindEventLogInterval,

		CacheSize: config.Parameter{
			ID:                   "cache",
			Name:                 "Cache (Memory Hint) Size",
			Description:          "The amount of RAM (in MB) you want to suggest for Nethermind's cache. While there is no guarantee that Nethermind will stay under this limit, lower values are preferred for machines with less RAM.\n\nThe default value for this will be calculated dynamically based on your system's available RAM, but you can adjust it manually.",
			Type:                 config.ParameterType_Uint,
			Default:              map[config.Network]interface{}{config.Network_All: calculateNethermindCache()},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_CACHE_SIZE"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		MaxPeers: config.Parameter{
			ID:                   "maxPeers",
			Name:                 "Max Peers",
			Description:          "The maximum number of peers Nethermind should connect to. This can be lowered to improve performance on low-power systems or constrained config.Networks. We recommend keeping it at 12 or higher.",
			Type:                 config.ParameterType_Uint16,
			Default:              map[config.Network]interface{}{config.Network_All: calculateNethermindPeers()},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		PruneMemSize: config.Parameter{
			ID:                   "pruneMemSize",
			Name:                 "In-Memory Pruning Cache Size",
			Description:          "The amount of RAM (in MB) you want to dedicate to Nethermind for its in-memory pruning system. Higher values mean less writes to your SSD and slower overall database growth.\n\nThe default value for this will be calculated dynamically based on your system's available RAM, but you can adjust it manually.",
			Type:                 config.ParameterType_Uint,
			Default:              map[config.Network]interface{}{config.Network_All: calculateNethermindPruneMemSize()},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"NETHERMIND_PRUNE_MEM_SIZE"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		DownloadCompleteHistory: config.Parameter{
			ID:                   "downloadCompleteHistory",
			Name:                 "Download Complete History",
			Description:          "***For Mainnet Only - No Effect on Prater or Holesky***\n\nBy default, Nethermind will only download chain information from the block that the Beacon Chain was deployed on. This will use less disk space, but other Ethereum nodes won't be able to connect to your node and fully sync from scratch.\n\nIf you prefer to use the old behavior and have Nethermind download the complete chain history, enable this setting.\n\n[orange]NOTE: You may need to resync Nethermind for this change to take effect with `rocketpool service resync-eth1`.",
			Type:                 config.ParameterType_Bool,
			Default:              map[config.Network]interface{}{config.Network_All: false},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"NETHERMIND_COMPLETE_HISTORY"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		AdditionalModules: config.Parameter{
			ID:                   "additionalModules",
			Name:                 "Additional Modules",
			Description:          "Additional modules you want to add to the primary JSON-RPC route. The defaults are Eth,Net,Personal,Web3. You can add any additional ones you need here; separate multiple modules with commas, and do not use spaces.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"NETHERMIND_ADDITIONAL_MODULES"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		AdditionalUrls: config.Parameter{
			ID:                   "additionalUrls",
			Name:                 "Additional URLs",
			Description:          "Additional JSON-RPC URLs you want to run alongside the primary URL. These will be added to the \"--JsonRpc.AdditionalRpcUrls\" argument. Wrap each additional URL in quotes, and separate multiple URLs with commas (no spaces). Please consult the Nethermind documentation for more information on this flag, its intended usage, and its expected formatting.\n\nFor advanced users only.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"NETHERMIND_ADDITIONAL_URLS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Nethermind container you want to use on Docker Hub.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: nethermindTagProd,
				config.Network_Prater:  nethermindTagTest,
				config.Network_Devnet:  nethermindTagTest,
				config.Network_Holesky: nethermindTagTest,
			},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: config.Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to Nethermind, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 config.ParameterType_String,
			Default:              map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:    []config.ContainerID{config.ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Calculate the recommended size for Nethermind's cache based on the amount of system RAM
func calculateNethermindCache() uint64 {
	totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024

	if totalMemoryGB == 0 {
		return 0
	} else if totalMemoryGB < 9 {
		return 512
	} else if totalMemoryGB < 13 {
		return 512
	} else if totalMemoryGB < 17 {
		return 1024
	} else if totalMemoryGB < 25 {
		return 1024
	} else if totalMemoryGB < 33 {
		return 1024
	} else {
		return 2048
	}
}

// Calculate the recommended size for Nethermind's in-memory pruning based on the amount of system RAM
func calculateNethermindPruneMemSize() uint64 {
	totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024

	if totalMemoryGB == 0 {
		return 0
	} else if totalMemoryGB < 9 {
		return 512
	} else if totalMemoryGB < 13 {
		return 512
	} else if totalMemoryGB < 17 {
		return 1024
	} else if totalMemoryGB < 25 {
		return 1024
	} else if totalMemoryGB < 33 {
		return 1024
	} else {
		return 1024
	}
}

// Calculate the default number of Nethermind peers
func calculateNethermindPeers() uint16 {
	if runtime.GOARCH == "arm64" {
		return 25
	}
	return 50
}

// Get the parameters for this config
func (cfg *NethermindConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.CacheSize,
		&cfg.MaxPeers,
		&cfg.PruneMemSize,
		&cfg.DownloadCompleteHistory,
		&cfg.AdditionalModules,
		&cfg.AdditionalUrls,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// The the title for the config
func (cfg *NethermindConfig) GetConfigTitle() string {
	return cfg.Title
}
