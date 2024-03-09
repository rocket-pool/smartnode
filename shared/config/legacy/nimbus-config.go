package config

import (
	"runtime"
)

const (
	// Testnet
	nimbusBnTagTest string = "statusim/nimbus-eth2:multiarch-v24.1.1"
	nimbusVcTagTest string = "statusim/nimbus-validator-client:multiarch-v24.1.1"

	// Mainnet
	nimbusBnTagProd string = "statusim/nimbus-eth2:multiarch-v24.1.1"
	nimbusVcTagProd string = "statusim/nimbus-validator-client:multiarch-v24.1.1"

	defaultNimbusMaxPeersArm uint16 = 100
	defaultNimbusMaxPeersAmd uint16 = 160
)

// Configuration for Nimbus
type NimbusConfig struct {
	Title string `yaml:"-"`

	// The max number of P2P peers to connect to
	MaxPeers Parameter `yaml:"maxPeers,omitempty"`

	// Common parameters that Nimbus doesn't support and should be hidden
	UnsupportedCommonParams []string `yaml:"-"`

	// The Docker Hub tag for the BN
	BnContainerTag Parameter `yaml:"bnContainerTag,omitempty"`

	// The Docker Hub tag for the VC
	VcContainerTag Parameter `yaml:"vcContainerTag,omitempty"`

	// The pruning mode to use in the BN
	PruningMode Parameter `yaml:"pruningMode,omitempty"`

	// Custom command line flags for the BN
	AdditionalBnFlags Parameter `yaml:"additionalBnFlags,omitempty"`

	// Custom command line flags for the VC
	AdditionalVcFlags Parameter `yaml:"additionalVcFlags,omitempty"`
}

// Generates a new Nimbus configuration
func NewNimbusConfig(cfg *RocketPoolConfig) *NimbusConfig {
	return &NimbusConfig{
		Title: "Nimbus Settings",

		MaxPeers: Parameter{
			ID:                 "maxPeers",
			Name:               "Max Peers",
			Description:        "The maximum number of peers your client should try to maintain. You can try lowering this if you have a low-resource system or a constrained network.",
			Type:               ParameterType_Uint16,
			Default:            map[Network]interface{}{Network_All: getNimbusDefaultPeers()},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
		},

		BnContainerTag: Parameter{
			ID:          "bnContainerTag",
			Name:        "Beacon Node Container Tag",
			Description: "The tag name of the Nimbus Beacon Node container you want to use on Docker Hub.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: nimbusBnTagProd,
				Network_Prater:  nimbusBnTagTest,
				Network_Devnet:  nimbusBnTagTest,
				Network_Holesky: nimbusBnTagTest,
			},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		VcContainerTag: Parameter{
			ID:          "containerTag",
			Name:        "Validator Client Container Tag",
			Description: "The tag name of the Nimbus Validator Client container you want to use on Docker Hub.",
			Type:        ParameterType_String,
			Default: map[Network]interface{}{
				Network_Mainnet: nimbusVcTagProd,
				Network_Prater:  nimbusVcTagTest,
				Network_Devnet:  nimbusVcTagTest,
				Network_Holesky: nimbusVcTagTest,
			},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         false,
			OverwriteOnUpgrade: true,
		},

		PruningMode: Parameter{
			ID:                 "pruningMode",
			Name:               "Pruning Mode",
			Description:        "Choose how Nimbus will prune its database. Highlight each option to learn more about it.",
			Type:               ParameterType_Choice,
			Default:            map[Network]interface{}{Network_All: NimbusPruningMode_Archive},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         false,
			OverwriteOnUpgrade: false,
			Options: []ParameterOption{{
				Name:        "Archive",
				Description: "Nimbus will download the entire Beacon Chain history and store it forever. This is healthier for the overall network, since people will be able to sync the entire chain from scratch using your node.",
				Value:       NimbusPruningMode_Archive,
			}, {
				Name:        "Pruned",
				Description: "Nimbus will only keep the last 5 months of data available, and will delete everything older than that. This will make Nimbus use less disk space overall, but you won't be able to access state older than 5 months (such as regenerating old rewards trees).\n\n[orange]WARNING: Pruning an *existing* database will take a VERY long time when Nimbus first starts. If you change from Archive to Pruned, you should delete your old chain data and do a checkpoint sync using `rocketpool service resync-eth2`. Make sure you have a checkpoint sync provider specified first!",
				Value:       NimbusPruningMode_Prune,
			}},
		},

		AdditionalBnFlags: Parameter{
			ID:                 "additionalBnFlags",
			Name:               "Additional Beacon Client Flags",
			Description:        "Additional custom command line flags you want to pass Nimbus's Beacon Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Eth2},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},

		AdditionalVcFlags: Parameter{
			ID:                 "additionalVcFlags",
			Name:               "Additional Validator Client Flags",
			Description:        "Additional custom command line flags you want to pass Nimbus's Validator Client, to take advantage of other settings that the Smartnode's configuration doesn't cover.",
			Type:               ParameterType_String,
			Default:            map[Network]interface{}{Network_All: ""},
			AffectsContainers:  []ContainerID{ContainerID_Validator},
			CanBeBlank:         true,
			OverwriteOnUpgrade: false,
		},
	}
}

// Get the parameters for this config
func (cfg *NimbusConfig) GetParameters() []*Parameter {
	return []*Parameter{
		&cfg.MaxPeers,
		&cfg.PruningMode,
		&cfg.BnContainerTag,
		&cfg.VcContainerTag,
		&cfg.AdditionalBnFlags,
		&cfg.AdditionalVcFlags,
	}
}

// Get the common params that this client doesn't support
func (cfg *NimbusConfig) GetUnsupportedCommonParams() []string {
	return cfg.UnsupportedCommonParams
}

// Get the Docker container name of the validator client
func (cfg *NimbusConfig) GetValidatorImage() string {
	return cfg.VcContainerTag.Value.(string)
}

// Get the Docker container name of the beacon client
func (cfg *NimbusConfig) GetBeaconNodeImage() string {
	return cfg.BnContainerTag.Value.(string)
}

// Get the name of the client
func (cfg *NimbusConfig) GetName() string {
	return "Nimbus"
}

// The the title for the config
func (cfg *NimbusConfig) GetConfigTitle() string {
	return cfg.Title
}

func getNimbusDefaultPeers() uint16 {
	if runtime.GOARCH == "arm64" {
		return defaultNimbusMaxPeersArm
	}

	return defaultNimbusMaxPeersAmd
}
