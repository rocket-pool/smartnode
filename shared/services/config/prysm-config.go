package config

import (
	"fmt"
	"runtime"
)

// v2.1.3
const (
	prysmBnTagAmd64           string = "prysmaticlabs/prysm-beacon-chain:HEAD-f9b3cd-debug"
	prysmVcTagAmd64           string = "prysmaticlabs/prysm-validator:HEAD-f9b3cd-debug"
	prysmTagArm64             string = "rocketpool/prysm:v2.1.3"
	defaultPrysmRpcPort       uint16 = 5053
	defaultPrysmOpenRpcPort   bool   = false
	defaultPrysmMaxPeers      uint16 = 45
	PrysmFeeRecipientFilename string = "rp-fee-recipients.json"
)

// Configuration for Prysm
type PrysmConfig struct {
	Title string `yaml:"title,omitempty"`

	// Common parameters that Prysm doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"unsupportedCommonParams,omitempty"`

	// The max number of P2P peers to connect to
	MaxPeers Parameter `yaml:"maxPeers,omitempty"`

	// The RPC port for BN / VC connections
	RpcPort Parameter `yaml:"rpcPort,omitempty"`

	// Toggle for forwarding the RPC API outside of Docker
	OpenRpcPort Parameter `yaml:"openRpcPort,omitempty"`

	// The Docker Hub tag for the Prysm BN
	BnContainerTag Parameter `yaml:"bnContainerTag,omitempty"`

	// The Docker Hub tag for the Prysm VC
	VcContainerTag Parameter `yaml:"vcContainerTag,omitempty"`

	// Custom command line flags for the BN
	AdditionalBnFlags Parameter `yaml:"additionalBnFlags,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Generates a new Prysm configuration
func NewPrysmConfig(config *RocketPoolConfig) *PrysmConfig {
	return &PrysmConfig{
		Title: "Prysm Settings",

		UnsupportedCommonParams: []string{
			//CheckpointSyncUrlID,
		},

		MaxPeers: Parameter{
			ID:                   "maxPeers",
			Name:                 "Max Peers",
			Description:          "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultPrysmMaxPeers},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_MAX_PEERS"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		RpcPort: Parameter{
			ID:                   "rpcPort",
			Name:                 "RPC Port",
			Description:          "The port Prysm should run its JSON-RPC API on.",
			Type:                 ParameterType_Uint16,
			Default:              map[Network]interface{}{Network_All: defaultPrysmRpcPort},
			AffectsContainers:    []ContainerID{ContainerID_Eth2, ContainerID_Validator},
			EnvironmentVariables: []string{"BN_RPC_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		OpenRpcPort: Parameter{
			ID:                   "openRpcPort",
			Name:                 "Expose RPC Port",
			Description:          "Enable this to expose Prysm's JSON-RPC port to your local network, so other machines can access it too.",
			Type:                 ParameterType_Bool,
			Default:              map[Network]interface{}{Network_All: defaultPrysmOpenRpcPort},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_OPEN_RPC_PORT"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   false,
		},

		BnContainerTag: Parameter{
			ID:                   "bnContainerTag",
			Name:                 "Beacon Node Container Tag",
			Description:          "The tag name of the Prysm Beacon Node container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: getPrysmBnTag()},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		VcContainerTag: Parameter{
			ID:                   "vcContainerTag",
			Name:                 "Validator Client Container Tag",
			Description:          "The tag name of the Prysm Validator Client container you want to use on Docker Hub.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: getPrysmVcTag()},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"VC_CONTAINER_TAG"},
			CanBeBlank:           false,
			OverwriteOnUpgrade:   true,
		},

		AdditionalBnFlags: Parameter{
			ID:                   "additionalBnFlags",
			Name:                 "Additional Beacon Node Flags",
			Description:          "Additional custom command line flags you want to pass Prysm's Beacon Node, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Eth2},
			EnvironmentVariables: []string{"BN_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},

		AdditionalVcFlags: Parameter{
			ID:                   "additionalVcFlags",
			Name:                 "Additional Validator Client Flags",
			Description:          "Additional custom command line flags you want to pass Prysm's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:                 ParameterType_String,
			Default:              map[Network]interface{}{Network_All: ""},
			AffectsContainers:    []ContainerID{ContainerID_Validator},
			EnvironmentVariables: []string{"VC_ADDITIONAL_FLAGS"},
			CanBeBlank:           true,
			OverwriteOnUpgrade:   false,
		},
	}
}

// Get the container tag for the Prysm BN based on the current architecture
func getPrysmBnTag() string {
	if runtime.GOARCH == "arm64" {
		return prysmTagArm64
	} else if runtime.GOARCH == "amd64" {
		return prysmBnTagAmd64
	} else {
		panic(fmt.Sprintf("Prysm doesn't support architecture %s", runtime.GOARCH))
	}
}

// Get the container tag for the Prysm VC based on the current architecture
func getPrysmVcTag() string {
	if runtime.GOARCH == "arm64" {
		return prysmTagArm64
	} else if runtime.GOARCH == "amd64" {
		return prysmVcTagAmd64
	} else {
		panic(fmt.Sprintf("Prysm doesn't support architecture %s", runtime.GOARCH))
	}
}

// Get the parameters for this config
func (config *PrysmConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&config.MaxPeers,
		&config.RpcPort,
		&config.OpenRpcPort,
		&config.BnContainerTag,
		&config.VcContainerTag,
		&config.AdditionalBnFlags,
		&config.AdditionalVcFlags,
	}
}

// Get the common params that this client doesn't support
func (config *PrysmConfig) GetUnsupportedCommonParams() []string {
	return config.UnsupportedCommonParams
}

// Get the Docker container name of the validator client
func (config *PrysmConfig) GetValidatorImage() string {
	return config.VcContainerTag.Value.(string)
}

// Get the name of the client
func (config *PrysmConfig) GetName() string {
	return "Prysm"
}

// The the title for the config
func (config *PrysmConfig) GetConfigTitle() string {
	return config.Title
}
