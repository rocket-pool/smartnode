package config

import (
	"runtime"

	"github.com/pbnjay/memory"
)

// Constants
const (
	gethTag              string = "ethereum/client-go:v1.10.17"
	gethEventLogInterval int    = 25000
)

// Configuration for Geth
type GethConfig struct {
	Title string `yaml:"-"`

	// Common parameters that Geth doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// Compatible consensus clients
	CompatibleConsensusClients []ConsensusClient `yaml:"-"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"-"`

	// Size of Geth's Cache
	CacheSize Parameter `yaml:"cacheSize,omitempty"`

	// Max number of P2P peers to connect to
	MaxPeers Parameter `yaml:"maxPeers,omitempty"`

	// The Docker Hub tag for Geth
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Geth configuration
func NewGethConfig(config *RocketPoolConfig, isFallback bool) *GethConfig {

	prefix := ""
	if isFallback {
		prefix = "FALLBACK_"
	}

	title := "Geth Settings"
	if isFallback {
		title = "Fallback Geth Settings"
	}

	return &GethConfig{
		Title: title,

		UnsupportedCommonParams: []string{},

		CompatibleConsensusClients: []ConsensusClient{
			ConsensusClient_Lighthouse,
			ConsensusClient_Nimbus,
			ConsensusClient_Prysm,
			ConsensusClient_Teku,
		},

		EventLogInterval: gethEventLogInterval,

		CacheSize: Parameter{
			ID:                   "cache",
			Name:                 "Cache Size",
			Description:          "The amount of RAM (in MB) you want Geth's cache to use. Larger values mean your disk space usage will increase slower, and you will have to prune less frequently. The default is based on how much total RAM your system has but you can adjust it manually.",
			Type:                 ParameterType_Uint,
			Default:              map[Network]interface{}{Network_All: calculateGethCache()},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_CACHE_SIZE"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		MaxPeers: Parameter{
			ID:                   "maxPeers",
			Name:                 "Max Peers",
			Description:          "The maximum number of peers Geth should connect to. This can be lowered to improve performance on low-power systems or constrained networks. We recommend keeping it at 12 or higher.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: calculateGethPeers()},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Geth container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: gethTag},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to Geth, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Calculate the recommended size for Geth's cache based on the amount of system RAM
func calculateGethCache() uint64 {
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

// Calculate the default number of Geth peers
func calculateGethPeers() uint16 {
	if runtime.GOARCH == "arm64" {
		return 25
	}
	return 50
}

// Get the parameters for this config
func (config *GethConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.CacheSize,
		&config.MaxPeers,
		&config.ContainerTag,
		&config.AdditionalFlags,
	}
}

// The the title for the config
func (config *GethConfig) GetConfigTitle() string {
	return config.Title
}
