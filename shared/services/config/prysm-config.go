package config

import (
	"github.com/rocket-pool/smartnode/shared/types/config"
)

const (
	prysmBnTest             string = "rocketpool/prysm:v5.0.3"
	prysmBnProd             string = "rocketpool/prysm:v5.0.3"
	prysmVcTest             string = "rocketpool/prysm:v5.0.3"
	prysmVcProd             string = "rocketpool/prysm:v5.0.3"
	defaultPrysmRpcPort     uint16 = 5053
	defaultPrysmOpenRpcPort string = string(config.RPC_Closed)
	defaultPrysmMaxPeers    uint16 = 70
)

// Configuration for Prysm
type PrysmConfig struct {
	Title string `yaml:"title,omitempty"`

	// Common parameters that Prysm doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"unsupportedCommonParams,omitempty"`

	// The max number of P2P peers to connect to
	MaxPeers config.Parameter `yaml:"maxPeers,omitempty"`

	// The RPC port for BN / VC connections
	RpcPort config.Parameter `yaml:"rpcPort,omitempty"`

	// Toggle for forwarding the RPC API outside of Docker
	OpenRpcPort config.Parameter `yaml:"openRpcPort,omitempty"`

	// The Docker Hub tag for the Prysm BN
	BnContainerTag config.Parameter `yaml:"bnContainerTag,omitempty"`

	// The Docker Hub tag for the Prysm VC
	VcContainerTag config.Parameter `yaml:"vcContainerTag,omitempty"`

	// Custom command line flags for the BN
	AdditionalBnFlags config.Parameter `yaml:"additionalBnFlags,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags config.Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Generates a new Prysm configuration
func NewPrysmConfig(cfg *RocketPoolConfig) *PrysmConfig {
	rpcPortModes := config.PortModes("Allow connections from external hosts. This is safe if you're running your node on your local network. If you're a VPS user, this would expose your node to the internet and could make it vulnerable to MEV/tips theft")

	return &PrysmConfig{
		Title: "Prysm Settings",

		UnsupportedCommonParams: []string{},

		MaxPeers: config.Parameter{
			ID:                 "maxPeers",
			Name:               "Max Peers",
			Description:        "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultPrysmMaxPeers},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		RpcPort: config.Parameter{
			ID:                 "rpcPort",
			Name:               "RPC Port",
			Description:        "The port Prysm should run its JSON-RPC API on.",
			Type:               config.ParameterType_Uint16,
			Default:            map[config.Network]interface{}{config.Network_All: defaultPrysmRpcPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2, config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		OpenRpcPort: config.Parameter{
			ID:                 "openRpcPort",
			Name:               "Expose RPC Port",
			Description:        "Expose Prysm's JSON-RPC port to other processes on your machine, or to your local network so other machines can access it too.",
			Type:               config.ParameterType_Choice,
			Default:            map[config.Network]interface{}{config.Network_All: defaultPrysmOpenRpcPort},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options:            rpcPortModes,
		},

		BnContainerTag: config.Parameter{
			ID:          "bnContainerTag",
			Name:        "Beacon Node Container Tag",
			Description: "The tag name of the Prysm Beacon Node container you want to use on Docker Hub.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: prysmBnProd,
				config.Network_Devnet:  prysmBnTest,
				config.Network_Holesky: prysmBnTest,
			},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		VcContainerTag: config.Parameter{
			ID:          "vcContainerTag",
			Name:        "Validator Client Container Tag",
			Description: "The tag name of the Prysm Validator Client container you want to use on Docker Hub.",
			Type:        config.ParameterType_String,
			Default: map[config.Network]interface{}{
				config.Network_Mainnet: prysmVcProd,
				config.Network_Devnet:  prysmVcTest,
				config.Network_Holesky: prysmVcTest,
			},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		AdditionalBnFlags: config.Parameter{
			ID:                 "additionalBnFlags",
			Name:               "Additional Beacon Node Flags",
			Description:        "Additional custom command line flags you want to pass Prysm's Beacon Node, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Eth2},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		AdditionalVcFlags: config.Parameter{
			ID:                 "additionalVcFlags",
			Name:               "Additional Validator Client Flags",
			Description:        "Additional custom command line flags you want to pass Prysm's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               config.ParameterType_String,
			Default:            map[config.Network]interface{}{config.Network_All: ""},
			AffectsContainers:  []config.ContainerID{config.ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *PrysmConfig) GetParameters() []*config.Parameter {
	return []*config.Parameter{
		&cfg.MaxPeers,
		&cfg.RpcPort,
		&cfg.OpenRpcPort,
		&cfg.BnContainerTag,
		&cfg.VcContainerTag,
		&cfg.AdditionalBnFlags,
		&cfg.AdditionalVcFlags,
	}
}

// Get the common params that this client doesn't support
func (cfg *PrysmConfig) GetUnsupportedCommonParams() []string {
	return cfg.UnsupportedCommonParams
}

// Get the Docker container name of the validator client
func (cfg *PrysmConfig) GetValidatorImage() string {
	return cfg.VcContainerTag.Value.(string)
}

// Get the Docker container name of the beacon client
func (cfg *PrysmConfig) GetBeaconNodeImage() string {
	return cfg.BnContainerTag.Value.(string)
}

// Get the name of the client
func (cfg *PrysmConfig) GetName() string {
	return "Prysm"
}

// The the title for the config
func (cfg *PrysmConfig) GetConfigTitle() string {
	return cfg.Title
}
