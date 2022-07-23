package config

import (
	"github.com/pbnjay/memory"
)

// Constants
const (
	besuTagTest          string = "hyperledger/besu:22.7.0-SNAPSHOT-openjdk-latest"
	besuTagProd          string = "hyperledger/besu:22.4.4-openjdk-latest"
	besuEventLogInterval int    = 25000
	besuMaxPeers         uint16 = 25
	besuStopSignal       string = "SIGTERM"
)

// Configuration for Besu
type BesuConfig struct {
	Title string `yaml:"-"`

	// Common parameters that Besu doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// Compatible consensus clients
	CompatibleConsensusClients []ConsensusClient `yaml:"-"`

	// Max number of P2P peers to connect to
	JvmHeapSize Parameter `yaml:"jvmHeapSize,omitempty"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"-"`

	// Max number of P2P peers to connect to
	MaxPeers Parameter `yaml:"maxPeers,omitempty"`

	// Historical state block regeneration limit
	MaxBackLayers Parameter `yaml:"maxBackLayers,omitempty"`

	// The Docker Hub tag for Besu
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Besu configuration
func NewBesuConfig(config *RocketPoolConfig) *BesuConfig {
	return &BesuConfig{
		Title: "Besu Settings",

		UnsupportedCommonParams: []string{},

		CompatibleConsensusClients: []ConsensusClient{
			ConsensusClient_Lighthouse,
			ConsensusClient_Nimbus,
			ConsensusClient_Prysm,
			ConsensusClient_Teku,
		},

		EventLogInterval: besuEventLogInterval,

		JvmHeapSize: Parameter{
			ID:                   "jvmHeapSize",
			Name:                 "JVM Heap Size",
			Description:          "The max amount of RAM, in MB, that Besu's JVM should limit itself to. Setting this lower will cause Besu to use less RAM, though it will always use more than this limit.\n\nUse 0 for automatic allocation.",
			Type:                 ParameterType_Uint,
			Default:              map[Network]interface{}{Network_All: getBesuHeapSize()},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"BESU_JVM_HEAP_SIZE"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		MaxPeers: Parameter{
			ID:                   "maxPeers",
			Name:                 "Max Peers",
			Description:          "The maximum number of peers Besu should connect to. This can be lowered to improve performance on low-power systems or constrained networks. We recommend keeping it at 12 or higher.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: besuMaxPeers},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		MaxBackLayers: Parameter{
			ID:                   "maxBackLayers",
			Name:                 "Historical Block Replay Limit",
			Description:          "Besu has the ability to revisit the state of any historical block on the chain by \"replaying\" all of the previous blocks to get back to the target. This limit controls how many blocks you can replay - in other words, how far back Besu can go in time. Normal Execution client processing will be paused while a replay is in progress.\n\n[orange]NOTE: If you try to replay a state from a long time ago, it may take Besu several minutes to rebuild the state!",
			Type:                 ParameterType_Uint,
			Default:              map[Network]interface{}{Network_All: uint64(512)},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"BESU_MAX_BACK_LAYERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: Parameter{
			ID:          "containerTag",
			Name:        "Container Tag",
			Description: "The tag name of the Besu container you want to use on Docker Hub.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: besuTagProd,
				Network_Prater:  besuTagTest,
				Network_Kiln:    besuTagTest,
				Network_Ropsten: besuTagTest,
			},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalFlags: Parameter{
			ID:                   "additionalFlags",
			Name:                 "Additional Flags",
			Description:          "Additional custom command line flags you want to pass to Besu, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{"EC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the recommended heap size for Besu
func getBesuHeapSize() uint64 {
	totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
	if totalMemoryGB < 9 {
		return 3052
	}
	return 0
}

// Get the parameters for this config
func (config *BesuConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.JvmHeapSize,
		&config.MaxPeers,
		&config.MaxBackLayers,
		&config.ContainerTag,
		&config.AdditionalFlags,
	}
}

// The the title for the config
func (config *BesuConfig) GetConfigTitle() string {
	return config.Title
}
