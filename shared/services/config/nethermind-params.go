package config

import (
	"runtime"

	"github.com/pbnjay/memory"
	"github.com/rocket-pool/smartnode/shared/types/config"
)

// Constants
const (
	nethermindTagProd          string = "nethermind/nethermind:1.31.0"
	nethermindTagTest          string = "nethermind/nethermind:1.31.0"
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

	// Nethermind's memory for in-memory pruning
	PruneMemSize config.Parameter `yaml:"pruneMemSize,omitempty"`

	// Nethermind's memory budget for full pruning
	FullPruneMemoryBudget config.Parameter `yaml:"fullPruneMemoryBudget,omitempty"`

	// Nethermind's remaining disk space to trigger a pruning
	FullPruningThresholdMb config.Parameter `yaml:"fullPruningThresholdMb,omitempty"`

	// The number of parallel tasks/threads that can be used by pruning
	FullPruningMaxDegreeOfParallelism config.Parameter `yaml:"fullPruningMaxDegreeOfParallelism,omitempty"`

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
			ID:                 "cache",
			Name:               "Cache (Memory Hint) Size",
			Description:        "The amount of RAM (in MB) you want to suggest for Nethermind's cache. While there is no guarantee that Nethermind will stay under this limit, lower values are preferred for machines with less RAM.\n\nThe default value for this will be calculated dynamically based on your system's available RAM, but you can adjust it manually.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: calculateNethermindCache()},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		MaxPeers: config.Parameter{
			ID:                 "maxPeers",
			Name:               "Max Peers",
			Description:        "The maximum number of peers Nethermind should connect to. This can be lowered to improve performance on low-power systems or constrained config.Networks. We recommend keeping it at 12 or higher.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: calculateNethermindPeers()},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		PruneMemSize: config.Parameter{
			ID:                 "pruneMemSize",
			Name:               "In-Memory Pruning Cache Size",
			Description:        "The amount of RAM (in MB) you want to dedicate to Nethermind for its in-memory pruning system. Higher values mean less writes to your SSD and slower overall database growth.\n\nThe default value for this will be calculated dynamically based on your system's available RAM, but you can adjust it manually.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: calculateNethermindPruneMemSize()},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		FullPruneMemoryBudget: config.Parameter{
			ID:                 "fullPruneMemoryBudget",
			Name:               "Full Prune Memory Budget Size",
			Description:        "The amount of RAM (in MB) you want to dedicate to Nethermind for its full pruning system. Higher values mean less writes to your SSD and faster pruning times.\n\nThe default value for this will be calculated dynamically based on your system's available RAM, but you can adjust it manually.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_All: calculateNethermindFullPruneMemBudget()},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		FullPruningThresholdMb: config.Parameter{
			ID:                 "fullPruningThresholdMb",
			Name:               "Prune threshold (MB)",
			Description:        "When the volume free space (in MB) hits this level, Nethermind will automatically start full pruning to reclaim disk space.",
			Type:               config.ParameterType_Uint,
			Default:            map[config.Network]interface{}{config.Network_Mainnet: uint64(375809), config.Network_Holesky: uint64(51200), config.Network_Devnet: uint64(51200)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		FullPruningMaxDegreeOfParallelism: config.Parameter{
			ID:                 "fullPruningMaxDegreeOfParallelism",
			Name:               "Full pruning parallelism",
			Description:        "This option will be used to determine the number of threads allocated to concurrently by Nethermind to prune data.",
			Type:               config.ParameterType_Int,
			Default:            map[config.Network]interface{}{config.Network_All: int64(0)},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		AdditionalModules: config.Parameter{
			ID:                 "additionalModules",
			Name:               "Additional Modules",
			Description:        "Additional modules you want to add to the primary JSON-RPC route. The defaults are Eth,Net,Personal,Web3. You can add any additional ones you need here; separate multiple modules with commas, and do not use spaces.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		AdditionalUrls: config.Parameter{
			ID:                 "additionalUrls",
			Name:               "Additional URLs",
			Description:        "Additional JSON-RPC URLs you want to run alongside the primary URL. These will be added to the \"--JsonRpc.AdditionalRpcUrls\" argument. Wrap each additional URL in quotes, and separate multiple URLs with commas (no spaces). Please consult the Nethermind documentation for more information on this flag, its intended usage, and its expected formatting.\n\nFor advanced users only.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		ContainerTag: config.Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Nethermind container you want to use on Docker Hub.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: nethermindTagProd,
				config.Network_Devnet:  nethermindTagTest,
				config.Network_Holesky: nethermindTagTest,
			},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalFlags: config.Parameter{
			ID:                 "additionalFlags",
			Name:               "Additional Flags",
			Description:        "Additional custom command line flags you want to pass to Nethermind, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth1},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
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

// Calculate the recommended size for Nethermind's full pruning based on the amount of system RAM
func calculateNethermindFullPruneMemBudget() uint64 {
	totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024

	if totalMemoryGB == 0 {
		return 0
	} else if totalMemoryGB < 9 {
		return 1024
	} else if totalMemoryGB < 17 {
		return 1024
	} else if totalMemoryGB < 25 {
		return 1024
	} else if totalMemoryGB < 33 {
		return 2048
	} else {
		return 4096
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
		&cfg.FullPruneMemoryBudget,
		&cfg.FullPruningThresholdMb,
		&cfg.FullPruningMaxDegreeOfParallelism,
		&cfg.AdditionalModules,
		&cfg.AdditionalUrls,
		&cfg.ContainerTag,
		&cfg.AdditionalFlags,
	}
}

// The title for the config
func (cfg *NethermindConfig) GetConfigTitle() string {
	return cfg.Title
}
