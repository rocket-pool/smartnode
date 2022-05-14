package config

import (
	"fmt"
	"runtime"

	"github.com/pbnjay/memory"
)

// Constants
const (
	nethermindTagAmd64         string = "nethermind/nethermind:1.13.0"
	nethermindTagArm64         string = "rocketpool/nethermind:1.13.0-pi"
	nethermindEventLogInterval int    = 25000
)

// Configuration for Nethermind
type NethermindConfig struct {
	Title string `yaml:"-"`

	// Common parameters that Nethermind doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// Compatible consensus clients
	CompatibleConsensusClients []ConsensusClient `yaml:"-"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"-"`

	// Nethermind's cache memory hint
	CacheSize Parameter `yaml:"cacheSize,omitempty"`

	// Max number of P2P peers to connect to
	MaxPeers Parameter `yaml:"maxPeers,omitempty"`

	// Nethermind's memory for pruning
	PruneMemSize Parameter `yaml:"pruneMemSize,omitempty"`

	// The Docker Hub tag for Nethermind
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Nethermind configuration
func NewNethermindConfig(config *RocketPoolConfig, isFallback bool) *NethermindConfig {

	prefix := ""
	if isFallback {
		prefix = "FALLBACK_"
	}

	title := "Nethermind Settings"
	if isFallback {
		title = "Fallback Nethermind Settings"
	}

	return &NethermindConfig{
		Title: title,

		UnsupportedCommonParams: []string{},

		CompatibleConsensusClients: []ConsensusClient{
			ConsensusClient_Lighthouse,
			ConsensusClient_Nimbus,
			ConsensusClient_Prysm,
			ConsensusClient_Teku,
		},

		EventLogInterval: nethermindEventLogInterval,

		CacheSize: Parameter{
			ID:                   "cache",
			Name:                 "Cache (Memory Hint) Size",
			Description:          "The amount of RAM (in MB) you want to suggest for Nethermind's cache. While there is no guarantee that Nethermind will stay under this limit, lower values are preferred for machines with less RAM.\n\nThe default value for this will be calculated dynamically based on your system's available RAM, but you can adjust it manually.",
			Type:                 ParameterType_Uint,
			Default:              map[Network]interface{}{Network_All: calculateNethermindCache()},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_CACHE_SIZE"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		MaxPeers: Parameter{
			ID:                   "maxPeers",
			Name:                 "Max Peers",
			Description:          "The maximum number of peers Nethermind should connect to. This can be lowered to improve performance on low-power systems or constrained networks. We recommend keeping it at 12 or higher.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: calculateNethermindPeers()},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		PruneMemSize: Parameter{
			ID:                   "pruneMemSize",
			Name:                 "In-Memory Pruning Cache Size",
			Description:          "The amount of RAM (in MB) you want to dedicate to Nethermind for its in-memory pruning system. Higher values mean less writes to your SSD and slower overall database growth.\n\nThe default value for this will be calculated dynamically based on your system's available RAM, but you can adjust it manually.",
			Type:                 ParameterType_Uint,
			Default:              map[Network]interface{}{Network_All: calculateNethermindPruneMemSize()},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "NETHERMIND_PRUNE_MEM_SIZE"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Nethermind container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: getNethermindTag()},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to Nethermind, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_ADDITIONAL_FLAGS"},
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
		return 256
	} else if totalMemoryGB < 13 {
		return 1024
	} else if totalMemoryGB < 17 {
		return 2048
	} else if totalMemoryGB < 25 {
		return 4096
	} else if totalMemoryGB < 33 {
		return 6144
	} else {
		return 8192
	}
}

// Calculate the recommended size for Nethermind's in-memory pruning based on the amount of system RAM
func calculateNethermindPruneMemSize() uint64 {
	totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024

	if totalMemoryGB == 0 {
		return 0
	} else if totalMemoryGB < 9 {
		return 256
	} else if totalMemoryGB < 13 {
		return 1024
	} else if totalMemoryGB < 17 {
		return 2048
	} else if totalMemoryGB < 25 {
		return 4096
	} else if totalMemoryGB < 33 {
		return 6144
	} else {
		return 8192
	}
}

// Calculate the default number of Nethermind peers
func calculateNethermindPeers() uint16 {
	if runtime.GOARCH == "arm64" {
		return 25
	}
	return 50
}

// Get the container tag for Nethermind based on the current architecture
func getNethermindTag() string {
	if runtime.GOARCH == "arm64" {
		return nethermindTagArm64
	} else if runtime.GOARCH == "amd64" {
		return nethermindTagAmd64
	} else {
		panic(fmt.Sprintf("Nethermind doesn't support architecture %s", runtime.GOARCH))
	}
}

// Get the parameters for this config
func (config *NethermindConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.CacheSize,
		&config.MaxPeers,
		&config.PruneMemSize,
		&config.ContainerTag,
		&config.AdditionalFlags,
	}
}

// The the title for the config
func (config *NethermindConfig) GetConfigTitle() string {
	return config.Title
}
