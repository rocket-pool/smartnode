package config

import (
	"fmt"
	"runtime"
)

// Constants
const (
	besuTagAmd64         string = "hyperledger/besu:22.4.1-SNAPSHOT-openjdk-latest"
	besuTagArm64         string = "hyperledger/besu:22.4.1-SNAPSHOT-openjdk-latest"
	besuEventLogInterval int    = 25000
	besuMaxPeers         uint16 = 25
)

// Configuration for Besu
type BesuConfig struct {
	Title string `yaml:"-"`

	// Common parameters that Besu doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// Compatible consensus clients
	CompatibleConsensusClients []ConsensusClient `yaml:"-"`

	// The max number of events to query in a single event log query
	EventLogInterval int `yaml:"-"`

	// Max number of P2P peers to connect to
	MaxPeers Parameter `yaml:"maxPeers,omitempty"`

	// The Docker Hub tag for Besu
	ContainerTag Parameter `yaml:"containerTag,omitempty"`

	// Custom command line flags
	AdditionalFlags Parameter `yaml:"additionalFlags,omitempty"`
}

// Generates a new Besu configuration
func NewBesuConfig(config *RocketPoolConfig, isFallback bool) *BesuConfig {

	prefix := ""
	if isFallback {
		prefix = "FALLBACK_"
	}

	title := "Besu Settings"
	if isFallback {
		title = "Fallback Besu Settings"
	}

	return &BesuConfig{
		Title: title,

		UnsupportedCommonParams: []string{},

		CompatibleConsensusClients: []ConsensusClient{
			ConsensusClient_Lighthouse,
			ConsensusClient_Nimbus,
			ConsensusClient_Prysm,
			ConsensusClient_Teku,
		},

		EventLogInterval: nethermindEventLogInterval,

		MaxPeers: Parameter{
			ID:                   "maxPeers",
			Name:                 "Max Peers",
			Description:          "The maximum number of peers Besu should connect to. This can be lowered to improve performance on low-power systems or constrained networks. We recommend keeping it at 12 or higher.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: besuMaxPeers},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		ContainerTag: Parameter{
			ID:                   "containerTag",
			Name:                 "Container Tag",
			Description:          "The tag name of the Besu container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: getBesuTag()},
			AffectsContainers:    []ContainerID{ContainerID_Eth1},
			EnvironmentVariables: []string{prefix + "EC_CONTAINER_TAG"},
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
			EnvironmentVariables: []string{prefix + "EC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the container tag for Besu based on the current architecture
func getBesuTag() string {
	if runtime.GOARCH == "arm64" {
		return besuTagArm64
	} else if runtime.GOARCH == "amd64" {
		return besuTagAmd64
	} else {
		panic(fmt.Sprintf("Besu doesn't support architecture %s", runtime.GOARCH))
	}
}

// Get the parameters for this config
func (config *BesuConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.MaxPeers,
		&config.ContainerTag,
		&config.AdditionalFlags,
	}
}

// The the title for the config
func (config *BesuConfig) GetConfigTitle() string {
	return config.Title
}
